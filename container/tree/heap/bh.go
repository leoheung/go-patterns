package heap

// BinaryHeap 基于数组实现的二叉堆（默认最小堆：根最小）。
// 用 prioritize(a,b) 定义优先级：prioritize(a,b) 为 true 表示 a 比 b 优先。
type BinaryHeap[T any] struct {
	data       []T
	prioritize func(a, b T) bool
}

// NewBinaryHeap 创建一个二叉堆。prioritize 定义优先级，nil 则 panic。
func NewBinaryHeap[T any](prioritize func(a, b T) bool) *BinaryHeap[T] {
	if prioritize == nil {
		panic("BinaryHeap: prioritize function cannot be nil")
	}
	return &BinaryHeap[T]{prioritize: prioritize}
}

// NewBinaryHeapFrom 用已有切片建堆（O(n) heapify），共享底层切片。
func NewBinaryHeapFrom[T any](data []T, prioritize func(a, b T) bool) *BinaryHeap[T] {
	if prioritize == nil {
		panic("BinaryHeap: prioritize function cannot be nil")
	}
	h := &BinaryHeap[T]{data: data, prioritize: prioritize}
	// heapify
	n := len(h.data)
	for i := n/2 - 1; i >= 0; i-- {
		h.down(i, n)
	}
	return h
}

// Len 返回堆中元素个数，O(1)。
func (h *BinaryHeap[T]) Len() int { return len(h.data) }

// Push 入堆，O(log n)。
func (h *BinaryHeap[T]) Push(x T) {
	h.data = append(h.data, x)
	h.up(len(h.data) - 1)
}

// Pop 取出堆顶（优先级最高）并移除，O(log n)。空堆返回 (zero, false)。
func (h *BinaryHeap[T]) Pop() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}
	n := len(h.data)
	h.swap(0, n-1)
	item := h.data[n-1]
	h.data = h.data[:n-1]
	h.down(0, len(h.data))
	return item, true
}

// Peek 查看堆顶但不移除，O(1)。空堆返回 (zero, false)。
func (h *BinaryHeap[T]) Peek() (T, bool) {
	if len(h.data) == 0 {
		var zero T
		return zero, false
	}
	return h.data[0], true
}

// up 上浮下标 j 的元素，维持堆性质。
func (h *BinaryHeap[T]) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !h.prioritize(h.data[j], h.data[i]) {
			break
		}
		h.swap(i, j)
		j = i
	}
}

// down 下沉下标 i0 的元素到位置 < n，维持堆性质。返回是否发生了下沉。
func (h *BinaryHeap[T]) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 {
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && h.prioritize(h.data[j2], h.data[j1]) {
			j = j2 // right child 更优先
		}
		if !h.prioritize(h.data[j], h.data[i]) {
			break
		}
		h.swap(i, j)
		i = j
	}
	return i > i0
}

func (h *BinaryHeap[T]) swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}
