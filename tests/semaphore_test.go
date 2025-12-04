package tests

import (
	"testing"
	"time"
	"github.com/leoheung/go-patterns/parallel/semaphore"
)

func TestSemaphore(t *testing.T) {
	sem := semaphore.NewSemaphore(2)
	sem.Acquire()
	sem.Acquire()
	done := make(chan struct{})
	go func() {
		sem.Release()
		sem.Release()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Semaphore release timeout")
	}
}
func TestSemaphoreByCond(t *testing.T) {
	sem := semaphore.NewSemaphoreByCond(2)
	sem.Acquire()
	sem.Acquire()
	done := make(chan struct{})
	go func() {
		sem.Release()
		sem.Release()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Semaphore release timeout")
	}
}
