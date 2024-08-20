package volume_mount_options

import (
	"fmt"
	"strconv"

	"code.cloudfoundry.org/volume-mount-options/utils"
)

type MountOptsMask struct {
	Allowed        []string
	Defaults       map[string]interface{}
	KeyPerms       map[string]string
	Ignored        []string
	Mandatory      []string
	SloppyMount    bool
	ValidationFunc []UserOptsValidation
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . UserOptsValidation
type UserOptsValidation interface {
	Validate(string, string) error
}

type UserOptsValidationFunc func(string, string) error

func (v UserOptsValidationFunc) Validate(a string, b string) error {
	return v(a, b)
}

func NewMountOptsMask(allowed []string,
	defaults map[string]interface{},
	keyPerms map[string]string,
	ignored, mandatory []string,
	f ...UserOptsValidation) (MountOptsMask, error) {
	mask := MountOptsMask{
		Allowed:        allowed,
		Defaults:       defaults,
		KeyPerms:       keyPerms,
		Ignored:        ignored,
		Mandatory:      mandatory,
		ValidationFunc: f,
	}

	if defaults == nil {
		mask.Defaults = make(map[string]interface{})
	}

	if v, ok := defaults["sloppy_mount"]; ok {
		vc := utils.InterfaceToString(v)

		var err error
		mask.SloppyMount, err = strconv.ParseBool(vc)

		if err != nil {
			return MountOptsMask{}, fmt.Errorf("Invalid sloppy_mount option: %w", err)
		}
	}

	return mask, nil
}
