package main

import (
	"encoding/json"
	"os"

	"github.com/pivotal-cf/brokerapi/v11/domain"
)

type Services interface {
	List() []domain.Service
}

type services struct {
	services []domain.Service
}

func NewServicesFromConfig(pathToServicesConfig string) (Services, error) {
	/* #nosec */
	contents, err := os.ReadFile(pathToServicesConfig)
	if err != nil {
		return nil, err
	}

	var s []domain.Service
	err = json.Unmarshal(contents, &s)
	if err != nil {
		return nil, err
	}

	return &services{s}, nil
}

func (s *services) List() []domain.Service {
	return s.services
}
