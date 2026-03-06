package utils

import "time"

func TryEnqueue[T any](c chan<- T, data T) bool {
	select {
	case c <- data:
		return true
	default:
		return false
	}
}

func TryDequeue[T any](c <-chan T) (*T, bool) {
	select {
	case ret := <-c:
		return &ret, true
	default:
		return nil, false
	}
}

func EnqueueWithTimeout[T any](c chan<- T, data T, timeout time.Duration) bool {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case c <- data:
		return true
	case <-timer.C:
		return false
	}
}

func DequeueWithTimeout[T any](c <-chan T, timeout time.Duration) (*T, bool) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case ret := <-c:
		return &ret, true
	case <-timer.C:
		return nil, false
	}
}
