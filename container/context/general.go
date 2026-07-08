package context

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

var _ context.Context = new(RenewableContext[struct{}])

type RenewableContext[T any] struct {
	baseCtx     context.Context
	cancel      context.CancelFunc
	data        T
	mu          sync.RWMutex
	subscribers []chan struct{}
}

func NewRenewableContext[T any](timeout *time.Duration, data T) *RenewableContext[T] {
	ctx := context.Background()
	var cancel context.CancelFunc

	if timeout != nil {
		ctx, cancel = context.WithTimeout(ctx, *timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}

	gc := &RenewableContext[T]{
		baseCtx:     ctx,
		cancel:      cancel,
		data:        data,
		mu:          sync.RWMutex{},
		subscribers: make([]chan struct{}, 0),
	}
	return gc
}

func (gt *RenewableContext[T]) Cancel() {
	gt.mu.Lock()
	defer gt.mu.Unlock()
	gt.cancel()
}

func (gt *RenewableContext[T]) Done() <-chan struct{} {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.baseCtx.Done()
}

func (gt *RenewableContext[T]) IsAlive() bool {
	gt.mu.RLock()
	defer gt.mu.RUnlock()

	select {
	case <-gt.baseCtx.Done():
		return false
	default:
		return true
	}
}

func (gt *RenewableContext[T]) Reactivate(timeout *time.Duration) error {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	select {
	case <-gt.baseCtx.Done():
		ctx := context.Background()
		var cancel context.CancelFunc

		if timeout != nil {
			ctx, cancel = context.WithTimeout(ctx, *timeout)
		} else {
			ctx, cancel = context.WithCancel(ctx)
		}

		gt.baseCtx = ctx
		gt.cancel = cancel

		for i := 0; i < len(gt.subscribers); i++ {
			utils.TryEnqueue(gt.subscribers[i], struct{}{})
		}

		return nil
	default:
		return fmt.Errorf("failed to reactivate: the current context is still alive")
	}
}

func (gt *RenewableContext[T]) OnReactivate() <-chan struct{} {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	ret := make(chan struct{})
	gt.subscribers = append(gt.subscribers, ret)
	return ret
}

func (gt *RenewableContext[T]) GetData() T {
	gt.mu.RLock()
	defer gt.mu.RUnlock()

	return gt.data
}

func (gt *RenewableContext[T]) SetData(newData T) {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	gt.data = newData
}

func (gt *RenewableContext[T]) MergeContext(otherCtx context.Context) {
	gt.mu.Lock()
	defer gt.mu.Unlock()

	newCtx, newCancel := utils.MergeContexts(gt.baseCtx, otherCtx)

	gt.baseCtx, gt.cancel = newCtx, newCancel
}

// Deadline implements [context.Context].
func (gt *RenewableContext[T]) Deadline() (deadline time.Time, ok bool) {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.baseCtx.Deadline()
}

// Err implements [context.Context].
func (gt *RenewableContext[T]) Err() error {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.baseCtx.Err()
}

// Value implements [context.Context].
func (gt *RenewableContext[T]) Value(key any) any {
	gt.mu.RLock()
	defer gt.mu.RUnlock()
	return gt.baseCtx.Value(key)
}
