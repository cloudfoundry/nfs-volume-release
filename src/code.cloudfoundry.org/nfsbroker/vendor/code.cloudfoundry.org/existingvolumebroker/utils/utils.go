package utils

import (
	"os"

	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/lager/v3"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"
)

func ExitOnFailure(logger lager.Logger, err error) {
	if err != nil {
		logger.Error("fatal-error-aborting", err)
		os.Exit(1)
	}
}

func UntilTerminated(logger lager.Logger, process ifrit.Process) {
	err := <-process.Wait()
	ExitOnFailure(logger, err)
}

func ProcessRunnerFor(servers grouper.Members) ifrit.Runner {
	return sigmon.New(grouper.NewOrdered(os.Interrupt, servers))
}

func IsThereAProxy(os osshim.Os, logger lager.Logger) bool {
	lgr := logger.Session("is-there-a-proxy")
	lgr.Info("start")
	defer lgr.Info("end")

	https_proxy, ok := os.LookupEnv("https_proxy")

	if ok == true && https_proxy != "" {
		lgr.Info("proxy-found", lager.Data{"https_proxy": https_proxy})
		return true
	}

	lgr.Info("no-proxy-found")

	return false
}
