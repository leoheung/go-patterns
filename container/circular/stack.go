package circular

import (
	"fmt"
	"sync"

	lctx "github.com/leoheung/go-patterns/container/context"
	"github.com/leoheung/go-patterns/container/safeslice"
	"github.com/leoheung/go-patterns/utils"
)

type CircularStack[T any] struct {
	subscribers   *safeslice.SafeSlice[chan struct{}]
	cs            []T
	ctx           *lctx.RenewableContext[any]
	mu            sync.RWMutex
	top_idx       int
	notify_buffer chan struct{}
}

func NewCircularStack[T any](size int, buffer_size int) *CircularStack[T] {
	if size <= 0 {
		size = 1
	}

	ret := &CircularStack[T]{
		subscribers:   safeslice.NewSafeSlice[chan struct{}](0, 0),
		ctx:           lctx.NewRenewableContext[any](nil, nil),
		cs:            make([]T, size),
		top_idx:       -1,
		mu:            sync.RWMutex{},
		notify_buffer: make(chan struct{}, 1),
	}

	go ret.notify()

	return ret
}

func (cs *CircularStack[T]) notify() {
	resume_ch := cs.ctx.SubscribeReactivation()

	for {
		_, paused := utils.TryDequeue(cs.ctx.Done())
		if paused {
			<-resume_ch
		}

		<-cs.notify_buffer

		cs.subscribers.Range(func(index int, item chan struct{}) bool {
			utils.TryEnqueue(item, struct{}{})
			return true
		})

	}
}

func (cs *CircularStack[T]) Push(data T) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.top_idx = (cs.top_idx + 1) % len(cs.cs)
	cs.cs[cs.top_idx] = data

	utils.TryEnqueue(cs.notify_buffer, struct{}{})
}

func (cs *CircularStack[T]) Peek() (T, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var zero T
	if cs.top_idx == -1 {
		return zero, fmt.Errorf("CircularStack is empty")
	}

	return cs.cs[cs.top_idx], nil
}

func (cs *CircularStack[T]) Pop() (T, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var zero T
	if cs.top_idx == -1 {
		return zero, fmt.Errorf("CircularStack is empty")
	}

	ret := cs.cs[cs.top_idx]
	cs.top_idx = (cs.top_idx - 1) % len(cs.cs)
	return ret, nil
}

func (v *CircularStack[T]) Pause() {
	v.ctx.Cancel()
}

func (v *CircularStack[T]) Resume() error {
	return v.ctx.Reactivate(nil)
}

func (v *CircularStack[T]) Subscribe(buffer int) (<-chan struct{}, func(), error) {
	if !v.ctx.IsAlive() {
		return nil, nil, fmt.Errorf("ValueProvider is paused")
	}

	if buffer < 0 {
		buffer = 0
	}
	ch := make(chan struct{}, buffer)
	v.subscribers.Append(ch)

	unsubscribe := func() {
		removed_count := v.subscribers.RemoveIf(func(c chan struct{}) bool { return c == ch })
		if removed_count == 1 {
			close(ch)
		}
	}

	return ch, unsubscribe, nil
}
