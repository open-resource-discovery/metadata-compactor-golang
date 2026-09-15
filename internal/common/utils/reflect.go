package utils

import (
	"reflect"
)

func SafeCast[T any](value any) (result T) {
	if rtype, rvalue := reflect.TypeOf(result), reflect.ValueOf(value); value != nil && rvalue.CanConvert(rtype) {
		result = rvalue.Convert(rtype).Interface().(T)
	}

	return result
}
