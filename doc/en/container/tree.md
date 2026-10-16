# Tree

Overview of the tree-based containers: self-balancing binary search trees (`bst` + `treap`), a dependency-tree utility (`dbtree`), and a binary heap (`heap`). Each sub-module lives in its own package with its own import path.

| Sub-module | Package | Description |
|---|---|---|
| bst / treap | `container/tree/bst`, `container/tree/bst/treap` | Self-balancing BST framework + Treap implementation |
| dbtree | `container/tree/dbtree` | Dependency-tree ordering (topological order) |
| heap | `container/tree/heap` | Generic binary heap (priority queue) |

## 1. Self-Balancing BST (bst / treap)

An ordered-set framework (`bst`) defining the `SelfBalancingBST[T]` contract and shared node/rotation utilities, with a ready-to-use **Treap** implementation (`treap`).

### Installation

```go
import (
    "github.com/leoheung/go-patterns/container/tree/bst"
    "github.com/leoheung/go-patterns/container/tree/bst/treap"
)
```

### SelfBalancingBST[T] Interface

The contract every balancing strategy (treap / splay / avl / ...) implements. The comparison function `IsLessThan()` is a three-way comparator: `-1` a < b, `0` a == b, `1` a > b. Equal elements may coexist (duplicates allowed, inserted to the left).

```go
type SelfBalancingBST[T any] interface {
    Insert(item T)                        // insert; duplicates coexist (never overwritten)
    Update(item T, callback func(item T)) // locate equal node, mutate it, re-place if the sort key changed
    Get(item T) (T, bool)                 // lookup by sample item
    Delete(item T) bool                   // delete equal node; true if removed
    Size() int                            // number of elements
    IsEmpty() bool                        // is the tree empty
    Clear()                               // remove all nodes
    InOrderTraverse(fn func(T))           // ascending order
    PreorderTraverse(fn func(T))          // parent → left → right
    PostorderTraverse(fn func(T))         // left → right → parent
    Min() (T, bool)                       // smallest element
    Max() (T, bool)                       // largest element
    Predecessor(item T) (T, bool)         // largest element strictly < item
    Successor(item T) (T, bool)           // smallest element strictly > item
    RangeVisit(low, high T, callback func(T)) // ascending visit over closed range [low, high]
    Rank(item T) int                      // count of elements strictly < item (0-based)
    Select(rank int) (T, bool)            // element at rank (0 = smallest)
    IsLessThan() func(a, b T) int         // the bound three-way comparator
    DrawTree(out io.Writer)               // ASCII tree dump for debugging
}
```

### Treap

A treap = BST + heap: nodes are ordered by `cmp` (BST invariant) and by a random `priority` (heap invariant), keeping the tree balanced with high probability. Implements `SelfBalancingBST[T]`.

```go
// Create a treap with a three-way comparator (-1 / 0 / 1).
func NewTreap[T any](cmp func(a, b T) int) *Treap[T]
```

`Treap[T]` exposes the full `SelfBalancingBST[T]` API (`Insert` / `Update` / `Get` / `Delete` / `Rank` / `Select` / `RangeVisit` / traversals / `DrawTree` / ...).

### Shared Utilities (package bst)

The `bst` package also provides standalone helpers for custom node-based trees:

```go
type CMPFN[T any] = func(a, b T) int

// Node[T]: default BSTNodeInterface implementation, embed *Node[T] in custom nodes.
func NewNode[T any](val T, cmp func(a, b T) int) *Node[T]

// Rotations (return the new subtree root) and size maintenance
func RotateLeft[T any](p BSTNodeInterface[T]) BSTNodeInterface[T]
func RotateRight[T any](p BSTNodeInterface[T]) BSTNodeInterface[T]
func UpdateSize[T any](n BSTNodeInterface[T])
func RefreshSizeUp[T any](n BSTNodeInterface[T])
func IsLeftChild[T any](parent, child BSTNodeInterface[T]) bool

// Generic algorithms over any BST node
func Get[T any](p BSTNodeInterface[T], item T) (BSTNodeInterface[T], bool)
func Insert[T any](p *Node[T], pp *Node[T], left bool, item T, cmp CMPFN[T]) *Node[T]
func Min[T any](p BSTNodeInterface[T]) (T, bool)
func Max[T any](p BSTNodeInterface[T]) (T, bool)
func Predecessor[T any](p BSTNodeInterface[T], item T) (T, bool)
func Successor[T any](p BSTNodeInterface[T], item T) (T, bool)
func Rank[T any](p BSTNodeInterface[T], item T) int
func Select[T any](p BSTNodeInterface[T], rank int) (T, bool)
func RangeVisit[T any](p BSTNodeInterface[T], low, high T, fn func(T))
func InOrderTraverse[T any](p BSTNodeInterface[T], fn func(T))
func PreorderTraverse[T any](p BSTNodeInterface[T], fn func(T))
func PostorderTraverse[T any](p BSTNodeInterface[T], fn func(T))
func IsOrderedNode[T any](n BSTNodeInterface[T]) bool // check BST ordering invariant
func DrawTree[T any](p BSTNodeInterface[T], out io.Writer)
```

## 2. Dependency Tree (dbtree)

Computes a build/execution order for tables with dependencies (topological order): every table appears **after** the tables it depends on.

### Installation

```go
import "github.com/leoheung/go-patterns/container/tree/dbtree"
```

### API

```go
// BuildDependencyOrder returns tables in dependency order.
//   - tables: the tables to order
//   - tablesDependencies: table → tables it depends on
//   - returns nil if tables is empty
func BuildDependencyOrder(tables []string, tablesDependencies map[string][]string) []string
```

## 3. Binary Heap (heap)

A generic array-based binary heap (min-heap by default: the root is the "most prioritized" element). Priority is defined by a `prioritize(a, b T) bool` function: `true` means `a` beats `b`.

### Installation

```go
import "github.com/leoheung/go-patterns/container/tree/heap"
```

### API

```go
// Create a binary heap. prioritize defines priority; nil panics.
func NewBinaryHeap[T any](prioritize func(a, b T) bool) *BinaryHeap[T]

// Build a heap from an existing slice (O(n) heapify). Shares the backing slice.
func NewBinaryHeapFrom[T any](data []T, prioritize func(a, b T) bool) *BinaryHeap[T]

func (h *BinaryHeap[T]) Len() int          // number of elements, O(1)
func (h *BinaryHeap[T]) Push(x T)          // push, O(log n)
func (h *BinaryHeap[T]) Pop() (T, bool)    // remove and return the top; (zero, false) if empty, O(log n)
func (h *BinaryHeap[T]) Peek() (T, bool)   // return the top without removing; (zero, false) if empty, O(1)
func (h *BinaryHeap[T]) Slice() []T        // copy of the backing elements (unordered), for inspection only
```

## Complete Example

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/container/tree/bst/treap"
	"github.com/leoheung/go-patterns/container/tree/dbtree"
	"github.com/leoheung/go-patterns/container/tree/heap"
)

func main() {
	// ── Treap: ordered multiset with rank/select ──
	// three-way comparator: -1 (a<b) / 0 (equal) / 1 (a>b)
	t := treap.NewTreap(func(a, b int) int {
		switch {
		case a < b:
			return -1
		case a > b:
			return 1
		default:
			return 0
		}
	})

	for _, v := range []int{5, 3, 9, 1, 7} {
		t.Insert(v)
	}
	fmt.Println("size:", t.Size()) // 5

	if min, ok := t.Min(); ok {
		fmt.Println("min:", min) // 1
	}
	if v, ok := t.Select(2); ok {
		fmt.Println("rank 2:", v) // 5 (0-based: 1,3,5,7,9)
	}
	fmt.Println("rank of 7:", t.Rank(7)) // 3

	t.RangeVisit(3, 7, func(v int) { fmt.Print(v, " ") }) // 3 5 7
	fmt.Println()

	// ── dbtree: dependency order (dependencies first) ──
	order := dbtree.BuildDependencyOrder(
		[]string{"orders", "users"},
		map[string][]string{
			"orders": {"users"},
			"users":  {},
		},
	)
	fmt.Println("build order:", order) // [users orders]

	// ── heap: generic priority queue ──
	h := heap.NewBinaryHeap(func(a, b int) bool { return a < b }) // min-heap
	for _, v := range []int{5, 3, 9, 1, 7} {
		h.Push(v)
	}
	fmt.Println("heap len:", h.Len()) // 5
	if top, ok := h.Peek(); ok {
		fmt.Println("top:", top) // 1
	}
	for {
		v, ok := h.Pop()
		if !ok {
			break
		}
		fmt.Print(v, " ") // 1 3 5 7 9
	}
	fmt.Println()
}
```

## Features

- **Ordered Containers with Order Statistics**: `Rank` / `Select` / `Predecessor` / `Successor` / `RangeVisit` on the treap — ideal for leaderboards, schedulers and range scans.
- **Multiset Semantics**: equal elements coexist (never overwritten) across the BST framework.
- **Pluggable Balancing**: the `SelfBalancingBST[T]` contract is shared; treap ships as the default implementation, with shared rotation/size utilities for building more (splay / AVL / ...).
- **Topological Ordering**: `dbtree.BuildDependencyOrder` for dependency-driven build/execution orders.
- **Generic Binary Heap**: O(log n) push/pop with any priority relation; O(n) heapify from a slice.
- **Debugging Aid**: `DrawTree` renders the tree structure as ASCII art.
