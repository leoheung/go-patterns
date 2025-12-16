package utils

import (
	"fmt"
	"time"
)

func TimeoutWork(work func() (any, error), timeout time.Duration) (any, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	type Response struct {
		Data any
		Err  error
	}

	out := make(chan Response, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				out <- Response{
					Data: nil,
					Err:  fmt.Errorf("panic: %v", r),
				}
			}
		}()

		data, err := work()
		out <- Response{
			Data: data,
			Err:  err,
		}
	}()

	select {
	case <-timer.C:
		return nil, fmt.Errorf("timeout")
	case result := <-out:
		return result.Data, result.Err
	}
}
