package utils

import (
	"errors"
	"fmt"
	"reflect"
)

func AssertNotNilParams(params map[string]any) error {
	var errs []error

	for name, param := range params {
		if isNilParam(param) {
			errs = append(errs, fmt.Errorf("%s does not to be nil", name))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func isNilParam(param any) bool {
	if param == nil {
		return true
	}

	value := reflect.ValueOf(param)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Interface, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
