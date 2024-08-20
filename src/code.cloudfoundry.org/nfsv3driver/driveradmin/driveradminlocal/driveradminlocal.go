package driveradminlocal

import (
	"os"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/nfsv3driver/driveradmin"
	"github.com/tedsuo/ifrit"
)

type DriverAdminLocal struct {
	serverProcess ifrit.Process
	drainables    []driveradmin.Drainable
}

func NewDriverAdminLocal() *DriverAdminLocal {
	d := &DriverAdminLocal{}

	return d
}

func (d *DriverAdminLocal) SetServerProc(p ifrit.Process) {
	d.serverProcess = p
}

func (d *DriverAdminLocal) RegisterDrainable(rhs driveradmin.Drainable) {
	d.drainables = append(d.drainables, rhs)
}

func (d *DriverAdminLocal) Evacuate(env dockerdriver.Env) driveradmin.ErrorResponse {
	logger := env.Logger().Session("evacuate")
	logger.Info("start")
	defer logger.Info("end")

	if d.serverProcess == nil {
		return driveradmin.ErrorResponse{Err: "unexpected error: server process not found"}
	}

	for _, svr := range d.drainables {
		if err := svr.Drain(env); err != nil {
			logger.Error("failed-draining", err)
		}
	}

	d.serverProcess.Signal(os.Interrupt)

	return driveradmin.ErrorResponse{}
}

func (d *DriverAdminLocal) Ping(env dockerdriver.Env) driveradmin.ErrorResponse {
	logger := env.Logger().Session("ping")
	logger.Info("start")
	defer logger.Info("end")

	return driveradmin.ErrorResponse{}
}
