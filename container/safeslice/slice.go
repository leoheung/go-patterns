package safelist

import "sync"

type SafeSlice[T any] struct {
	mu    sync.RWMutex
	items []T
}

func NewSafeSlice[T any](capacity, length int) *SafeSlice[T] {
	if capacity < 0 {
		capacity = 0
	}

	if length < 0 {
		length = 0
	}

	if capacity < length {
		capacity = length
	}

	return &SafeSlice[T]{
		items: make([]T, length, capacity),
	}
}

// Peek 查看但不取出
func (l *SafeSlice[T]) Peek(index int) (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if index < 0 || index >= len(l.items) {
		var zero T
		return zero, false
	}
	return l.items[index], true
}

// PeekFirst 查看队首
func (l *SafeSlice[T]) PeekFirst() (T, bool) {
	return l.Peek(0)
}

// PeekLast 查看队尾
func (l *SafeSlice[T]) PeekLast() (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.items) == 0 {
		var zero T
		return zero, false
	}
	return l.items[len(l.items)-1], true
}

// Range 遍历（读锁，不阻塞其他读者）
func (l *SafeSlice[T]) Range(f func(index int, item T) bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for i, item := range l.items {
		if !f(i, item) {
			return
		}
	}
}

// Append 追加
func (l *SafeSlice[T]) Append(item T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.items = append(l.items, item)
}

// Remove 按条件删除
func (l *SafeSlice[T]) RemoveIf(predicate func(T) bool) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	removed := 0
	n := 0
	for _, item := range l.items {
		if predicate(item) {
			removed++
		} else {
			l.items[n] = item
			n++
		}
	}
	l.items = l.items[:n]
	return removed
}

// Len 长度
func (l *SafeSlice[T]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.items)
}
