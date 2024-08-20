package driveradmin

import (
	"code.cloudfoundry.org/dockerdriver"
	"github.com/tedsuo/rata"
)

const (
	EvacuateRoute = "evacuate"
	PingRoute     = "ping"
)

var Routes = rata.Routes{
	{Path: "/evacuate", Method: "GET", Name: EvacuateRoute},
	{Path: "/ping", Method: "GET", Name: PingRoute},
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o ../nfsdriverfakes/fake_driver_admin.go . DriverAdmin

type DriverAdmin interface {
	Evacuate(env dockerdriver.Env) ErrorResponse
	Ping(env dockerdriver.Env) ErrorResponse
}

type ErrorResponse struct {
	Err string
}

//counterfeiter:generate -o ../nfsdriverfakes/fake_drainable.go . Drainable
type Drainable interface {
	Drain(env dockerdriver.Env) error
}
