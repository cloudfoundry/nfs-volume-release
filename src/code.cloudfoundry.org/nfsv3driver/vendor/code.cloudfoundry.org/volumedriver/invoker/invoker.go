package invoker

import (
	"errors"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/lager/v3"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o ../invokerfakes/fake_invoke_result.go . InvokeResult
type InvokeResult interface {
	StdError() string
	StdOutput() string
	Wait() error
	WaitFor(string, time.Duration) error
}

//counterfeiter:generate -o ../invokerfakes/fake_invoker.go . Invoker
type Invoker interface {
	Invoke(env dockerdriver.Env, executable string, args []string, envVars ...string) InvokeResult
}

type invokeResult struct {
	cmdDone      *bool
	cmd          *exec.Cmd
	outputBuffer *Buffer
	errorBuffer  *Buffer
	logger       lager.Logger
	invokeErr    error
}

func (i invokeResult) StdError() string {
	return i.errorBuffer.String()
}

func (i invokeResult) StdOutput() string {
	return i.outputBuffer.String()
}

func (i invokeResult) Wait() error {
	if i.invokeErr != nil {
		return i.invokeErr
	}
	wait := i.cmd.Wait()
	*i.cmdDone = true
	return wait
}

func (i invokeResult) WaitFor(stringToWaitFor string, duration time.Duration) error {
	if i.invokeErr != nil {
		return i.invokeErr
	}
	var errChan = make(chan error, 1)
	go func() {
		err := i.cmd.Wait()
		if err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	timeout := time.After(duration)
	for {
		select {
		case e := <-errChan:
			if e == nil && !i.isExpectedTextContainedInStdOut(stringToWaitFor) {
				return errors.New("command finished without expected Text")
			}
			return e
		case <-timeout:
			err := syscall.Kill(-i.cmd.Process.Pid, syscall.SIGKILL)
			if err != nil {
				i.logger.Info("command-sigkill-error", lager.Data{"desc": err.Error()})
			}
			return errors.New("command timed out")
		default:
			if i.isExpectedTextContainedInStdOut(stringToWaitFor) {
				*i.cmdDone = true
				return nil
			}
		}
	}
}

func (i invokeResult) isExpectedTextContainedInStdOut(stringToWaitFor string) bool {
	return strings.Contains(i.StdOutput(), stringToWaitFor)
}
