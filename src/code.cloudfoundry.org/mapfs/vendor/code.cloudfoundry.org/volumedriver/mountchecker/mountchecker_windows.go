package mountchecker

import (
	"os"

	"code.cloudfoundry.org/goshims/bufioshim"
	"code.cloudfoundry.org/goshims/osshim"
)

type MountChecker interface {
	Exists(string) (bool, error)
	List(string) ([]string, error)
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
	_, err := c.os.Stat(mountPath)
	if err != nil {
		if err == os.ErrNotExist {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c Checker) List(mountPathRegexp string) ([]string, error) {
	return []string{}, nil
}
