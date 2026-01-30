# Worker Pool

用於管理並發任務的 Worker Pool 模式。

## 安裝

```go
import "github.com/leoxiang66/go-patterns/parallel/pool"
```

## API 參考

### 建立 Worker Pool

```go
// 以指定的 Worker 數量建立 Worker Pool
p := pool.NewWorkerPool(5) // 5 個 Worker
```

### 提交任務

```go
// 提交任務到 Pool
p.Submit(func() {
    // 任務邏輯
})
```

### 停止 Pool

```go
// 優雅地停止 Pool
p.Stop()
```

## 完整範例

```go
package main

import (
    "fmt"
    "sync/atomic"
    "time"
    "github.com/leoxiang66/go-patterns/parallel/pool"
)

func main() {
    // 建立有 3 個 Worker 的 Pool
    p := pool.NewWorkerPool(3)
    
    var counter int32 = 0
    
    // 提交 10 個任務
    for i := 0; i < 10; i++ {
        taskID := i
        p.Submit(func() {
            fmt.Printf("任務 %d 開始\n", taskID)
            time.Sleep(100 * time.Millisecond)
            atomic.AddInt32(&counter, 1)
            fmt.Printf("任務 %d 完成\n", taskID)
        })
    }
    
    // 等待任務完成
    time.Sleep(1 * time.Second)
    
    // 停止 Pool
    p.Stop()
    
    fmt.Printf("總共完成: %d\n", counter)
}
```

## 輸出

```
任務 0 開始
任務 1 開始
任務 2 開始
任務 0 完成
任務 3 開始
任務 1 完成
任務 4 開始
...
總共完成: 10
```

## 特性

- **固定大小 Pool**: 控制並發級別
- **任務隊列**: 用於待處理任務的緩衝 Channel
- **優雅關閉**: 等待運行中的任務完成
