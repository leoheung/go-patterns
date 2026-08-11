package treap

import (
	"math/rand/v2"

	"github.com/leoheung/go-patterns/container/tree/bst"
)

type treapNode[T any] struct {
	*bst.Node[T] // 嵌入指针,所有 BSTNodeInterface method 会被 method promotion 提升到 *treapNode[T]
	priority     int
}

type Treap[T any] struct {
	root *treapNode[T]
	cmp  func(a, b T) int
}

var _ bst.SelfBalancingBST[int] = new(Treap[int])

func new_treap_node[T any](val T, cmp func(a, b T) int) *treapNode[T] {
	return &treapNode[T]{
		bst.NewNode(val, cmp),
		rand.Int(),
	}
}

func NewTreap[T any](cmp func(a, b T) int) *Treap[T] {
	return &Treap[T]{
		root: nil,
		cmp:  cmp,
	}
}

// Clear implements [bst.SelfBalancingBST].
func (t *Treap[T]) Clear() {
	t.root = nil
}

// Delete implements [bst.SelfBalancingBST].
func (t *Treap[T]) Delete(item T) bool {
	panic("unimplemented")
}

// Get implements [bst.SelfBalancingBST].
func (t *Treap[T]) Get(item T) (T, bool) {
	return bst.Get(t.root, item)
}

// InOrderTraverse implements [bst.SelfBalancingBST].
func (t *Treap[T]) InOrderTraverse(fn func(T)) {
	panic("unimplemented")
}

// IsEmpty implements [bst.SelfBalancingBST].
func (t *Treap[T]) IsEmpty() bool {
	panic("unimplemented")
}

// IsLessThan implements [bst.SelfBalancingBST].
func (t *Treap[T]) IsLessThan() func(a T, b T) int {
	panic("unimplemented")
}

// Max implements [bst.SelfBalancingBST].
func (t *Treap[T]) Max() (T, bool) {
	panic("unimplemented")
}

// Min implements [bst.SelfBalancingBST].
func (t *Treap[T]) Min() (T, bool) {
	panic("unimplemented")
}

// Predecessor implements [bst.SelfBalancingBST].
func (t *Treap[T]) Predecessor(item T) (T, bool) {
	panic("unimplemented")
}

// Put implements [bst.SelfBalancingBST].
func (t *Treap[T]) Put(item T) (inserted bool) {
	panic("unimplemented")
}

// RangeVisit implements [bst.SelfBalancingBST].
func (t *Treap[T]) RangeVisit(low T, high T, callback func(T)) {
	panic("unimplemented")
}

// Rank implements [bst.SelfBalancingBST].
func (t *Treap[T]) Rank(item T) int {
	panic("unimplemented")
}

// Select implements [bst.SelfBalancingBST].
func (t *Treap[T]) Select(rank int) (T, bool) {
	panic("unimplemented")
}

// Size implements [bst.SelfBalancingBST].
func (t *Treap[T]) Size() int {
	panic("unimplemented")
}

// Successor implements [bst.SelfBalancingBST].
func (t *Treap[T]) Successor(item T) (T, bool) {
	panic("unimplemented")
}
