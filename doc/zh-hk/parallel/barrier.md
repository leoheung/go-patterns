# Barrier

允許多個 Goroutine 互相等待到達某個點的同步原語。

## 安裝

```go
import "github.com/leoxiang66/go-patterns/parallel/barrier"
```

## API 參考

### 建立 Barrier

```go
// 為 N 個 Goroutine 建立新的 Barrier
b := barrier.NewBarrier(5)
```

### 等待

```go
// 等待所有 Goroutine 到達此點
b.Wait()
```

## 完整範例

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/leoxiang66/go-patterns/parallel/barrier"
)

func main() {
    const numWorkers = 3
    b := barrier.NewBarrier(numWorkers)
    var wg sync.WaitGroup

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            fmt.Printf("Worker %d: 階段 1\n", id)
            time.Sleep(time.Duration(id*100) * time.Millisecond)
            
            // 等待所有 Worker 到達此點
            b.Wait()
            
            fmt.Printf("Worker %d: 階段 2 (所有 Worker 已到達 Barrier)\n", id)
        }(i)
    }

    wg.Wait()
}
```

## 輸出

```
Worker 0: 階段 1
Worker 1: 階段 1
Worker 2: 階段 1
Worker 0: 階段 2 (所有 Worker 已到達 Barrier)
Worker 1: 階段 2 (所有 Worker 已到達 Barrier)
Worker 2: 階段 2 (所有 Worker 已到達 Barrier)
```

## 使用條件變數的 Barrier

使用條件變數的替代實現：

```go
import "github.com/leoxiang66/go-patterns/parallel/barrier"

// 使用條件變數建立 Barrier
b := barrier.NewBarrierWithCond(5)
b.Wait()
```

## 特性

- **循環**: 所有 Goroutine 通過後可重複使用
- **線程安全**: 可安全地並發使用
- **兩種實現**: 基於 Channel 及條件變數
