package pq

import (
	"errors"

	"github.com/leoheung/go-patterns/container/tree/heap"
)

// PriorityQueue 基于二叉堆的无界优先队列，支持泛型。
type PriorityQueue[T any] struct {
	h *heap.BinaryHeap[T]
}

// NewPriorityQueue 创建优先队列。better(a,b) 返回 true 表示 a 应排在 b 前面。
func NewPriorityQueue[T any](better func(a, b T) bool) (*PriorityQueue[T], error) {
	if better == nil {
		return nil, errors.New("better function cannot be nil")
	}
	return &PriorityQueue[T]{h: heap.NewBinaryHeap(better)}, nil
}

// Len 返回队列当前长度，O(1)。
func (pq *PriorityQueue[T]) Len() int { return pq.h.Len() }

// Enqueue 入队，O(log n)。
func (pq *PriorityQueue[T]) Enqueue(item T) error {
	pq.h.Push(item)
	return nil
}

// Dequeue 移除并返回优先级最高的元素，O(log n)。空队返回 error。
func (pq *PriorityQueue[T]) Dequeue() (T, error) {
	var zero T
	v, ok := pq.h.Pop()
	if !ok {
		return zero, errors.New("queue is empty")
	}
	return v, nil
}

// Peek 查看优先级最高的元素，O(1)。空队返回 error。
func (pq *PriorityQueue[T]) Peek() (T, error) {
	var zero T
	v, ok := pq.h.Peek()
	if !ok {
		return zero, errors.New("queue is empty")
	}
	return v, nil
}

// Data 返回队列中全部元素的副本（无序，仅供遍历/查看）。
func (pq *PriorityQueue[T]) Data() []T {
	return pq.h.Slice()
}