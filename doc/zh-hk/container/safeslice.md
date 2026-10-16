# SafeSlice

並發安全嘅泛型 Slice。

`SafeSlice[T]` 用 `sync.RWMutex` 包住一個普通嘅 `[]T`，為生產者／消費者同共享狀態場景提供安全嘅並發存取。讀取操作用讀鎖（唔會阻塞其他讀者）；修改操作用寫鎖。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/safeslice"
```

## API 參考

### 建立

```go
// 建立一個 SafeSlice。
//   - capacity：slice 容量（細過 0 會被設為 0，並且唔會細過 length）
//   - length：初始長度（細過 0 會被設為 0）
func NewSafeSlice[T any](capacity, length int) *SafeSlice[T]
```

### 元素存取

```go
// Peek 查看指定索引嘅元素但唔取出。
// 索引越界返回 (zero, false)。
func (l *SafeSlice[T]) Peek(index int) (T, bool)

// PeekFirst 查看隊首元素（索引 0）。
func (l *SafeSlice[T]) PeekFirst() (T, bool)

// PeekLast 查看隊尾元素。
func (l *SafeSlice[T]) PeekLast() (T, bool)
```

### 遍歷

```go
// Range 喺讀鎖下遍歷所有元素（唔會阻塞其他讀者）。
// f 返回 false 可以提前停止。
func (l *SafeSlice[T]) Range(f func(index int, item T) bool)
```

### 修改

```go
// Append 將元素追加到末尾。
func (l *SafeSlice[T]) Append(item T)

// RemoveIf 移除所有符合條件嘅元素，返回移除數量。
func (l *SafeSlice[T]) RemoveIf(predicate func(T) bool) int
```

### 查詢

```go
// Len 返回當前長度。
func (l *SafeSlice[T]) Len() int
```

## 完整範例

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/container/safeslice"
)

func main() {
	// 建立一個空嘅字符串 SafeSlice
	l := safeslice.NewSafeSlice[string](0, 0)

	// 追加元素
	l.Append("apple")
	l.Append("banana")
	l.Append("cherry")
	fmt.Println("長度:", l.Len()) // 3

	// 查看元素
	if v, ok := l.Peek(1); ok {
		fmt.Println("索引 1:", v) // banana
	}
	if v, ok := l.PeekFirst(); ok {
		fmt.Println("隊首:", v) // apple
	}
	if v, ok := l.PeekLast(); ok {
		fmt.Println("隊尾:", v) // cherry
	}

	// Range 提前停止遍歷
	l.Range(func(index int, item string) bool {
		fmt.Println(index, item)
		return index < 1 // 第二個元素之後停止
	})

	// RemoveIf：移除長度大過 5 嘅元素
	removed := l.RemoveIf(func(s string) bool { return len(s) > 5 })
	fmt.Println("移除數量:", removed) // 2（banana、cherry）
	fmt.Println("長度:", l.Len())     // 1
}
```

## 特性

- **並發安全**：所有操作都由 `sync.RWMutex` 保護
- **讀鎖遍歷**：`Range` 只持有讀鎖——並發讀者之間唔會互相阻塞
- **適合管理訂閱者列表**：`RemoveIf` 令維護訂閱者列表變得簡單（`circular` 同 `subscribe` 內部都用佢）
- **支援泛型**：任何元素類型 `T` 都用得
- **零值安全存取**：`Peek` 越界返回 `(zero, false)`，唔會 panic
