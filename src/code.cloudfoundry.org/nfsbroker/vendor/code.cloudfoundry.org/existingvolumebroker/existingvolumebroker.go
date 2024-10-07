package existingvolumebroker

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"sync"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/service-broker-store/brokerstore"
	vmo "code.cloudfoundry.org/volume-mount-options"
	vmou "code.cloudfoundry.org/volume-mount-options/utils"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"github.com/pivotal-cf/brokerapi/v11/domain/apiresponses"
)

const (
	DEFAULT_CONTAINER_PATH = "/var/vcap/data"
	SHARE_KEY              = "share"
	SOURCE_KEY             = "source"
	VERSION_KEY            = "version"
)

type lock interface {
	Lock()
	Unlock()
}

type BrokerType int

const (
	BrokerTypeNFS BrokerType = iota
	BrokerTypeSMB
)

type Broker struct {
	brokerType              BrokerType
	logger                  lager.Logger
	os                      osshim.Os
	mutex                   lock
	clock                   clock.Clock
	store                   brokerstore.Store
	services                Services
	configMask              vmo.MountOptsMask
	DisallowedBindOverrides []string
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/fake_services.go . Services
type Services interface {
	List() []domain.Service
}

func New(
	brokerType BrokerType,
	logger lager.Logger,
	services Services,
	os osshim.Os,
	clock clock.Clock,
	store brokerstore.Store,
	configMask vmo.MountOptsMask,
) *Broker {
	theBroker := Broker{
		brokerType:              brokerType,
		logger:                  logger,
		os:                      os,
		mutex:                   &sync.Mutex{},
		clock:                   clock,
		store:                   store,
		services:                services,
		configMask:              configMask,
		DisallowedBindOverrides: []string{SHARE_KEY, SOURCE_KEY},
	}

	return &theBroker
}

func (b *Broker) isNFSBroker() bool {
	return b.brokerType == BrokerTypeNFS
}

func (b *Broker) Services(_ context.Context) ([]domain.Service, error) {
	logger := b.logger.Session("services")
	logger.Info("start")
	defer logger.Info("end")

	return b.services.List(), nil
}

func (b *Broker) Provision(context context.Context, instanceID string, details domain.ProvisionDetails, _ bool) (_ domain.ProvisionedServiceSpec, e error) {
	logger := b.logger.Session("provision").WithData(lager.Data{"instanceID": instanceID, "details": details})
	logger.Info("start")
	defer logger.Info("end")

	var configuration map[string]interface{}

	var decoder = json.NewDecoder(bytes.NewBuffer(details.RawParameters))
	err := decoder.Decode(&configuration)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, apiresponses.ErrRawParamsInvalid
	}

	share := stringifyShare(configuration[SHARE_KEY])
	if share == "" {
		return domain.ProvisionedServiceSpec{}, errors.New("config requires a \"share\" key")
	}

	if _, ok := configuration[SOURCE_KEY]; ok {
		return domain.ProvisionedServiceSpec{}, errors.New("create configuration contains the following invalid option: ['" + SOURCE_KEY + "']")
	}

	if b.isNFSBroker() {
		re := regexp.MustCompile("^[^/]+:/")
		match := re.MatchString(share)

		if match {
			return domain.ProvisionedServiceSpec{}, errors.New("syntax error for share: no colon allowed after server")
		}
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()
	defer func() {
		out := b.store.Save(logger)
		if e == nil {
			e = out
		}
	}()

	instanceDetails := brokerstore.ServiceInstance{
		ServiceID:          details.ServiceID,
		PlanID:             details.PlanID,
		OrganizationGUID:   details.OrganizationGUID,
		SpaceGUID:          details.SpaceGUID,
		ServiceFingerPrint: configuration,
	}

	if b.instanceConflicts(instanceDetails, instanceID) {
		return domain.ProvisionedServiceSpec{}, apiresponses.ErrInstanceAlreadyExists
	}

	err = b.store.CreateInstanceDetails(instanceID, instanceDetails)
	if err != nil {
		return domain.ProvisionedServiceSpec{}, fmt.Errorf("failed to store instance details: %s", err.Error())
	}

	logger.Info("service-instance-created", lager.Data{"instanceDetails": instanceDetails})

	return domain.ProvisionedServiceSpec{IsAsync: false}, nil
}

func (b *Broker) Deprovision(context context.Context, instanceID string, details domain.DeprovisionDetails, _ bool) (_ domain.DeprovisionServiceSpec, e error) {
	logger := b.logger.Session("deprovision")
	logger.Info("start")
	defer logger.Info("end")

	b.mutex.Lock()
	defer b.mutex.Unlock()
	defer func() {
		out := b.store.Save(logger)
		if e == nil {
			e = out
		}
	}()

	_, err := b.store.RetrieveInstanceDetails(instanceID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, apiresponses.ErrInstanceDoesNotExist
	}

	err = b.store.DeleteInstanceDetails(instanceID)
	if err != nil {
		return domain.DeprovisionServiceSpec{}, err
	}

	return domain.DeprovisionServiceSpec{IsAsync: false, OperationData: "deprovision"}, nil
}

func (b *Broker) Bind(context context.Context, instanceID string, bindingID string, bindDetails domain.BindDetails, _ bool) (_ domain.Binding, e error) {
	logger := b.logger.Session("bind")
	logger.Info("start", lager.Data{"bindingID": bindingID, "details": bindDetails})
	defer logger.Info("end")

	b.mutex.Lock()
	defer b.mutex.Unlock()
	defer func() {
		out := b.store.Save(logger)
		if e == nil {
			e = out
		}
	}()

	logger.Info("starting-broker-bind")
	instanceDetails, err := b.store.RetrieveInstanceDetails(instanceID)
	if err != nil {
		return domain.Binding{}, apiresponses.ErrInstanceDoesNotExist
	}

	if bindDetails.AppGUID == "" {
		return domain.Binding{}, apiresponses.ErrAppGuidNotProvided
	}

	opts, err := getFingerprint(instanceDetails.ServiceFingerPrint)
	if err != nil {
		return domain.Binding{}, err
	}

	var bindOpts map[string]interface{}
	if len(bindDetails.RawParameters) > 0 {
		if err = json.Unmarshal(bindDetails.RawParameters, &bindOpts); err != nil {
			return domain.Binding{}, err
		}
	}

	for k, v := range bindOpts {
		for _, disallowed := range b.DisallowedBindOverrides {
			if k == disallowed {
				err := fmt.Errorf("bind configuration contains the following invalid option: ['%s']", k)
				logger.Error("err-override-not-allowed-in-bind", err, lager.Data{"key": k})
				return domain.Binding{}, apiresponses.NewFailureResponse(
					err, http.StatusBadRequest, "invalid-raw-params",
				)

			}
		}
		opts[k] = v
	}

	mode, err := evaluateMode(opts)
	if err != nil {
		logger.Error("error-evaluating-mode", err)
		return domain.Binding{}, err
	}

	mountOpts, err := vmo.NewMountOpts(opts, b.configMask)
	if err != nil {
		logger.Error("error-generating-mount-options", err)
		return domain.Binding{}, apiresponses.NewFailureResponse(err, http.StatusBadRequest, "invalid-params")
	}

	if b.bindingConflicts(bindingID, bindDetails) {
		return domain.Binding{}, apiresponses.ErrBindingAlreadyExists
	}

	logger.Info("retrieved-instance-details", lager.Data{"instanceDetails": instanceDetails})

	err = b.store.CreateBindingDetails(bindingID, bindDetails)
	if err != nil {
		return domain.Binding{}, err
	}

	driverName := "smbdriver"
	if b.isNFSBroker() {
		driverName = "nfsv3driver"

		// for backwards compatibility the nfs flavor has to issue source strings
		// with nfs:// prefix (otherwise the mapfs-mounter wont construct the correct
		// mount string for the kernel mount
		//
		// see (https://github.com/cloudfoundry/nfsv3driver/blob/ac1e1d26fec9a8551cacfabafa6e035f233c83e0/mapfs_mounter.go#L121)
		mountOpts[SOURCE_KEY] = fmt.Sprintf("nfs://%s", mountOpts[SOURCE_KEY])
	}

	logger.Debug("volume-service-binding", lager.Data{"driver": driverName, "mountOpts": mountOpts})

	s, err := b.hash(mountOpts)
	if err != nil {
		logger.Error("error-calculating-volume-id", err, lager.Data{"config": mountOpts, "bindingID": bindingID, "instanceID": instanceID})
		return domain.Binding{}, err
	}
	volumeId := fmt.Sprintf("%s-%s", instanceID, s)

	mountConfig := map[string]interface{}{}

	for k, v := range mountOpts {
		mountConfig[k] = v
	}

	ret := domain.Binding{
		Credentials: struct{}{}, // if nil, cloud controller chokes on response
		VolumeMounts: []domain.VolumeMount{{
			ContainerDir: evaluateContainerPath(opts, instanceID),
			Mode:         mode,
			Driver:       driverName,
			DeviceType:   "shared",
			Device: domain.SharedDevice{
				VolumeId:    volumeId,
				MountConfig: mountConfig,
			},
		}},
	}
	return ret, nil
}

func (b *Broker) hash(mountOpts map[string]interface{}) (string, error) {
	var (
		bytes []byte
		err   error
	)
	if bytes, err = json.Marshal(mountOpts); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(bytes)), nil
}

func (b *Broker) Unbind(context context.Context, instanceID string, bindingID string, details domain.UnbindDetails, _ bool) (_ domain.UnbindSpec, e error) {
	logger := b.logger.Session("unbind")
	logger.Info("start")
	defer logger.Info("end")

	b.mutex.Lock()
	defer b.mutex.Unlock()
	defer func() {
		out := b.store.Save(logger)
		if e == nil {
			e = out
		}
	}()

	if _, err := b.store.RetrieveInstanceDetails(instanceID); err != nil {
		return domain.UnbindSpec{}, apiresponses.ErrInstanceDoesNotExist
	}

	if _, err := b.store.RetrieveBindingDetails(bindingID); err != nil {
		return domain.UnbindSpec{}, apiresponses.ErrBindingDoesNotExist
	}

	if err := b.store.DeleteBindingDetails(bindingID); err != nil {
		return domain.UnbindSpec{}, err
	}
	return domain.UnbindSpec{}, nil
}

func (b *Broker) Update(context context.Context, instanceID string, details domain.UpdateDetails, _ bool) (domain.UpdateServiceSpec, error) {
	return domain.UpdateServiceSpec{},
		apiresponses.NewFailureResponse(
			errors.New("this service does not support instance updates. Please delete your service instance and create a new one with updated configuration"),
			422,
			"",
		)
}

func (b *Broker) LastOperation(_ context.Context, instanceID string, _ domain.PollDetails) (domain.LastOperation, error) {
	logger := b.logger.Session("last-operation").WithData(lager.Data{"instanceID": instanceID})
	logger.Info("start")
	defer logger.Info("end")

	b.mutex.Lock()
	defer b.mutex.Unlock()

	return domain.LastOperation{}, errors.New("unrecognized operationData")
}

func (b *Broker) GetInstance(ctx context.Context, instanceID string, details domain.FetchInstanceDetails) (domain.GetInstanceDetailsSpec, error) {
	panic("implement me")
}

func (b *Broker) LastBindingOperation(ctx context.Context, instanceID, bindingID string, details domain.PollDetails) (domain.LastOperation, error) {
	panic("implement me")
}

func (b *Broker) GetBinding(ctx context.Context, instanceID, bindingID string, details domain.FetchBindingDetails) (domain.GetBindingSpec, error) {
	panic("implement me")
}

func (b *Broker) instanceConflicts(details brokerstore.ServiceInstance, instanceID string) bool {
	return b.store.IsInstanceConflict(instanceID, brokerstore.ServiceInstance(details))
}

func (b *Broker) bindingConflicts(bindingID string, details domain.BindDetails) bool {
	return b.store.IsBindingConflict(bindingID, details)
}

func evaluateContainerPath(parameters map[string]interface{}, volId string) string {
	if containerPath, ok := parameters["mount"]; ok && containerPath != "" {
		return containerPath.(string)
	}

	return path.Join(DEFAULT_CONTAINER_PATH, volId)
}

func evaluateMode(parameters map[string]interface{}) (string, error) {
	if ro, ok := parameters["readonly"]; ok {
		roc := vmou.InterfaceToString(ro)
		if roc == "true" {
			return "r", nil
		}

		return "", apiresponses.NewFailureResponse(fmt.Errorf("invalid ro parameter value: %q", roc), http.StatusBadRequest, "invalid-ro-param")
	}

	return "rw", nil
}

func getFingerprint(rawObject interface{}) (map[string]interface{}, error) {
	fingerprint, ok := rawObject.(map[string]interface{})
	if ok {
		return fingerprint, nil
	} else {
		// legacy service instances only store the "share" key in the service fingerprint.
		share, ok := rawObject.(string)
		if ok {
			return map[string]interface{}{SHARE_KEY: share}, nil
		}
		return nil, errors.New("unable to deserialize service fingerprint")
	}
}

func stringifyShare(data interface{}) string {
	if val, ok := data.(string); ok {
		return val
	}

	return ""
}
