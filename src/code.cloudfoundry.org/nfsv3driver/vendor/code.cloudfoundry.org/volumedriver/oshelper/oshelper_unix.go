//go:build linux || darwin
// +build linux darwin

package oshelper

import (
	"syscall"

	"code.cloudfoundry.org/volumedriver"
)

type osHelper struct {
}

func NewOsHelper() volumedriver.OsHelper {
	return &osHelper{}
}

func (o *osHelper) Umask(mask int) (oldmask int) {
	return syscall.Umask(mask)
}
