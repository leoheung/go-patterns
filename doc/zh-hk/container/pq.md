# 優先隊列與調度器

提供通用的優先隊列實現及基於時間的任務調度管理器。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/pq"
```

## 1. 優先隊列 (Priority Queue)

### 建立隊列

```go
// 以指定容量及比較函數建立新的優先隊列
// better(a, b) 若 a 應排在 b 之前則返回 true
pq, err := pq.NewPriorityQueue[int](10, func(a, b int) bool { return a < b })
```

### 基本操作

```go
// 入隊
err := pq.Enqueue(5)

// 出隊（取出最高優先級的項目）
item, err := pq.Dequeue()

// 查看（取得最高優先級的項目而不移除）
item, err := pq.Peek()

// 取得目前隊列長度
length := pq.Len()
```

## 2. 任務調度管理器 (PriorityScheduledTaskManager)

用於在指定時間執行特定任務的調度器。

### 建立管理器

```go
ptm, err := pq.NewPriorityScheduledTaskManager()
if err != nil {
    // 處理錯誤
}
```

### 提交任務

```go
// 提交一個在 5 秒後執行的任務
cancel, err := ptm.PendNewTask(func() {
    fmt.Println("任務執行中...")
}, time.Now().Add(5 * time.Second))
```

### 停止管理器

```go
// 等待所有任務完成後優雅停止
err := ptm.FinishAndQuit()
```

## 3. 可取消對象 (Cancelable)

`PendNewTask` 返回的 `Cancelable` 對象可用於管理任務狀態。

```go
// 取消任務
success := cancel.Cancel()

// 恢復任務（若尚未執行）
success := cancel.Recover()

// 檢查是否已取消
isCanceled := cancel.IsCanceled()
```

## 完整範例

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/container/pq"
)

func main() {
    // 1. 使用調度器
    ptm, _ := pq.NewPriorityScheduledTaskManager()

    // 提交一個任務
    cancel, _ := ptm.PendNewTask(func() {
        fmt.Println("這是一個延遲任務")
    }, time.Now().Add(1 * time.Second))

    // 隨後決定取消它
    if cancel.Cancel() {
        fmt.Println("任務已成功取消")
    }

    // 2. 使用優先隊列
    queue, _ := pq.NewPriorityQueue[string](5, func(a, b string) bool {
        return len(a) < len(b) // 短字符串優先
    })

    queue.Enqueue("apple")
    queue.Enqueue("go")
    queue.Enqueue("banana")

    item, _ := queue.Dequeue()
    fmt.Printf("優先級最高（最短）: %s\n", item) // 輸出: go

    // 停止調度器
    ptm.FinishAndQuit()
}
```

## 特性

- **泛型支援**: `PriorityQueue` 適用於任何數據類型
- **精確調度**: `PriorityScheduledTaskManager` 內部使用優先隊列管理執行時間，確保下一個任務總是準時觸發
- **任務控制**: 每個調度任務都擁有獨立的 `Cancelable` 控制器
- **併發安全**: 所有操作均受互斥鎖保護
