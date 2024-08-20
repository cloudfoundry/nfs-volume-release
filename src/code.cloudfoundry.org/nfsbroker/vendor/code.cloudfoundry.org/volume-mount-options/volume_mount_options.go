package volume_mount_options

import (
	"fmt"
	"strconv"
	"strings"
)

const ValidationFailErrorMessage = "- validation mount options failed: %s"
const NotAllowedErrorMessage = "- Not allowed options: %s"
const MissingOptionErrorMessage = "- Missing mandatory options: %s"

type MountOpts map[string]interface{}

func NewMountOpts(userOpts map[string]interface{}, mask MountOptsMask) (MountOpts, error) {
	mountOpts := make(map[string]interface{})
	for k, v := range mask.Defaults {
		mountOpts[k] = v
	}

	allowedErrorList := []string{}
	for k, v := range userOpts {
		var canonicalKey string
		var ok bool
		if canonicalKey, ok = mask.KeyPerms[k]; !ok {
			canonicalKey = k
		}

		if inArray(mask.Ignored, canonicalKey) {
			continue
		}

		if inArray(mask.Allowed, canonicalKey) {
			uv := uniformKeyData(canonicalKey, v)
			mountOpts[canonicalKey] = uv
		} else if !mask.SloppyMount {
			allowedErrorList = append(allowedErrorList, k)
		}
	}

	var validationErrorList []string
	if mask.ValidationFunc != nil {
		for key, val := range mountOpts {
			for _, validationFunc := range mask.ValidationFunc {
				err := validationFunc.Validate(key, fmt.Sprintf("%v", val))
				if err != nil {
					validationErrorList = append(validationErrorList, err.Error())
				}
			}
		}
	}

	var mandatoryErrorList []string
	for _, k := range mask.Mandatory {
		if _, ok := mountOpts[k]; !ok {
			mandatoryErrorList = append(mandatoryErrorList, k)
		}
	}

	errorString := buildErrorMessage(validationErrorList, ValidationFailErrorMessage)
	errorString += buildErrorMessage(allowedErrorList, NotAllowedErrorMessage)
	errorString += buildErrorMessage(mandatoryErrorList, MissingOptionErrorMessage)

	if hasErrors(allowedErrorList, validationErrorList, mandatoryErrorList) {
		return MountOpts{}, fmt.Errorf(errorString)
	}

	return mountOpts, nil
}

func hasErrors(allowedErrorList []string, validationErrorList []string, mandatoryErrorList []string) bool {
	return len(allowedErrorList) > 0 || len(validationErrorList) > 0 || len(mandatoryErrorList) > 0
}

func buildErrorMessage(validationErrorList []string, errorDesc string) string {
	if len(validationErrorList) > 0 {
		return fmt.Sprintln(fmt.Sprintf(errorDesc, strings.Join(validationErrorList, ", ")))
	}
	return ""
}

func inArray(list []string, key string) bool {
	for _, k := range list {
		if k == key {
			return true
		}
	}

	return false
}

func uniformKeyData(key string, data interface{}) string {
	switch key {
	case "auto-traverse-mounts":
		return uniformData(data, true)

	case "dircache":
		return uniformData(data, true)

	}

	return uniformData(data, false)
}

func uniformData(data interface{}, boolAsInt bool) string {
	switch data.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return fmt.Sprintf("%#v", data)

	case string:
		return data.(string)

	case bool:
		if boolAsInt {
			if data.(bool) {
				return "1"
			} else {
				return "0"
			}
		} else {
			return strconv.FormatBool(data.(bool))
		}
	}

	return ""
}
