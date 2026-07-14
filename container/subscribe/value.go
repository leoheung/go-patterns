package subscribe

import (
	"fmt"
	"sync"

	lctx "github.com/leoheung/go-patterns/container/context"
	"github.com/leoheung/go-patterns/container/safeslice"
	"github.com/leoheung/go-patterns/utils"
)

type ValueProvider[T any] struct {
	value       T
	subscribers *safeslice.SafeSlice[chan T]
	valueBuffer chan T
	mu          sync.RWMutex
	ctx         *lctx.RenewableContext[any]
}

func NewValueProvider[T any](val T, buffer_size int) *ValueProvider[T] {
	if buffer_size < 0 {
		buffer_size = 0
	}

	ret := &ValueProvider[T]{
		value:       val,
		subscribers: safeslice.NewSafeSlice[chan T](0, 0),
		valueBuffer: make(chan T, buffer_size),
		mu:          sync.RWMutex{},
		ctx:         lctx.NewRenewableContext[any](nil, nil),
	}

	go ret.notify_worker()

	return ret
}

func (v *ValueProvider[T]) Get() T {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.value
}

func (v *ValueProvider[T]) Set(val T) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.ctx.IsAlive() {
		return fmt.Errorf("ValueProvider is paused")
	}

	v.value = val
	v.valueBuffer <- val
	return nil
}

func (v *ValueProvider[T]) Subscribe(buffer int) (<-chan T, func(), error) {
	if !v.ctx.IsAlive() {
		return nil, nil, fmt.Errorf("ValueProvider is paused")
	}

	if buffer < 0 {
		buffer = 0
	}
	ch := make(chan T, buffer)
	v.subscribers.Append(ch)

	unsubscribe := func() {
		removed_count := v.subscribers.RemoveIf(func(c chan T) bool { return c == ch })
		if removed_count == 1 {
			close(ch)
		}
	}

	return ch, unsubscribe, nil
}

func (v *ValueProvider[T]) Pause() {
	v.ctx.Cancel()
}

func (v *ValueProvider[T]) Resume() error {
	return v.ctx.Reactivate(nil)
}

func (v *ValueProvider[T]) notify_worker() {
	resume_ch := v.ctx.SubscribeReactivation()

	for {
		_, paused := utils.TryDequeue(v.ctx.Done())
		if paused {
			<-resume_ch
		}

		new_val := <-v.valueBuffer

		v.subscribers.Range(func(index int, item chan T) bool {
			utils.TryEnqueue(item, new_val)
			return true
		})

	}
}
