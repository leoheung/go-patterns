package circular

import (
	"fmt"
	"sync"

	lctx "github.com/leoheung/go-patterns/container/context"
	"github.com/leoheung/go-patterns/container/safeslice"
	"github.com/leoheung/go-patterns/utils"
)


// CircularStack is a fixed-capacity, last-in-first-out (LIFO/FILO) buffer for
// producer/consumer scenarios where a producer pushes data at a high frequency
// while a consumer reads the latest item more slowly.
//
// With a plain stack, a fast producer and a slow consumer cause unbounded
// growth. CircularStack bounds memory by fixing its size: once full, every new
// Push overwrites the oldest element and advances the top pointer, so the buffer
// always retains at most `size` of the most recently pushed items.
//
// Consumers subscribe to change notifications and retrieve the newest value via
// Peek or Pop. The stack can be paused and resumed: pausing stops notifications
// and terminates the background notification goroutine, while resuming restarts
// it.
type CircularStack[T any] struct {
	subscribers   *safeslice.SafeSlice[chan struct{}]
	cs            []T
	ctx           *lctx.RenewableContext[any]
	mu            sync.RWMutex
	top_idx       int
	count         int
	notify_buffer chan struct{}
}

func NewCircularStack[T any](size int) *CircularStack[T] {
	if size <= 0 {
		size = 1
	}

	ret := &CircularStack[T]{
		subscribers:   safeslice.NewSafeSlice[chan struct{}](0, 0),
		ctx:           lctx.NewRenewableContext[any](nil, nil),
		cs:            make([]T, size),
		top_idx:       -1,
		count:         0,
		mu:            sync.RWMutex{},
		notify_buffer: make(chan struct{}, 1),
	}

	go ret.notify()

	return ret
}

func (cs *CircularStack[T]) notify() {
	for {
		select {
		case <-cs.ctx.Done():
			return

		case <-cs.notify_buffer:
			cs.subscribers.Range(func(index int, item chan struct{}) bool {
				utils.TryEnqueue(item, struct{}{})
				return true
			})
		}
	}
}

func (cs *CircularStack[T]) Push(data T) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.ctx.IsAlive() {
		return fmt.Errorf("CircularStack is paused")
	}

	cs.top_idx = utils.ModEuclid((cs.top_idx + 1), len(cs.cs))
	cs.cs[cs.top_idx] = data
	if cs.count < len(cs.cs) {
		cs.count++
	}

	utils.TryEnqueue(cs.notify_buffer, struct{}{})
	return nil
}

func (cs *CircularStack[T]) Peek() (T, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var zero T
	if cs.count == 0 {
		return zero, fmt.Errorf("CircularStack is empty")
	}

	return cs.cs[cs.top_idx], nil
}

func (cs *CircularStack[T]) Pop() (T, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var zero T
	if !cs.ctx.IsAlive() {
		return zero, fmt.Errorf("CircularStack is paused")
	}

	if cs.count == 0 {
		return zero, fmt.Errorf("CircularStack is empty")
	}

	ret := cs.cs[cs.top_idx]
	cs.top_idx = utils.ModEuclid((cs.top_idx - 1), len(cs.cs))
	cs.count--
	return ret, nil
}

func (v *CircularStack[T]) Pause() {
	v.ctx.Cancel()
}

func (v *CircularStack[T]) Resume() error {
	err := v.ctx.Reactivate(nil)
	if err != nil {
		return err
	}

	go v.notify()
	return nil
}

func (v *CircularStack[T]) Subscribe(buffer int) (<-chan struct{}, func(), error) {
	if !v.ctx.IsAlive() {
		return nil, nil, fmt.Errorf("CircularStack is paused")
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
