package brokerstore

import (
	"encoding/json"
	"errors"
	"reflect"

	"code.cloudfoundry.org/lager/v3"
	"code.cloudfoundry.org/service-broker-store/brokerstore/credhub_shims"
	"github.com/pivotal-cf/brokerapi/v11/domain"
	"golang.org/x/crypto/bcrypt"
)

type ServiceInstance struct {
	ServiceID          string `json:"service_id"`
	PlanID             string `json:"plan_id"`
	OrganizationGUID   string `json:"organization_guid"`
	SpaceGUID          string `json:"space_guid"`
	ServiceFingerPrint interface{}
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o ./brokerstorefakes/fake_store.go . Store
type Store interface {
	RetrieveInstanceDetails(id string) (ServiceInstance, error)
	RetrieveBindingDetails(id string) (domain.BindDetails, error)

	RetrieveAllInstanceDetails() (map[string]ServiceInstance, error)
	RetrieveAllBindingDetails() (map[string]domain.BindDetails, error)

	CreateInstanceDetails(id string, details ServiceInstance) error
	CreateBindingDetails(id string, details domain.BindDetails) error

	DeleteInstanceDetails(id string) error
	DeleteBindingDetails(id string) error

	IsInstanceConflict(id string, details ServiceInstance) bool
	IsBindingConflict(id string, details domain.BindDetails) bool

	Restore(logger lager.Logger) error
	Save(logger lager.Logger) error
	Cleanup() error
}

func NewStore(
	logger lager.Logger,
	credhubURL,
	credhubCACert,
	clientID,
	clientSecret,
	uaaCACert string,
	storeID string,
) Store {
	if credhubURL != "" {
		ch, err := credhub_shims.NewCredhubShim(credhubURL, credhubCACert, clientID, clientSecret, uaaCACert, &credhub_shims.CredhubAuthShim{})
		if err != nil {
			logger.Fatal("failed-creating-credhub-store", err)
		}
		return NewCredhubStore(logger, ch, storeID)
	}
	logger.Fatal("failed-creating-broker-store", errors.New("invalid brokerstore configuration"))
	return nil
}

// Utility methods for storing bindings with secrets stripped out
const HashKey = "paramsHash"

func isInstanceConflict(s Store, id string, details ServiceInstance) bool {
	if existing, err := s.RetrieveInstanceDetails(id); err == nil {
		if !reflect.DeepEqual(details, existing) {
			return true
		}
	}
	return false
}

func isBindingConflict(s Store, id string, details domain.BindDetails) bool {
	if existing, err := s.RetrieveBindingDetails(id); err == nil {
		if existing.AppGUID != details.AppGUID {
			return true
		}
		if existing.PlanID != details.PlanID {
			return true
		}
		if existing.ServiceID != details.ServiceID {
			return true
		}
		if !reflect.DeepEqual(details.BindResource, existing.BindResource) {
			return true
		}
		if (len(details.RawParameters) == 0) && (len(existing.RawParameters) == 0) {
			return false
		}
		if (len(details.RawParameters) == 0) || (len(existing.RawParameters) == 0) {
			return true
		}

		var opts map[string]interface{}
		if err := json.Unmarshal(existing.RawParameters, &opts); err != nil {
			return false
		}

		if bcrypt.CompareHashAndPassword([]byte(opts[HashKey].(string)), details.RawParameters) != nil {
			return true
		}
	}
	return false
}
