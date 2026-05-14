package pool

import (
	"fmt"
	"sync"

	"github.com/leoheung/go-patterns/parallel/token"
	"github.com/leoheung/go-patterns/utils"
)

type taskWithOnError struct {
	task    Task
	onError OnError
}

type AsyncPool struct {
	taskBuffer chan *taskWithOnError
	submitLock sync.RWMutex
	tokens     *token.StaticTokens
	closed     *token.BoolToken
}

// NewAsyncPool creates a new AsyncPool.
// taskBufferCapacity is the capacity of the task buffer.
// numWorkers is the number of workers in the pool.
func NewAsyncPool(taskBufferCapacity int, numWorkers int) (*AsyncPool, error) {
	if numWorkers <= 0 {
		return nil, fmt.Errorf("numWorkers <= 0 : %d", numWorkers)
	}

	if taskBufferCapacity <= 0 {
		return nil, fmt.Errorf("taskBufferCapacity <= 0 : %d", taskBufferCapacity)
	}

	tokens, _ := token.NewStaticTokens(numWorkers)

	return &AsyncPool{
		taskBuffer: make(chan *taskWithOnError, taskBufferCapacity),
		tokens:     tokens,
		closed:     token.NewBoolToken(false),
	}, nil
}

func (ap *AsyncPool) Shutdown() {
	ap.closed.Set(true)
}

// AsyncSubmit submits a task asynchronously.
//
// If the task buffer is full, it will return an error.
// If the pool is closed, it will return an error.
//
// Task type: func() error
//
// OnError type: func(error)
func (ap *AsyncPool) AsyncSubmit(task Task, onError OnError) error {
	ap.submitLock.RLock()
	defer ap.submitLock.RUnlock()

	if ap.closed.Get() {
		return fmt.Errorf("AsyncPool is already closed")
	}

	if ap.tokens.GrantNextToken() {
		go ap.work(task, onError)
		return nil
	} else {
		bufferFull := !utils.TryEnqueue(ap.taskBuffer, &taskWithOnError{
			task:    task,
			onError: onError,
		})

		if bufferFull {
			return fmt.Errorf("failed to async submit: task buffer is full, consider increase the buffer capacity or retry later")
		} else {
			return nil
		}
	}
}

func (ap *AsyncPool) work(task Task, onError OnError) {
	defer func() {
		if r := recover(); r != nil {
			utils.LogMessage(fmt.Sprintf("panic when executing task: %v", r))
		}

		if ap.workerCanQuit() {
			return
		}

		newTask := <-ap.taskBuffer
		ap.work(newTask.task, newTask.onError)
	}()

	if task != nil {
		err := task()
		if err != nil && onError != nil {
			onError(err)
		}
	}
}

func (ap *AsyncPool) workerCanQuit() bool {
	ap.submitLock.Lock()
	defer ap.submitLock.Unlock()
	return ap.closed.Get() && len(ap.taskBuffer) == 0
}
