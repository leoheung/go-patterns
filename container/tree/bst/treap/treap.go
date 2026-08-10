package treap

import "github.com/leoheung/go-patterns/container/tree/bst"

type treapNode[T any] struct {
	*bst.Node[T] // 嵌入指针,所有 BSTNodeInterface method 会被 method promotion 提升到 *treapNode[T]
	priority     int
}
