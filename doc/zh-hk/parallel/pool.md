# Worker Pool

基於信號量（Semaphore）實現的輕量級工作池，用於精確控制並發任務數量。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/pool"
```

## API 參考

### 建立 Worker Pool

```go
// 建立具有固定 Worker 數量的 Pool
p := pool.NewWorkerPool(5)
```

### 提交任務

```go
// 阻塞式提交：若無可用 Worker 則等待
p.Submit(func() {
    // 任務邏輯
})

// 非阻塞式提交：若無可用 Worker 則立即返回 false
success := p.TrySubmit(func() {
    // 任務邏輯
})
```

## 完整範例

```go
package main

import (
    "fmt"
    "sync/atomic"
    "time"
    "github.com/leoheung/go-patterns/parallel/pool"
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

    // 等待任務完成（實際應用中建議使用 WaitGroup）
    time.Sleep(1 * time.Second)

    fmt.Printf("總共完成: %d\n", counter)
}
```

## 特性

- **並發控制**: 嚴格限制同時運行的 Goroutine 數量。
- **崩潰保護**: 自動捕獲任務中的 Panic 並記錄日誌，防止單個任務崩潰導致整個進程退出。
- **靈活提交**: 支持阻塞 (`Submit`) 與非阻塞 (`TrySubmit`) 兩種任務調度方式。
- **高效能**: 基於 `SemaphoreByCond` 實現，內存佔用極低。
