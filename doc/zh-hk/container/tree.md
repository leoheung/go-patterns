# Tree

樹形容器概覽：自我平衡二叉搜尋樹（`bst` + `treap`）、依賴樹工具（`dbtree`）同二叉堆（`heap`）。每個子模組都有自己嘅套件同 import 路徑。

| 子模組 | 套件 | 說明 |
|---|---|---|
| bst / treap | `container/tree/bst`, `container/tree/bst/treap` | 自我平衡 BST 框架 + Treap 實作 |
| dbtree | `container/tree/dbtree` | 依賴樹排序（拓撲排序） |
| heap | `container/tree/heap` | 泛型二叉堆（優先隊列） |

## 1. 自我平衡 BST（bst / treap）

一個有序集合框架（`bst`），定義咗 `SelfBalancingBST[T]` 契約同共享嘅節點／旋轉工具，並提供一個開箱即用嘅 **Treap** 實作（`treap`）。

### 安裝

```go
import (
    "github.com/leoheung/go-patterns/container/tree/bst"
    "github.com/leoheung/go-patterns/container/tree/bst/treap"
)
```

### SelfBalancingBST[T] 介面

每個平衡策略（treap / splay / avl / ...）都要實作嘅契約。比較函數 `IsLessThan()` 係三態比較器：`-1` a < b、`0` a == b、`1` a > b。相等元素可以共存（允許重複，插入時行左邊）。

```go
type SelfBalancingBST[T any] interface {
    Insert(item T)                        // 插入；重複元素共存（唔覆蓋）
    Update(item T, callback func(item T)) // 定位相等節點、修改內容，排序鍵變咗就重新擺位
    Get(item T) (T, bool)                 // 用樣板 item 搜尋
    Delete(item T) bool                   // 刪除相等節點；成功移除返回 true
    Size() int                            // 元素總數
    IsEmpty() bool                        // 樹係咪空
    Clear()                               // 清空所有節點
    InOrderTraverse(fn func(T))           // 升序遍歷
    PreorderTraverse(fn func(T))          // 父 → 左 → 右
    PostorderTraverse(fn func(T))         // 左 → 右 → 父
    Min() (T, bool)                       // 最小元素
    Max() (T, bool)                       // 最大元素
    Predecessor(item T) (T, bool)         // 嚴格細過 item 嘅最大元素
    Successor(item T) (T, bool)           // 嚴格大過 item 嘅最小元素
    RangeVisit(low, high T, callback func(T)) // 閉區間 [low, high] 升序遍歷
    Rank(item T) int                      // 嚴格細過 item 嘅元素個數（0-based）
    Select(rank int) (T, bool)            // 第 rank 小嘅元素（0 = 最小）
    IsLessThan() func(a, b T) int         // 綁定嘅三態比較器
    DrawTree(out io.Writer)               // ASCII 樹形輸出，方便 debug
}
```

### Treap

Treap = BST + 堆：節點按 `cmp` 排序（BST 不變量），同時按隨機 `priority` 排序（堆不變量），以高概率保持樹平衡。實作咗 `SelfBalancingBST[T]`。

```go
// 用三態比較器（-1 / 0 / 1）建立 treap。
func NewTreap[T any](cmp func(a, b T) int) *Treap[T]
```

`Treap[T]` 提供完整嘅 `SelfBalancingBST[T]` API（`Insert` / `Update` / `Get` / `Delete` / `Rank` / `Select` / `RangeVisit` / 各種遍歷 / `DrawTree` / ...）。

### 共享工具（bst 套件）

`bst` 套件仲提供獨立嘅輔助函數，方便自建基於節點嘅樹：

```go
type CMPFN[T any] = func(a, b T) int

// Node[T]：BSTNodeInterface 默認實作，自訂節點嵌入 *Node[T] 即可
func NewNode[T any](val T, cmp func(a, b T) int) *Node[T]

// 旋轉（返回子樹新根）同 size 維護
func RotateLeft[T any](p BSTNodeInterface[T]) BSTNodeInterface[T]
func RotateRight[T any](p BSTNodeInterface[T]) BSTNodeInterface[T]
func UpdateSize[T any](n BSTNodeInterface[T])
func RefreshSizeUp[T any](n BSTNodeInterface[T])
func IsLeftChild[T any](parent, child BSTNodeInterface[T]) bool

// 通用 BST 節點算法
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
func IsOrderedNode[T any](n BSTNodeInterface[T]) bool // 檢查 BST 有序不變量
func DrawTree[T any](p BSTNodeInterface[T], out io.Writer)
```

## 2. 依賴樹（dbtree）

為有依賴關係嘅表計算建置／執行順序（拓撲排序）：每個表都會出現喺佢依賴嘅表**之後**。

### 安裝

```go
import "github.com/leoheung/go-patterns/container/tree/dbtree"
```

### API

```go
// BuildDependencyOrder 按依賴順序返回表列表。
//   - tables：要排序嘅表
//   - tablesDependencies：表 → 佢依賴嘅表
//   - tables 為空時返回 nil
func BuildDependencyOrder(tables []string, tablesDependencies map[string][]string) []string
```

## 3. 二叉堆（heap）

泛型、基於陣列嘅二叉堆（默認最小堆：根係「最優先」嘅元素）。優先級由 `prioritize(a, b T) bool` 定義：`true` 表示 `a` 比 `b` 優先。

### 安裝

```go
import "github.com/leoheung/go-patterns/container/tree/heap"
```

### API

```go
// 建立二叉堆。prioritize 定義優先級；傳 nil 會 panic。
func NewBinaryHeap[T any](prioritize func(a, b T) bool) *BinaryHeap[T]

// 用已有切片建堆（O(n) heapify）。共享底層切片。
func NewBinaryHeapFrom[T any](data []T, prioritize func(a, b T) bool) *BinaryHeap[T]

func (h *BinaryHeap[T]) Len() int          // 元素個數，O(1)
func (h *BinaryHeap[T]) Push(x T)          // 入堆，O(log n)
func (h *BinaryHeap[T]) Pop() (T, bool)    // 取出堆頂並移除；空堆返回 (zero, false)，O(log n)
func (h *BinaryHeap[T]) Peek() (T, bool)   // 查看堆頂但唔移除；空堆返回 (zero, false)，O(1)
func (h *BinaryHeap[T]) Slice() []T        // 底層元素嘅副本（無序），只用嚟查看
```

## 完整範例

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/container/tree/bst/treap"
	"github.com/leoheung/go-patterns/container/tree/dbtree"
	"github.com/leoheung/go-patterns/container/tree/heap"
)

func main() {
	// ── Treap：支援 rank/select 嘅有序多重集 ──
	// 三態比較器：-1（a<b）/ 0（相等）/ 1（a>b）
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
	fmt.Println("大小:", t.Size()) // 5

	if min, ok := t.Min(); ok {
		fmt.Println("最小:", min) // 1
	}
	if v, ok := t.Select(2); ok {
		fmt.Println("rank 2:", v) // 5（0-based：1,3,5,7,9）
	}
	fmt.Println("7 嘅排名:", t.Rank(7)) // 3

	t.RangeVisit(3, 7, func(v int) { fmt.Print(v, " ") }) // 3 5 7
	fmt.Println()

	// ── dbtree：依賴順序（依賴排先）──
	order := dbtree.BuildDependencyOrder(
		[]string{"orders", "users"},
		map[string][]string{
			"orders": {"users"},
			"users":  {},
		},
	)
	fmt.Println("建置順序:", order) // [users orders]

	// ── heap：泛型優先隊列 ──
	h := heap.NewBinaryHeap(func(a, b int) bool { return a < b }) // 最小堆
	for _, v := range []int{5, 3, 9, 1, 7} {
		h.Push(v)
	}
	fmt.Println("堆大小:", h.Len()) // 5
	if top, ok := h.Peek(); ok {
		fmt.Println("堆頂:", top) // 1
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

## 特性

- **有序容器 + 順序統計**：treap 支援 `Rank` / `Select` / `Predecessor` / `Successor` / `RangeVisit`——最適合排行榜、調度器同區間掃描
- **多重集語義**：BST 框架入面相等元素可以共存（唔覆蓋）
- **可插拔平衡策略**：`SelfBalancingBST[T]` 契約共享；treap 係默認實作，另有共享旋轉／size 工具方便再做 splay / AVL / ...
- **拓撲排序**：`dbtree.BuildDependencyOrder` 計算依賴驅動嘅建置／執行順序
- **泛型二叉堆**：任何優先關係都得，O(log n) push/pop；由切片 O(n) heapify
- **Debug 工具**：`DrawTree` 將樹結構輸出成 ASCII 樹形圖
