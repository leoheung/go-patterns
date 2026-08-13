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

// InOrderTraverse 中序遍历，按 cmp 升序对每个元素调用 fn。
func (t *Treap[T]) InOrderTraverse(fn func(T)) {
	bst.InOrderTraverse(t.root, fn)
}

// PreorderTraverse 前序遍历。
func (t *Treap[T]) PreorderTraverse(fn func(T)) {
	bst.PreorderTraverse(t.root, fn)
}

// PostorderTraverse 后序遍历。
func (t *Treap[T]) PostorderTraverse(fn func(T)) {
	bst.PostorderTraverse(t.root, fn)
}

// IsEmpty implements [bst.SelfBalancingBST].
func (t *Treap[T]) IsEmpty() bool {
	return t.root == nil
}

// IsLessThan implements [bst.SelfBalancingBST].
func (t *Treap[T]) IsLessThan() func(a T, b T) int {
	return t.cmp
}

// Max 返回树中最大元素；空树返回 (zero, false)。
func (t *Treap[T]) Max() (T, bool) {
	return bst.Max(t.root)
}

// Min 返回树中最小元素；空树返回 (zero, false)。
func (t *Treap[T]) Min() (T, bool) {
	return bst.Min(t.root)
}

// Predecessor 返回严格小于 item 的最大元素；不存在时返回 (zero, false)。
func (t *Treap[T]) Predecessor(item T) (T, bool) {
	return bst.Predecessor(t.root, item)
}

// Insert 按 BST 规则定位插入位置，用 new_treap_node 创建带 priority 的节点，
// 挂接后沿父链刷新 size，再经 reorganize_by_priority 上浮维护 treap 堆性质。
func (t *Treap[T]) Insert(item T) {
	n := new_treap_node(item, t.cmp)
	if t.root == nil {
		t.root = n
		return
	}
	insert_rec(t.root, n)
	reorganize_by_priority(n, &t.root)
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
	if t.root == nil {
		return 0
	}
	return t.root.GetSize()
}

// Successor 返回严格大于 item 的最小元素；不存在时返回 (zero, false)。
func (t *Treap[T]) Successor(item T) (T, bool) {
	return bst.Successor(t.root, item)
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
			bst.RefreshSizeUp(pp)
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
			bst.RefreshSizeUp(pp)
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
			bst.RefreshSizeUp(pp)
		}
		return true
	} else {
		// case 3: 2 children
		if pl.priority >= pr.priority {
			bst.RotateRight(p)
			if isRoot {
				*rootPtr = pl
			}
			return delete_rec(p, rootPtr)
		} else {
			bst.RotateLeft(p)
			if isRoot {
				*rootPtr = pr
			}
			return delete_rec(p, rootPtr)
		}
	}
}

func reorganize_by_priority[T any](p *treapNode[T], rootPtr **treapNode[T]) {
	if p == nil {
		return
	}
	pp := p.parentNode()
	if pp == nil {
		*rootPtr = p
		return
	}

	isLeftChild := bst.IsLeftChild(pp, p)

	if pp.priority < p.priority {
		if isLeftChild {
			bst.RotateRight(pp)
			reorganize_by_priority(p, rootPtr)
		} else {
			bst.RotateLeft(pp)
			reorganize_by_priority(p, rootPtr)
		}
	}
}

// insert_rec 把新节点 nn 挂到以 cur 为根的子树中；相等元素走左，允许重复共存。
func insert_rec[T any](cur, nn *treapNode[T]) {
	if cur.CompareFn()(cur.GetVal(), nn.GetVal()) >= 0 {
		if cur.leftNode() == nil {
			cur.SetLeft(nn)
			nn.SetParent(cur)
			bst.RefreshSizeUp(cur)
		} else {
			insert_rec(cur.leftNode(), nn)
		}
	} else {
		if cur.rightNode() == nil {
			cur.SetRight(nn)
			nn.SetParent(cur)
			bst.RefreshSizeUp(cur)
		} else {
			insert_rec(cur.rightNode(), nn)
		}
	}
}
