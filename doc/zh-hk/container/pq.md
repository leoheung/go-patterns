# 優先隊列

通用優先隊列實現，支援自定義優先級比較。

## 安裝

```go
import "github.com/leoxiang66/go-patterns/container/pq"
```

## API 參考

### 建立優先隊列

```go
// 以指定容量及比較函數建立新的優先隊列
// better(a, b) 若 a 應排在 b 之前則返回 true
pq, err := pq.NewPriorityQueue[int](10, func(a, b int) bool { return a < b })
```

### 入隊

```go
// 將項目加入隊列
err := pq.Enqueue(5)
```

### 出隊

```go
// 取出最高優先級的項目
item, err := pq.Dequeue()
```

### 查看

```go
// 取得最高優先級的項目而不移除
item, err := pq.Peek()
```

### 隊列長度

```go
// 取得目前隊列長度
length := pq.Len()
```

## 完整範例

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/container/pq"
)

type Task struct {
    ID       int
    Priority int
    Name     string
}

func main() {
    // 建立優先隊列，高優先級任務優先處理
    pq, err := pq.NewPriorityQueue[Task](5, func(a, b Task) bool {
        return a.Priority > b.Priority
    })
    if err != nil {
        fmt.Printf("建立優先隊列錯誤: %v\n", err)
        return
    }

    // 加入不同優先級的任務
    tasks := []Task{
        {ID: 1, Priority: 3, Name: "任務 1"},
        {ID: 2, Priority: 1, Name: "任務 2"},
        {ID: 3, Priority: 5, Name: "任務 3"},
        {ID: 4, Priority: 2, Name: "任務 4"},
        {ID: 5, Priority: 4, Name: "任務 5"},
    }

    for _, task := range tasks {
        if err := pq.Enqueue(task); err != nil {
            fmt.Printf("任務 %d 入隊錯誤: %v\n", task.ID, err)
        }
    }

    // 按優先級順序處理任務
    for pq.Len() > 0 {
        task, err := pq.Dequeue()
        if err != nil {
            fmt.Printf("任務出隊錯誤: %v\n", err)
            continue
        }
        fmt.Printf("處理任務 %d: %s (優先級: %d)\n", task.ID, task.Name, task.Priority)
    }
}
```

## 輸出

```
處理任務 3: 任務 3 (優先級: 5)
處理任務 5: 任務 5 (優先級: 4)
處理任務 1: 任務 1 (優先級: 3)
處理任務 4: 任務 4 (優先級: 2)
處理任務 2: 任務 2 (優先級: 1)
```

## 特性

- **支援泛型**: 適用於任何類型
- **自定義比較**: 定義你自己的優先級邏輯
- **二元堆積**: 高效的 O(log n) 入隊/出隊操作
- **類型安全**: 編譯時類型檢查
