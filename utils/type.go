package utils

func IsInstance[T any](obj any) bool {
	_, ok := obj.(T)
	return ok
}
