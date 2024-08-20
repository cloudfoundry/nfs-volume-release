//go:build linux || darwin
// +build linux darwin

package mountchecker

import (
	"io"
	"regexp"
	"strings"

	"code.cloudfoundry.org/goshims/bufioshim"
	"code.cloudfoundry.org/goshims/osshim"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o ../volumedriverfakes/fake_mount_checker.go . MountChecker

type MountChecker interface {
	Exists(string) (bool, error)
	List(*regexp.Regexp) ([]string, error)
}

type Checker struct {
	bufio bufioshim.Bufio
	os    osshim.Os

	mounts []string
}

func NewChecker(bufio bufioshim.Bufio, os osshim.Os) Checker {
	return Checker{
		bufio: bufio,
		os:    os,
	}
}

func (c Checker) Exists(mountPath string) (bool, error) {
	err := c.loadProcMounts()
	if err != nil {
		return false, err
	}

	for _, mount := range c.mounts {
		if mountPath == mount {
			return true, nil
		}
	}

	return false, nil
}

func (c Checker) List(pattern *regexp.Regexp) ([]string, error) {
	err := c.loadProcMounts()
	if err != nil {
		return []string{}, err
	}

	mounts := []string{}

	for _, mount := range c.mounts {
		exists := pattern.MatchString(mount)

		if exists {
			mounts = append(mounts, mount)
		}
	}

	return mounts, nil
}

// The named return of the error is required to allow the error from the
// defered file close to be returned.
func (c *Checker) loadProcMounts() (err error) {
	var file osshim.File
	file, err = c.os.Open("/proc/mounts")
	if err != nil {
		return err
	}

	defer func(err *error) {
		e := file.Close()
		if *err == nil {
			*err = e
		}
	}(&err)

	reader := c.bufio.NewReader(file)
	var (
		line    string
		readErr error
	)

	for {
		line, readErr = reader.ReadString('\n')
		if readErr != nil {
			break
		}

		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}

		c.mounts = append(c.mounts, parts[1])
	}

	if readErr != io.EOF {
		err = readErr
	}

	return err
}
