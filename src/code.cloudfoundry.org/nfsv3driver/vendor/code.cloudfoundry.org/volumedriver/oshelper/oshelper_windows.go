//go:build windows
// +build windows

package oshelper

import "code.cloudfoundry.org/volumedriver"

type osHelper struct {
}

func NewOsHelper() volumedriver.OsHelper {
	return &osHelper{}
}

func (o *osHelper) Umask(mask int) (oldmask int) {
	return 0
}
