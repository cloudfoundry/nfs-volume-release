package volumedriver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/ioutilshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/goshims/timeshim"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/volumedriver/internal/syncmap"
	"code.cloudfoundry.org/volumedriver/mountchecker"
)

type NfsVolumeInfo struct {
	Opts                    map[string]interface{} `json:"-"` // don't store opts
	dockerdriver.VolumeInfo                        // see dockerdriver.resources.go
}

type OsHelper interface {
	Umask(mask int) (oldmask int)
}

type VolumeDriver struct {
	volumes       *syncmap.SyncMap[NfsVolumeInfo]
	os            osshim.Os
	filepath      filepathshim.Filepath
	ioutil        ioutilshim.Ioutil
	time          timeshim.Time
	mountChecker  mountchecker.MountChecker
	mountPathRoot string
	mounter       Mounter
	osHelper      OsHelper
}

func NewVolumeDriver(logger lager.Logger, os osshim.Os, filepath filepathshim.Filepath, ioutil ioutilshim.Ioutil, time timeshim.Time, mountChecker mountchecker.MountChecker, mountPathRoot string, mounter Mounter, oshelper OsHelper) *VolumeDriver {
	d := &VolumeDriver{
		volumes:       syncmap.New[NfsVolumeInfo](),
		os:            os,
		filepath:      filepath,
		ioutil:        ioutil,
		time:          time,
		mountChecker:  mountChecker,
		mountPathRoot: mountPathRoot,
		mounter:       mounter,
		osHelper:      oshelper,
	}

	ctx := context.TODO()
	env := driverhttp.NewHttpDriverEnv(logger, ctx)

	d.restoreState(env)

	return d
}

func (d *VolumeDriver) Activate(env dockerdriver.Env) dockerdriver.ActivateResponse {
	return dockerdriver.ActivateResponse{
		Implements: []string{"VolumeDriver"},
	}
}

func (d *VolumeDriver) Create(env dockerdriver.Env, createRequest dockerdriver.CreateRequest) dockerdriver.ErrorResponse {
	logger := env.Logger().Session("create")
	logger.Info("start")
	defer logger.Info("end")

	if createRequest.Name == "" {
		return dockerdriver.ErrorResponse{Err: "Missing mandatory 'volume_name'"}
	}

	var ok bool
	if _, ok = createRequest.Opts["source"].(string); !ok {
		logger.Info("mount-config-missing-source", lager.Data{"volume_name": createRequest.Name})
		return dockerdriver.ErrorResponse{Err: `Missing mandatory 'source' field in 'Opts'`}
	}

	existing, err := d.getVolume(driverhttp.EnvWithLogger(logger, env), createRequest.Name)

	if err != nil {
		logger.Info("creating-volume", lager.Data{"volume_name": createRequest.Name})
		logger.Info("with-opts", lager.Data{"opts": createRequest.Opts})

		volInfo := NfsVolumeInfo{
			VolumeInfo: dockerdriver.VolumeInfo{Name: createRequest.Name},
			Opts:       createRequest.Opts,
		}

		d.volumes.Put(createRequest.Name, volInfo)
	} else {
		existing.Opts = createRequest.Opts

		d.volumes.Put(createRequest.Name, existing)
	}

	err = d.persistState(driverhttp.EnvWithLogger(logger, env))
	if err != nil {
		logger.Error("persist-state-failed", err)
		return dockerdriver.ErrorResponse{Err: fmt.Sprintf("persist state failed when creating: %s", err.Error())}
	}

	return dockerdriver.ErrorResponse{}
}

func (d *VolumeDriver) List(_ dockerdriver.Env) dockerdriver.ListResponse {
	listResponse := dockerdriver.ListResponse{
		Volumes: []dockerdriver.VolumeInfo{},
	}

	for _, val := range d.volumes.Values() {
		listResponse.Volumes = append(listResponse.Volumes, val.VolumeInfo)
	}
	listResponse.Err = ""
	return listResponse
}

func (d *VolumeDriver) Mount(env dockerdriver.Env, mountRequest dockerdriver.MountRequest) dockerdriver.MountResponse {
	logger := env.Logger().Session("mount", lager.Data{"volume": mountRequest.Name})
	logger.Info("start")
	defer logger.Info("end")

	if mountRequest.Name == "" {
		return dockerdriver.MountResponse{Err: "Missing mandatory 'volume_name'"}
	}

	volume, ok := d.volumes.Get(mountRequest.Name)
	if !ok {
		return dockerdriver.MountResponse{Err: fmt.Sprintf("Volume '%s' must be created before being mounted", mountRequest.Name)}
	}

	mountPath := d.mountPath(driverhttp.EnvWithLogger(logger, env), volume.Name)
	volume.Mountpoint = mountPath
	logger.Info("mounting-volume", lager.Data{"id": volume.Name, "mountpoint": mountPath})
	logger.Info("mount-source", lager.Data{"source": volume.Opts["source"].(string)})

	doMount := volume.MountCount < 1
	volume.MountCount++
	logger.Info("volume-ref-count-incremented", lager.Data{"name": volume.Name, "count": volume.MountCount})

	d.volumes.Put(mountRequest.Name, volume)
	if err := d.persistState(driverhttp.EnvWithLogger(logger, env)); err != nil {
		logger.Error("persist-state-failed", err)
		return dockerdriver.MountResponse{Err: fmt.Sprintf("persist state failed when mounting: %s", err.Error())}
	}

	if doMount {
		mountStartTime := d.time.Now()

		err := d.mount(driverhttp.EnvWithLogger(logger, env), copyOpts(volume.Opts), mountPath)

		mountEndTime := d.time.Now()
		mountDuration := mountEndTime.Sub(mountStartTime)
		if mountDuration > 8*time.Second {
			logger.Error("mount-duration-too-high", nil, lager.Data{"mount-duration-in-second": mountDuration / time.Second, "warning": "This may result in container creation failure!"})
		}

		switch err.(type) {
		case nil:
			return dockerdriver.MountResponse{Mountpoint: volume.Mountpoint}
		case dockerdriver.SafeError:
			errBytes, mErr := json.Marshal(err)
			if mErr != nil {
				logger.Error("failed-to-marshal-safeerror", mErr)
				return dockerdriver.MountResponse{Err: err.Error()}
			}
			return dockerdriver.MountResponse{Err: string(errBytes)}
		default:
			return dockerdriver.MountResponse{Err: err.Error()}
		}
	} else {
		// Check the volume to make sure it's still mounted before handing it out again.
		if !d.mounter.Check(driverhttp.EnvWithLogger(logger, env), volume.Name, volume.Mountpoint) {
			if err := d.mount(driverhttp.EnvWithLogger(logger, env), volume.Opts, mountPath); err != nil {
				logger.Error("remount-volume-failed", err)
				return dockerdriver.MountResponse{Err: fmt.Sprintf("Error remounting volume: %s", err.Error())}
			}
		}
		return dockerdriver.MountResponse{Mountpoint: volume.Mountpoint}
	}
}

func (d *VolumeDriver) Path(env dockerdriver.Env, pathRequest dockerdriver.PathRequest) dockerdriver.PathResponse {
	logger := env.Logger().Session("path", lager.Data{"volume": pathRequest.Name})

	if pathRequest.Name == "" {
		return dockerdriver.PathResponse{Err: "Missing mandatory 'volume_name'"}
	}

	vol, err := d.getVolume(driverhttp.EnvWithLogger(logger, env), pathRequest.Name)
	if err != nil {
		logger.Error("failed-no-such-volume-found", err, lager.Data{"mountpoint": vol.Mountpoint})

		return dockerdriver.PathResponse{Err: fmt.Sprintf("Volume '%s' not found", pathRequest.Name)}
	}

	if vol.Mountpoint == "" {
		errText := "Volume not previously mounted"
		logger.Error("failed-mountpoint-not-assigned", errors.New(errText))
		return dockerdriver.PathResponse{Err: errText}
	}

	return dockerdriver.PathResponse{Mountpoint: vol.Mountpoint}
}

func (d *VolumeDriver) Unmount(env dockerdriver.Env, unmountRequest dockerdriver.UnmountRequest) dockerdriver.ErrorResponse {
	logger := env.Logger().Session("unmount", lager.Data{"volume": unmountRequest.Name})
	logger.Info("start")
	defer logger.Info("end")

	if unmountRequest.Name == "" {
		return dockerdriver.ErrorResponse{Err: "Missing mandatory 'volume_name'"}
	}

	volume, ok := d.volumes.Get(unmountRequest.Name)
	if !ok {
		logger.Error("failed-no-such-volume-found", fmt.Errorf("could not find volume %s", unmountRequest.Name))

		return dockerdriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' not found", unmountRequest.Name)}
	}

	if volume.Mountpoint == "" {
		errText := "Volume not previously mounted"
		logger.Error("failed-mountpoint-not-assigned", errors.New(errText))
		return dockerdriver.ErrorResponse{Err: errText}
	}

	if volume.MountCount == 1 {
		if err := d.unmount(driverhttp.EnvWithLogger(logger, env), unmountRequest.Name, volume.Mountpoint); err != nil {
			return dockerdriver.ErrorResponse{Err: err.Error()}
		}
	}

	volume.MountCount--
	logger.Info("volume-ref-count-decremented", lager.Data{"name": volume.Name, "count": volume.MountCount})

	switch volume.MountCount {
	case 0:
		d.volumes.Delete(unmountRequest.Name)
	default:
		d.volumes.Put(unmountRequest.Name, volume)
	}

	if err := d.persistState(driverhttp.EnvWithLogger(logger, env)); err != nil {
		return dockerdriver.ErrorResponse{Err: fmt.Sprintf("failed to persist state when unmounting: %s", err.Error())}
	}

	return dockerdriver.ErrorResponse{}
}

func (d *VolumeDriver) Remove(env dockerdriver.Env, removeRequest dockerdriver.RemoveRequest) dockerdriver.ErrorResponse {
	logger := env.Logger().Session("remove", lager.Data{"volume": removeRequest})
	logger.Info("start")
	defer logger.Info("end")

	if removeRequest.Name == "" {
		return dockerdriver.ErrorResponse{Err: "Missing mandatory 'volume_name'"}
	}

	vol, err := d.getVolume(driverhttp.EnvWithLogger(logger, env), removeRequest.Name)

	if err != nil {
		logger.Error("warning-volume-removal", fmt.Errorf(fmt.Sprintf("Volume %s not found", removeRequest.Name)))
		return dockerdriver.ErrorResponse{}
	}

	if vol.Mountpoint != "" {
		if err := d.unmount(driverhttp.EnvWithLogger(logger, env), removeRequest.Name, vol.Mountpoint); err != nil {
			return dockerdriver.ErrorResponse{Err: err.Error()}
		}
	}

	logger.Info("removing-volume", lager.Data{"name": removeRequest.Name})

	d.volumes.Delete(removeRequest.Name)

	if err := d.persistState(driverhttp.EnvWithLogger(logger, env)); err != nil {
		return dockerdriver.ErrorResponse{Err: fmt.Sprintf("failed to persist state when removing: %s", err.Error())}
	}

	return dockerdriver.ErrorResponse{}
}

func (d *VolumeDriver) Get(env dockerdriver.Env, getRequest dockerdriver.GetRequest) dockerdriver.GetResponse {
	volume, err := d.getVolume(env, getRequest.Name)
	if err != nil {
		return dockerdriver.GetResponse{Err: err.Error()}
	}

	return dockerdriver.GetResponse{
		Volume: dockerdriver.VolumeInfo{
			Name:       getRequest.Name,
			Mountpoint: volume.Mountpoint,
		},
	}
}

func (d *VolumeDriver) getVolume(env dockerdriver.Env, volumeName string) (NfsVolumeInfo, error) {
	logger := env.Logger().Session("get-volume")

	if vol, ok := d.volumes.Get(volumeName); ok {
		logger.Info("getting-volume", lager.Data{"name": volumeName})
		return vol, nil
	}

	return NfsVolumeInfo{}, errors.New("Volume not found")
}

func (d *VolumeDriver) Capabilities(env dockerdriver.Env) dockerdriver.CapabilitiesResponse {
	return dockerdriver.CapabilitiesResponse{
		Capabilities: dockerdriver.CapabilityInfo{Scope: "local"},
	}
}

func (d *VolumeDriver) exists(path string) (bool, error) {
	_, err := d.os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (d *VolumeDriver) mountPath(env dockerdriver.Env, volumeId string) string {
	logger := env.Logger().Session("mount-path")
	orig := d.osHelper.Umask(000)
	defer d.osHelper.Umask(orig)

	dir, err := d.filepath.Abs(d.mountPathRoot)
	if err != nil {
		logger.Fatal("abs-failed", err)
	}

	if err := d.os.MkdirAll(dir, os.ModePerm); err != nil {
		logger.Fatal("mkdir-rootpath-failed", err)
	}

	return filepath.Join(dir, volumeId)
}

func (d *VolumeDriver) mount(env dockerdriver.Env, opts map[string]interface{}, mountPath string) error {
	source, sourceOk := opts["source"].(string)
	logger := env.Logger().Session("mount", lager.Data{"source": source, "target": mountPath})
	logger.Info("start")
	defer logger.Info("end")

	if !sourceOk {
		err := errors.New("no source information")
		logger.Error("unable-to-extract-source", err)
		return err
	}

	orig := d.osHelper.Umask(000)
	defer d.osHelper.Umask(orig)

	err := d.os.MkdirAll(mountPath, os.ModePerm)
	if err != nil {
		logger.Error("create-mountdir-failed", err)
		return err
	}

	err = d.mounter.Mount(env, source, mountPath, opts)
	if err != nil {
		logger.Error("mount-failed: ", err)
		rm_err := d.os.Remove(mountPath)
		if rm_err != nil {
			logger.Error("mountpoint-remove-failed", rm_err, lager.Data{"mount-path": mountPath})
		}
	}
	return err
}

func (d *VolumeDriver) persistState(env dockerdriver.Env) error {
	logger := env.Logger().Session("persist-state")
	logger.Info("start")
	defer logger.Info("end")

	orig := d.osHelper.Umask(000)
	defer d.osHelper.Umask(orig)

	stateFile := d.mountPath(env, "driver-state.json")

	stateData, err := json.Marshal(d.volumes)
	if err != nil {
		logger.Error("failed-to-marshall-state", err)
		return err
	}

	err = d.ioutil.WriteFile(stateFile, stateData, os.ModePerm)
	if err != nil {
		logger.Error("failed-to-write-state-file", err, lager.Data{"stateFile": stateFile})
		return err
	}

	logger.Debug("state-saved", lager.Data{"state-file": stateFile})
	return nil
}

func (d *VolumeDriver) restoreState(env dockerdriver.Env) {
	logger := env.Logger().Session("restore-state")
	logger.Info("start")
	defer logger.Info("end")

	stateFile := filepath.Join(d.mountPathRoot, "driver-state.json")

	stateData, err := d.ioutil.ReadFile(stateFile)
	if err != nil {
		logger.Info("failed-to-read-state-file", lager.Data{"err": err, "stateFile": stateFile})
		return
	}
	logger.Info("state", lager.Data{"state": string(stateData)})

	if err := json.Unmarshal(stateData, d.volumes); err != nil {
		logger.Error("failed-to-unmarshall-state", err, lager.Data{"stateFile": stateFile})
		return
	}
	logger.Info("state-restored", lager.Data{"state-file": stateFile})
}

func (d *VolumeDriver) unmount(env dockerdriver.Env, name string, mountPath string) error {
	logger := env.Logger().Session("unmount")
	logger.Info("start")
	defer logger.Info("end")

	exists, err := d.mountChecker.Exists(mountPath)
	if err != nil {
		logger.Error("failed-proc-mounts-check", err, lager.Data{"mountpoint": mountPath})
		return err
	}

	if !exists {
		err := d.os.Remove(mountPath)
		if err != nil {
			errText := fmt.Sprintf("Volume %s does not exist (path: %s) and unable to remove mount directory", name, mountPath)
			logger.Info("mountpoint-not-found", lager.Data{"msg": errText})
			return errors.New(errText)
		}

		errText := fmt.Sprintf("Volume %s does not exist (path: %s)", name, mountPath)
		logger.Info("mountpoint-not-found", lager.Data{"msg": errText})
		return errors.New(errText)
	}

	logger.Info("unmount-volume-folder", lager.Data{"mountpath": mountPath})

	err = d.mounter.Unmount(env, mountPath)
	if err != nil {
		logger.Error("unmount-failed", err)
		return fmt.Errorf("Error unmounting volume: %s", err.Error())
	}
	err = d.os.Remove(mountPath)
	if err != nil {
		logger.Error("remove-mountpoint-failed", err)
		return fmt.Errorf("Error removing mountpoint: %s", err.Error())
	}

	logger.Info("unmounted-volume")

	return nil
}

func (d *VolumeDriver) checkMounts(env dockerdriver.Env) {
	logger := env.Logger().Session("check-mounts")
	logger.Info("start")
	defer logger.Info("end")

	for _, key := range d.volumes.Keys() {
		mount, ok := d.volumes.Get(key)
		if ok && !d.mounter.Check(driverhttp.EnvWithLogger(logger, env), key, mount.VolumeInfo.Mountpoint) {
			d.volumes.Delete(key)
		}
	}
}

func (d *VolumeDriver) Drain(env dockerdriver.Env) error {
	logger := env.Logger().Session("check-mounts")
	logger.Info("start")
	defer logger.Info("end")

	// flush any volumes that are still in our map
	for _, key := range d.volumes.Keys() {
		if mount, ok := d.volumes.Get(key); ok {
			if mount.Mountpoint != "" && mount.MountCount > 0 {
				err := d.unmount(env, mount.Name, mount.Mountpoint)
				if err != nil {
					logger.Error("drain-unmount-failed", err, lager.Data{"mount-name": mount.Name, "mount-point": mount.Mountpoint})
				}
			}
			d.volumes.Delete(key)
		}
	}

	d.mounter.Purge(env, d.mountPathRoot)

	return nil
}

func copyOpts(input map[string]any) map[string]any {
	output := make(map[string]any)
	for k, v := range input {
		output[k] = v
	}
	return output
}
