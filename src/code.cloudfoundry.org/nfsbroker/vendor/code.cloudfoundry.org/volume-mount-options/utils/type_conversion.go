package utils

import (
	"fmt"
	"strconv"
)

func InterfaceToString(input interface{}) string {
	switch t := input.(type) {
	case string:
		return t
	case int64:
		return fmt.Sprintf("%d", input)
	case float64:
		return fmt.Sprintf("%f", input)
	case bool:
		return strconv.FormatBool(input.(bool))
	}
	return ""
}
