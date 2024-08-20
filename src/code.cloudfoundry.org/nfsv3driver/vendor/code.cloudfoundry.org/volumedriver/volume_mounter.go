package volumedriver

import (
	"code.cloudfoundry.org/dockerdriver"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o volumedriverfakes/fake_mounter.go . Mounter
type Mounter interface {
	Mount(env dockerdriver.Env, source string, target string, opts map[string]interface{}) error
	Unmount(env dockerdriver.Env, target string) error
	Check(env dockerdriver.Env, name, mountPoint string) bool
	Purge(env dockerdriver.Env, path string)
}
