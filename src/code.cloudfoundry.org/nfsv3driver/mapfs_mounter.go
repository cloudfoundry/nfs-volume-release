package nfsv3driver

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/goshims/syscallshim"
	"code.cloudfoundry.org/lager/v3"
	vmo "code.cloudfoundry.org/volume-mount-options"
	"code.cloudfoundry.org/volumedriver"
	"code.cloudfoundry.org/volumedriver/invoker"
	"code.cloudfoundry.org/volumedriver/mountchecker"
)

const MapfsDirectorySuffix = "_mapfs"
const MapfsMountTimeout = time.Minute * 5
const NobodyId = uint32(65534)
const UnknownId = uint32(4294967294)
const InvalidUidValueErrorMessage = "Invalid 'uid' option (0, negative, or non-integer)"
const InvalidGidValueErrorMessage = "Invalid 'gid' option (0, negative, or non-integer)"

type mapfsMounter struct {
	invoker      invoker.Invoker
	osshim       osshim.Os
	syscallshim  syscallshim.Syscall
	mountChecker mountchecker.MountChecker
	fstype       string
	defaultOpts  string
	resolver     IdResolver
	mask         vmo.MountOptsMask
	mapfsPath    string
}

var legacyNfsSharePattern *regexp.Regexp

var PurgeTimeToSleep = time.Millisecond * 100

func init() {
	legacyNfsSharePattern, _ = regexp.Compile("^nfs://([^/]+)(/.*)?$")
}

func NewMapfsMounter(
	invoker invoker.Invoker,
	osshim osshim.Os,
	syscallshim syscallshim.Syscall,
	mountChecker mountchecker.MountChecker,
	fstype string,
	defaultOpts string,
	resolver IdResolver,
	mask vmo.MountOptsMask,
	mapfsPath string,
) volumedriver.Mounter {
	return &mapfsMounter{invoker, osshim, syscallshim, mountChecker, fstype, defaultOpts, resolver, mask, mapfsPath}
}

func (m *mapfsMounter) Mount(env dockerdriver.Env, remote string, target string, opts map[string]interface{}) error {
	logger := env.Logger().Session("mount")
	logger.Info("mount-start", lager.Data{"remote": remote, "target": target, "opts": opts})
	defer logger.Info("mount-end")

	target = strings.TrimSuffix(target, "/")

	mountOptions := m.defaultOpts

	err := m.invoker.Invoke(env, "mount", []string{"-t", m.fstype, "-o", mountOptions, remote, target}).Wait()
	if err != nil {
		logger.Error("invoke-mount-failed", err, lager.Data{"mount-options": mountOptions})
		return dockerdriver.SafeError{SafeDescription: err.Error()}
	}

	err = os.Chown(target, 2000, 2000)
	if err != nil {
		logger.Error("unable-to-chown-new-mount", err)
		return dockerdriver.SafeError{SafeDescription: err.Error()}
	}

	return nil
}

func (m *mapfsMounter) Unmount(env dockerdriver.Env, target string) error {
	logger := env.Logger().Session("unmount")
	logger.Info("unmount-start")
	defer logger.Info("unmount-end")

	target = strings.TrimSuffix(target, "/")
	intermediateMount := target + MapfsDirectorySuffix

	waitError := m.invoker.Invoke(env, "umount", []string{"-l", target}).Wait()
	if waitError != nil {
		return dockerdriver.SafeError{SafeDescription: waitError.Error()}
	}

	if exists, err := m.mountChecker.Exists(intermediateMount); exists {
		err = m.invoker.Invoke(env, "umount", []string{"-l", intermediateMount}).Wait()
		if err != nil {
			logger.Error("warning-umount-intermediate-failed", err)
			return nil
		}

	} else if err != nil {
		logger.Error("warning-umount-check-intermediate-failed", err)
	}

	_, err := m.osshim.Stat(intermediateMount)
	if err == nil {
		if e := m.osshim.Remove(intermediateMount); e != nil {
			return dockerdriver.SafeError{SafeDescription: e.Error()}
		}
	}

	return nil
}

func (m *mapfsMounter) Check(env dockerdriver.Env, name, mountPoint string) bool {
	logger := env.Logger().Session("check")
	logger.Info("check-start")
	defer logger.Info("check-end")

	ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(time.Second*5))
	defer cancel()
	env = driverhttp.EnvWithContext(ctx, env)
	err := m.invoker.Invoke(env, "mountpoint", []string{"-q", mountPoint}).Wait()
	if err != nil {
		logger.Info(fmt.Sprintf("unable to verify volume %s (%s)", name, err.Error()))
		return false
	}
	return true
}

func (m *mapfsMounter) Purge(env dockerdriver.Env, path string) {
	logger := env.Logger().Session("purge")
	logger.Info("purge-start")
	defer logger.Info("purge-end")

	pkillInvokeResult := m.invoker.Invoke(env, "pkill", []string{"mapfs"})
	err := pkillInvokeResult.Wait()
	if err != nil {
		logger.Info("pkill", lager.Data{"err": err.Error(), "output": pkillInvokeResult.StdOutput()})
	} else {
		logger.Debug("pkill", lager.Data{"output": pkillInvokeResult.StdOutput()})
	}

	for i := 0; i < 30; i++ {
		logger.Info("waiting-for-kill")
		time.Sleep(PurgeTimeToSleep)
		invokeResult := m.invoker.Invoke(env, "pgrep", []string{"mapfs"})
		err = invokeResult.Wait()
		if err != nil {
			logger.Info("pgrep", lager.Data{"err": err.Error(), "output": invokeResult.StdOutput()})
			break
		}
		logger.Debug("pgrep", lager.Data{"output": invokeResult.StdOutput()})
	}

	mountPattern, err := regexp.Compile("^" + path + ".*" + MapfsDirectorySuffix + "$")
	if err != nil {
		logger.Error("unable-to-list-mounts", err)
		return
	}

	mounts, err := m.mountChecker.List(mountPattern)
	if err != nil {
		logger.Error("check-proc-mounts-failed", err, lager.Data{"path": path})
		return
	}

	logger.Info("mount-directory-list", lager.Data{"mounts": mounts})

	for _, mountDir := range mounts {
		realMountpoint := strings.TrimSuffix(mountDir, MapfsDirectorySuffix)

		err = m.invoker.Invoke(env, "umount", []string{"-l", "-f", realMountpoint}).Wait()
		if err != nil {
			logger.Error("warning-umount-command-intermediate-failed", err)
		}

		logger.Info("unmount-successful", lager.Data{"path": realMountpoint})

		if err := m.osshim.Remove(realMountpoint); err != nil {
			logger.Error("purge-cannot-remove-directory", err, lager.Data{"name": realMountpoint, "path": path})
		}

		logger.Info("remove-directory-successful", lager.Data{"path": realMountpoint})

		err = m.invoker.Invoke(env, "umount", []string{"-l", "-f", mountDir}).Wait()
		if err != nil {
			logger.Error("warning-umount-mapfs-failed", err)
		}

		logger.Info("unmount-successful", lager.Data{"path": mountDir})

		if err := m.osshim.Remove(mountDir); err != nil {
			logger.Error("purge-cannot-remove-directory", err, lager.Data{"name": mountDir, "path": path})
		}

		logger.Info("remove-directory-successful", lager.Data{"path": mountDir})
	}
}

func NewMapFsVolumeMountMask() (vmo.MountOptsMask, error) {
	allowed := []string{"auto_cache", "mount", "source", "experimental", "uid", "gid", "username", "password", "readonly", "version", "cache"}

	defaultMap := map[string]interface{}{
		"auto_cache": "true",
	}

	return vmo.NewMountOptsMask(
		allowed,
		defaultMap,
		nil,
		[]string{},
		[]string{},
	)

}

func uniformData(data interface{}) string {
	switch v := data.(type) {
	case int:
		return strconv.FormatInt(int64(v), 10)

	case string:
		return data.(string)
	}

	return ""
}

func mapfsOptions(opts vmo.MountOpts) []string {
	var ret []string
	if uid, ok := opts["uid"]; ok {
		ret = append(ret, "-uid", uniformData(uid))
	}
	if gid, ok := opts["gid"]; ok {
		ret = append(ret, "-gid", uniformData(gid))
	}
	if _, ok := opts["auto_cache"]; ok {
		ret = append(ret, "-auto_cache")
	}
	return ret
}
