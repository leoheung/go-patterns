package pool

import (
	"fmt"

	"github.com/leoheung/go-patterns/parallel/semaphore"
	"github.com/leoheung/go-patterns/utils"
)

type WorkerPool struct {
	sem *semaphore.SemaphoreByCond
}

func NewWorkerPool(numWorkers int) *WorkerPool {
	return &WorkerPool{
		sem: semaphore.NewSemaphoreByCond(numWorkers),
	}
}

func (wp *WorkerPool) Submit(task func()) {
	if task == nil {
		return
	}
	wp.sem.Acquire()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.LogMessage(fmt.Sprintf("%v", r))
			}
			wp.sem.Release()
		}()
		task()
	}()
}

func (wp *WorkerPool) TrySubmit(task func()) bool {
	if task == nil {
		return false
	}
	if !wp.sem.TryAcquire() {
		return false
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.LogMessage(fmt.Sprintf("%v", r))
			}
			wp.sem.Release()
		}()
		task()
	}()
	return true
}
