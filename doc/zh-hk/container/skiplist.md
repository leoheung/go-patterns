# SkipList

泛型跳表（Skip List），可選並發安全。

跳表係一個概率性嘅有序數據結構，平均提供 O(log n) 嘅搜尋、插入同刪除——係平衡樹嘅簡單替代方案。呢個實作係泛型（`T any`），配以自訂比較函數，同時提供純單 goroutine 模式同讀寫鎖並發模式。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/skiplist"
```

## 概念

- **排序函數** `better(a, b T) bool`：當 `a` 排喺 `b` 前面（即 `a` 細啲／優先級高啲）時返回 `true`。所有排序行為（搜尋位置、`GetMin`、`RangeQuery` 等）都由佢決定。
- **層數同概率**：每個節點有隨機嘅層數。默認配置係 `DefaultMaxLevel = 16`、`DefaultProbability = 0.5`（每層以 1/2 概率晉升）。
- **插入去重**：如果已存在相等元素，`Insert` 會直接忽略（唔會重複插入）。
- **可選加鎖**：`withRWLock` 決定係咪由 `sync.RWMutex` 保護所有操作。共享並發使用就傳 `true`，單 goroutine 使用就傳 `false`（零鎖開銷）。

## API 參考

### 建立

```go
const (
	DefaultMaxLevel    = 16  // 最大層數
	DefaultProbability = 0.5 // 晉升概率 1/2
)

// 用默認配置建立跳表。
// better：比較函數——true 表示 a 排喺 b 前面。
// withRWLock：true 啟用內置讀寫鎖（並發安全）。
func New[T any](better func(a, b T) bool, withRWLock bool) *SkipList[T]

// 用自訂嘅 maxLevel 同晉升概率建立跳表。
func NewWithConfig[T any](better func(a, b T) bool, maxLevel int, probability float64, withRWLock bool) *SkipList[T]
```

### 核心操作

```go
// Insert 插入元素。重複元素會被忽略（去重）。
func (sl *SkipList[T]) Insert(value T)

// Search 搜尋元素並返回佢並返回佢嘅節點，搵唔到就返回 nil。
// 注意：節點欄位係未匯出嘅；返回值主要用嚟判斷搵唔搵到——
// 純判斷存在與否建議用 Contains。
func (sl *SkipList[T]) Search(value T) *Node[T]

// Delete 刪除元素。成功移除返回 true，搵唔到返回 false。
func (sl *SkipList[T]) Delete(value T) bool
```

### 查詢操作

```go
// RangeQuery 返回閉區間 [min, max] 內嘅所有元素（升序）。
func (sl *SkipList[T]) RangeQuery(min, max T) []T

// GetMin 返回最小元素（空表返回零值）。
func (sl *SkipList[T]) GetMin() T

// GetMax 返回最大元素（空表返回零值）。
func (sl *SkipList[T]) GetMax() T

// Contains 檢查元素係咪存在。
func (sl *SkipList[T]) Contains(value T) bool

// GetAll 返回所有元素（升序）。
func (sl *SkipList[T]) GetAll() []T

// Len 返回元素數量。
func (sl *SkipList[T]) Len() int
```

### 工具

```go
// Clear 清空跳表。
func (sl *SkipList[T]) Clear()
```

## 完整範例

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/container/skiplist"
)

func main() {
	// 建立一個並發安全、升序排列嘅整數跳表
	sl := skiplist.New(func(a, b int) bool { return a < b }, true)

	// 插入元素（重複元素會被忽略）
	for _, v := range []int{5, 3, 9, 1, 7, 3} {
		sl.Insert(v)
	}
	fmt.Println("長度:", sl.Len()) // 5

	// Search / Contains
	fmt.Println("包含 7:", sl.Contains(7))   // true
	fmt.Println("包含 4:", sl.Contains(4))   // false

	// Min / Max
	fmt.Println("最小:", sl.GetMin()) // 1
	fmt.Println("最大:", sl.GetMax()) // 9

	// 區間查詢 [3, 7]
	fmt.Println("區間 [3,7]:", sl.RangeQuery(3, 7)) // [3 5 7]

	// 所有元素（有序）
	fmt.Println("全部:", sl.GetAll()) // [1 3 5 7 9]

	// 刪除
	fmt.Println("刪除 5:", sl.Delete(5)) // true
	fmt.Println("刪除 5:", sl.Delete(5)) // false（已經冇咗）

	// 清空
	sl.Clear()
	fmt.Println("清空後長度:", sl.Len()) // 0
}
```

## 特性

- **平均 O(log n) 複雜度**：搜尋、插入、刪除由層級晉升概率性地保持平衡
- **泛型 + 自訂排序**：任何類型 `T` 配任何全序關係，透過 `better` 定義
- **可選並發安全**：`withRWLock` 開關內置 `RWMutex` 加鎖——需要幾多就俾幾多
- **插入去重**：相等元素唔會重複插入
- **區間查詢**：閉區間 `RangeQuery(min, max)`，升序返回
- **可配置結構**：透過 `NewWithConfig` 自訂 `maxLevel` 同晉升 `probability`
