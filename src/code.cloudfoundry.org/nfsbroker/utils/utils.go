package utils

import (
	"os"

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
