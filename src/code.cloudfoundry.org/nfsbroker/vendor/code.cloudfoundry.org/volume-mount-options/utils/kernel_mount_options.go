package utils

import (
	"strings"
)

func ParseOptionStringToMap(optionString, separator string) map[string]interface{} {
	mountOpts := make(map[string]interface{}, 0)

	if optionString == "" {
		return mountOpts
	}

	opts := strings.Split(optionString, ",")

	for _, opt := range opts {
		optSegments := strings.SplitN(opt, separator, 2)

		if len(optSegments) == 1 {
			mountOpts[optSegments[0]] = ""
		} else {
			mountOpts[optSegments[0]] = optSegments[1]
		}
	}

	return mountOpts
}
