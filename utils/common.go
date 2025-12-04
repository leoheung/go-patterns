package utils

import (
	"reflect"
	"time"
)

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

func IsDigits(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func DelayDo(d time.Duration, fn func()) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	<-timer.C
	fn()
}
