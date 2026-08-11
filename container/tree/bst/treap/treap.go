package treap

import (
	"github.com/leoheung/go-patterns/container/tree/bst"
)

type Treap[T any] struct {
	root *treapNode[T]
	cmp  func(a, b T) int
}

var _ bst.SelfBalancingBST[int] = new(Treap[int])

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
	ptr, ok := bst.Get(t.root, item)
	if !ok {
		return false
	}
	return delete_rec(ptr.(*treapNode[T]), &t.root)
}

// Get implements [bst.SelfBalancingBST].
func (t *Treap[T]) Get(item T) (T, bool) {
	var zero T
	ptr, ok := bst.Get(t.root, item)

	if ok {
		return ptr.GetVal(), true
	} else {
		return zero, false
	}
}

// InOrderTraverse implements [bst.SelfBalancingBST].
func (t *Treap[T]) InOrderTraverse(fn func(T)) {
	panic("unimplemented")
}

// IsEmpty implements [bst.SelfBalancingBST].
func (t *Treap[T]) IsEmpty() bool {
	return t.root == nil
}

// IsLessThan implements [bst.SelfBalancingBST].
func (t *Treap[T]) IsLessThan() func(a T, b T) int {
	return t.cmp
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

// refreshSizeUp 自 n 起沿父链向上逐个调用 bst.UpdateSize,回填子树 size。
func refreshSizeUp[T any](n *treapNode[T]) {
	for n != nil {
		bst.UpdateSize(n)
		n = n.parentNode()
	}
}

func delete_rec[T any](p *treapNode[T], rootPtr **treapNode[T]) bool {
	if p == nil {
		return false
	}

	pp := p.parentNode()
	pl := p.leftNode()
	pr := p.rightNode()
	isRoot := pp == nil

	// case 1: leaf
	if pl == nil && pr == nil {
		if isRoot {
			*rootPtr = nil
		} else {
			if bst.IsLeftChild(pp, p) {
				pp.SetLeft(nil)
			} else {
				pp.SetRight(nil)
			}
			refreshSizeUp(pp)
		}
		return true
	}

	// case 2: only 1 child
	if pl != nil && pr == nil {
		if isRoot {
			pl.SetParent(nil)
			*rootPtr = pl
		} else {
			if bst.IsLeftChild(pp, p) {
				pp.SetLeft(pl)
			} else {
				pp.SetRight(pl)
			}
			pl.SetParent(pp)
			refreshSizeUp(pp)
		}
		return true
	} else if pl == nil && pr != nil {
		if isRoot {
			pr.SetParent(nil)
			*rootPtr = pr
		} else {
			if bst.IsLeftChild(pp, p) {
				pp.SetLeft(pr)
			} else {
				pp.SetRight(pr)
			}
			pr.SetParent(pp)
			refreshSizeUp(pp)
		}
		return true
	} else {
		// case 3: 2 children
		if pl.priority >= pr.priority {
			bst.RotateRight(p)
			if isRoot {
				*rootPtr = pl
			}
			return delete_rec(p,rootPtr)
		} else {
			bst.RotateLeft(p)
			if isRoot {
				*rootPtr = pr
			}
			return delete_rec(p,rootPtr)
		}
	}
}
