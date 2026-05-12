package pool

import (
	"fmt"

	"github.com/leoheung/go-patterns/parallel/semaphore"
	"github.com/leoheung/go-patterns/utils"
)

type WorkerPool struct {
	sem *semaphore.SemaphoreByCond
}

type Task func() error
type OnError func(error)

func NewWorkerPool(numWorkers int) *WorkerPool {
	if numWorkers <= 0 {
		panic("numWorkers must be > 0")
	}

	return &WorkerPool{
		sem: semaphore.NewSemaphoreByCond(numWorkers),
	}
}

func (wp *WorkerPool) Submit(task Task, onError OnError) {
	if task == nil {
		return
	}
	wp.sem.Acquire()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.LogMessage(fmt.Sprintf("panic when executing task: %v", r))
			}
			wp.sem.Release()
		}()
		err := task()
		if err != nil {
			utils.LogMessage(fmt.Sprintf("error when executing task: %s", err.Error()))
			if onError != nil {
				onError(err)
			}
		}
	}()
}

func (wp *WorkerPool) TrySubmit(task Task, onError OnError) bool {
	if task == nil {
		return false
	}
	if !wp.sem.TryAcquire() {
		return false
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				utils.LogMessage(fmt.Sprintf("panic when executing task: %v", r))
			}
			wp.sem.Release()
		}()

		err := task()
		if err != nil {
			utils.LogMessage(fmt.Sprintf("error when executing task: %s", err.Error()))
			if onError != nil {
				onError(err)
			}
		}
	}()
	return true
}
