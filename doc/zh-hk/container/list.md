# List

通用動態陣列實現，支援 Python list 及 JavaScript Array 操作。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/list"
```

## 基本操作

### 建立列表

```go
// 建立新的空列表
l := list.New[int]()

// 從切片建立列表
l := list.From([]int{1, 2, 3})

// 取得長度及容量
length := l.Len()
capacity := l.Cap()

// 轉換為切片
slice := l.ToSlice()

// 複製列表
clone := l.Clone()
```

## 元素存取

```go
// 以索引取得元素（支援負數索引）
elem := l.Get(0)      // 第一個元素
elem := l.Get(-1)     // 最後一個元素

// 以索引設定元素
l.Set(0, 10)

// 安全的元素存取
if elem, ok := l.At(0); ok {
    // 元素存在
}

// 返回修改某索引後的新 List 副本（原 List 不變）
l2 := l.With(0, 100)
```

## 新增與移除元素

```go
// 在末尾附加元素
l.Append(4, 5)
l.Push(6) // Append 的別名

// 以切片擴展
l.Extend([]int{7, 8})

// 在開頭新增元素
l.Unshift(0, -1)

// 在指定位置插入元素（支援負索引）
l.Insert(1, 99)

// 移除並返回第一個元素
if elem, ok := l.Shift(); ok {
    // 處理元素
}

// 移除並返回最後一個元素
if elem, ok := l.Pop(); ok {
    // 處理元素
}

// 移除指定值的第一次出現
l.RemoveFirst(5, func(a, b int) bool { return a == b })

// 移除指定索引的元素
if elem, ok := l.RemoveAt(2); ok {
    // 處理元素
}

// 清空列表
l.Clear()
```

## 進階修改操作 (JS-like)

```go
// Splice: 刪除並插入元素
removed := l.Splice(start, deleteCount, items...)

// CopyWithin: 將指定區間的元素複製到目標位置
l.CopyWithin(target, start, end)

// Fill: 用指定值填充區間元素
l.Fill(value, start, end)
```

## 搜尋及查詢

```go
// 檢查列表是否包含元素
contains := l.Includes(5, func(a, b int) bool { return a == b })

// 尋找元素索引
index := l.IndexOf(5, func(a, b int) bool { return a == b })
lastIndex := l.LastIndexOf(5, func(a, b int) bool { return a == b })

// 計算出現次數
count := l.Count(5, func(a, b int) bool { return a == b })

// 尋找元素
if elem, ok := l.Find(func(v, i int) bool { return v > 10 }); ok { /* ... */ }
index := l.FindIndex(func(v, i int) bool { return v > 10 })

// 從後搜尋
if elem, ok := l.FindLast(func(v, i int) bool { return v > 10 }); ok { /* ... */ }
lastIdx := l.FindLastIndex(func(v, i int) bool { return v > 10 })
```

## 遍歷與變換

```go
// 只讀遍歷
l.ForEach(func(v int, i int) { fmt.Println(v) })

// 並發遍歷
l.ForEachAsync(ctx, maxGoroutines, func(v int, i int) { /* ... */ })

// 映射元素到新列表 (包級泛型函數)
newList := list.Map(l, func(v int, i int) string { return fmt.Sprintf("%d", v) })

// 映射為 any 類型 (方法)
anyList := l.Map(func(v int, i int) any { return v * 2 })

// 並發映射
res, err := list.MapAsync(ctx, l, 4, func(v int, i int) int { return v * v })

// 過濾元素
filtered := l.Filter(func(v int, i int) bool { return v > 5 })

// 歸納元素 (包級泛型函數)
result := list.Reduce(l, 0, func(acc int, v int, i int) int { return acc + v })

// Every / Some
allMatch := l.Every(func(v int, i int) bool { return v > 0 })
anyMatch := l.Some(func(v int, i int) bool { return v > 100 })
```

## 排序與反轉

```go
// 就地排序
l.Sort(func(a, b int) bool { return a < b })

// 取得排序後的副本
lSorted := l.ToSorted(func(a, b int) bool { return a < b })

// 就地反轉
l.Reverse()

// 取得反轉後的副本
lReversed := l.ToReversed()

// 連接元素為字符串
str := l.Join(", ", func(v int) string { return fmt.Sprint(v) })
```

## 完整範例

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/container/list"
)

func main() {
    // 建立並填充列表
    l := list.From([]int{3, 1, 4, 1, 5, 9, 2, 6})

    // 過濾偶數
    evens := l.Filter(func(v, i int) bool { return v%2 == 0 })
    fmt.Println("偶數:", evens.ToSlice())

    // 映射為平方
    squares := list.Map(evens, func(v, i int) int { return v * v })
    fmt.Println("平方:", squares.ToSlice())

    // 排序
    squares.Sort(func(a, b int) bool { return a < b })
    fmt.Println("排序後:", squares.ToSlice())

    // 歸納為總和
    sum := list.Reduce(squares, 0, func(acc, v, i int) int { return acc + v })
    fmt.Println("總和:", sum)

    // 連接
    fmt.Println("連接結果:", squares.Join(" | ", func(v int) string { return fmt.Sprint(v) }))
}
```

## 特性

- **支援泛型**: 對任何數據類型均提供類型安全的操作
- **Python/JS 語義**: 提供如 `Append`, `Pop`, `Splice`, `Map`, `Filter` 等熟悉的介面
- **負數索引**: 支援 `l.Get(-1)` 直接存取最後一個元素
- **並發支持**: 提供 `ForEachAsync` 和 `MapAsync` 以利用多核性能
