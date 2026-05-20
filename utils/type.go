package utils

func IsInstanceOf[T any](obj any) bool {
	_, ok := obj.(T)
	return ok
}
