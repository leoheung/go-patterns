package utils

import "reflect"

// IsNil returns true if v is a nil value for nilable kinds or an untyped nil interface.
func IsNil[T any](v T) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() { // nil interface value
		return true
	}
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
