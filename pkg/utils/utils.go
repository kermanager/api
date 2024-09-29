package utils

import (
	"fmt"
)

func GetIntFromMap(input map[string]interface{}, key string) (int, error) {
	value, ok := input[key]
	if !ok || value == nil {
		return 0, fmt.Errorf("%s is missing or nil", key)
	}

	floatValue, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("%s is not a valid number", key)
	}

	return int(floatValue), nil
}
