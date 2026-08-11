package treap

import (
	"math/rand/v2"

	"github.com/leoheung/go-patterns/container/tree/bst"
)

type treapNode[T any] struct {
	*bst.Node[T] // 嵌入指针,所有 BSTNodeInterface method 会被 method promotion 提升到 *treapNode[T]
	priority     int
}

func new_treap_node[T any](val T, cmp func(a, b T) int) *treapNode[T] {
	return &treapNode[T]{
		bst.NewNode(val, cmp),
		rand.Int(),
	}
}

func (n *treapNode[T]) leftNode() *treapNode[T]   { return n.GetLeft().(*treapNode[T]) }
func (n *treapNode[T]) rightNode() *treapNode[T]  { return n.GetRight().(*treapNode[T]) }
func (n *treapNode[T]) parentNode() *treapNode[T] { return n.GetParent().(*treapNode[T]) }
