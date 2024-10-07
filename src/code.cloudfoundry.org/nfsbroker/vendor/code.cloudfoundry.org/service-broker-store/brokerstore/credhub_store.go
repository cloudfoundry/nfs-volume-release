package brokerstore

import (
	"encoding/json"
	"fmt"

	"code.cloudfoundry.org/credhub-cli/credhub/credentials"
	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/service-broker-store/brokerstore/credhub_shims"
	"github.com/pivotal-cf/brokerapi/v11/domain"
)

type CredhubStore struct {
	logger      lager.Logger
	credhubShim credhub_shims.Credhub
	storeID     string
}

func NewCredhubStore(logger lager.Logger, credhubShim credhub_shims.Credhub, storeID string) *CredhubStore {
	return &CredhubStore{
		logger:      logger,
		credhubShim: credhubShim,
		storeID:     storeID,
	}
}

func (s *CredhubStore) Activate() error {
	s.logger.Info("activating-credhub")
	_, err := s.credhubShim.SetValue(s.namespaced("migrated-from-sql"), "true")
	if err != nil {
		return err
	}

	return nil
}

func (s *CredhubStore) IsActivated() (bool, error) {
	logger := s.logger.Session("is-activated")
	logger.Info("start")
	defer logger.Info("end")

	results, err := s.credhubShim.FindByPath(s.namespaced("migrated-from-sql"))
	if err != nil {
		return false, err
	}

	return len(results.Credentials) > 0, nil
}

func (s *CredhubStore) CreateInstanceDetails(id string, details ServiceInstance) error {
	logger := s.logger.Session("create-instance-details")
	logger.Info("start")
	defer logger.Info("end")

	mappedDetails, err := toMap(details)
	if err != nil {
		return err
	}
	_, err = s.credhubShim.SetJSON(s.namespaced(id), mappedDetails)
	if err != nil {
		return err
	}
	return nil
}

func (s *CredhubStore) RetrieveInstanceDetails(id string) (ServiceInstance, error) {
	logger := s.logger.Session("retrieve-instance-details")
	logger.Info("start")
	defer logger.Info("end")

	creds, err := s.credhubShim.GetLatestJSON(s.namespaced(id))
	if err != nil {
		return ServiceInstance{}, err
	}

	var serviceInstance ServiceInstance
	err = toStruct(creds, &serviceInstance)
	if err != nil {
		return ServiceInstance{}, err
	}

	return serviceInstance, nil
}

func (s *CredhubStore) RetrieveBindingDetails(id string) (domain.BindDetails, error) {
	logger := s.logger.Session("retrieve-binding-details")
	logger.Info("start")
	defer logger.Info("end")

	creds, err := s.credhubShim.GetLatestJSON(s.namespaced(id))
	if err != nil {
		return domain.BindDetails{}, err
	}

	var bindDetails domain.BindDetails
	err = toStruct(creds, &bindDetails)
	if err != nil {
		return domain.BindDetails{}, err
	}

	return bindDetails, nil
}

func (s *CredhubStore) RetrieveAllInstanceDetails() (map[string]ServiceInstance, error) {
	panic("Not Implemented")
}

func (s *CredhubStore) RetrieveAllBindingDetails() (map[string]domain.BindDetails, error) {
	panic("Not Implemented")
}

func (s *CredhubStore) CreateBindingDetails(id string, details domain.BindDetails) error {
	logger := s.logger.Session("create-binding-details")
	logger.Info("start")
	defer logger.Info("end")

	mappedDetails, err := toMap(details)
	if err != nil {
		return err
	}

	_, err = s.credhubShim.SetJSON(s.namespaced(id), mappedDetails)
	if err != nil {
		return err
	}
	return nil
}

func (s *CredhubStore) DeleteInstanceDetails(id string) error {
	logger := s.logger.Session("delete-instance-details")
	logger.Info("start")
	defer logger.Info("end")

	return s.credhubShim.Delete(s.namespaced(id))
}
func (s *CredhubStore) DeleteBindingDetails(id string) error {
	logger := s.logger.Session("delete-binding-details")
	logger.Info("start")
	defer logger.Info("end")

	return s.credhubShim.Delete(s.namespaced(id))
}

func (s *CredhubStore) IsInstanceConflict(id string, details ServiceInstance) bool {
	return isInstanceConflict(s, id, details)
}
func (s *CredhubStore) IsBindingConflict(id string, details domain.BindDetails) bool {
	return isBindingConflict(s, id, details)
}

func (s *CredhubStore) Restore(logger lager.Logger) error {
	return nil
}

func (s *CredhubStore) Save(logger lager.Logger) error {
	return nil
}

func (s *CredhubStore) Cleanup() error {
	return nil
}

func (s *CredhubStore) namespaced(id string) string {
	return fmt.Sprintf("/%s/%s", s.storeID, id)
}

func toMap(subject interface{}) (map[string]interface{}, error) {
	var inInterface map[string]interface{}

	marshalledJson, err := json.Marshal(subject)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(marshalledJson, &inInterface)
	if err != nil {
		return nil, err
	}

	return inInterface, nil
}

func toStruct(creds credentials.JSON, target interface{}) error {
	//var serviceInstance ServiceInstance

	credsBytes, err := json.Marshal(creds.Value)
	if err != nil {
		return err
	}

	err = json.Unmarshal(credsBytes, &target)
	if err != nil {
		return err
	}

	return nil
}
