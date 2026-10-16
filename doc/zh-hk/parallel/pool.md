# 池

`pool` 套件提供三個面向並發嘅池:

- **WorkerPool** — 基於信號量嘅工作池,精確控制並發數量
- **AsyncPoolV2** — 支援 Context 嘅異步任務池,內置有界任務隊列同統計
- **ObjPool** — 通用對象池,避免頻繁分配同 GC 回收

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/pool"
```

## 任務類型

```go
type Task func() error            // WorkerPool 用
type OnError func(error)          // 錯誤回調(任務返回錯誤時被調用)
type TaskWithCtx func(ctx context.Context) error  // AsyncPoolV2 用
```

## WorkerPool

基於信號量嘅輕量級工作池,用嚟限制並發 Goroutine 數量。

### 建立 Worker Pool

```go
// 以固定 Worker 數量建立工作池(必須 > 0)
p := pool.NewWorkerPool(5)
```

### 提交任務

```go
// 阻塞式提交:冇可用 Worker 就等待
p.Submit(func() error {
    // 任務邏輯
    return nil
}, func(err error) {
    // 可選嘅錯誤回調
})

// 非阻塞式提交:冇可用 Worker 立即返回 false
success := p.TrySubmit(func() error {
    // 任務邏輯
    return nil
}, nil)
```

### 完整範例

```go
package main

import (
    "fmt"
    "sync/atomic"
    "time"

    "github.com/leoheung/go-patterns/parallel/pool"
)

func main() {
    // 建立有 3 個 Worker 嘅 Pool
    p := pool.NewWorkerPool(3)

    var counter int32 = 0

    // 提交 10 個任務
    for i := 0; i < 10; i++ {
        taskID := i
        p.Submit(func() error {
            fmt.Printf("任務 %d 開始\n", taskID)
            time.Sleep(100 * time.Millisecond)
            atomic.AddInt32(&counter, 1)
            fmt.Printf("任務 %d 完成\n", taskID)
            return nil
        }, nil)
    }

    // 等待任務完成
    time.Sleep(1 * time.Second)

    fmt.Printf("總共完成: %d\n", counter)
}
```

### WorkerPool 特性

- **並發控制**: 嚴格限制同時運行嘅 Goroutine 數量
- **崩潰保護**: 自動捕獲任務中嘅 Panic 並記錄日誌
- **錯誤處理**: 任務返回 `error`;可選嘅 `onError` 回調會喺失敗時被調用
- **靈活提交**: 支援阻塞(`Submit`)同非阻塞(`TrySubmit`)兩種調度方式
- **高效能**: 基於 `SemaphoreByCond` 實現,內存佔用極低

## AsyncPoolV2

支援 Context 嘅異步任務池,內置有界任務隊列。Worker 透過 `utils.MergeContexts` 將調用者 Context 同池自身 Context 合併,因此任務級取消同池關閉都會生效。

### 建立池

```go
// maxWorkers >= 0、queueSize >= 0(否則返回 nil)
p := pool.NewAsyncPoolV2(4, 64)
```

### 提交任務

```go
err := p.AsyncSubmit(ctx, func(ctx context.Context) error {
    // 支援 Context 嘅任務邏輯
    return nil
}, func(err error) {
    // 可選嘅錯誤回調
})
```

- 如果有空閒 Worker,任務即刻啟動;否則入隊等待
- 隊列滿咗會返回錯誤(`task queue if full`)
- `Shutdown` 之後再提交會返回錯誤

### 統計與關閉

```go
s := p.Stats()   // { Running, Pending, Completed, Failed }

p.Shutdown()     // 取消池 Context;Worker 執行完當前任務後停止
```

### AsyncPoolV2 範例

```go
package main

import (
    "context"
    "fmt"

    "github.com/leoheung/go-patterns/parallel/pool"
)

func main() {
    p := pool.NewAsyncPoolV2(4, 64)
    defer p.Shutdown()

    for i := 0; i < 10; i++ {
        taskID := i
        err := p.AsyncSubmit(context.Background(), func(ctx context.Context) error {
            fmt.Printf("任務 %d 運行中\n", taskID)
            return nil
        }, func(err error) {
            fmt.Printf("任務錯誤: %v\n", err)
        })
        if err != nil {
            fmt.Printf("提交錯誤: %v\n", err)
        }
    }

    fmt.Printf("統計: %+v\n", p.Stats())
}
```

### AsyncPoolV2 特性

- **Context 傳播**: 每個任務以合併 Context 運行(任務 ctx + 池 ctx)
- **有界隊列**: 多餘任務入隊;隊列滿時拒絕提交並返回錯誤
- **統計**: 實時計數 running/pending/completed/failed
- **優雅關閉**: `Shutdown` 取消池 Context

## ObjPool

固定大小嘅通用對象池,避免頻繁分配同 GC 回收。同 `sync.Pool` 唔同,呢度嘅對象**唔會**被 GC 清除。

### 建立對象池

```go
// poolSize 必須 > 0
p, err := pool.NewObjPool(10, func() *Buffer {
    return &Buffer{}
})
```

池會以 `newFn` 預先創建 `poolSize` 個對象填充。

### Get / Put

```go
// 從池中取得對象(池空時創建新對象)
buf := p.Get()

// 歸還對象(池滿時靜默丟棄)
p.Put(buf)
```

### ObjPool 範例

```go
pool, err := pool.NewObjPool(10, func() *Connection {
    return OpenConnection()
})
if err != nil {
    log.Fatal(err)
}

// 從池中取得對象
conn := pool.Get()
defer pool.Put(conn)

// 使用 conn...
```

### ObjPool 特性

- **泛型**: 適用於任何類型 `T`
- **預填充**: 初始化時預創建 `poolSize` 個對象
- **非阻塞**: `Get` 喺池空時創建新對象;`Put` 喺池滿時丟棄
- **GC 友好**: 對象唔會被 GC 清除——適合重用嘅熱路徑
- **線程安全**: `Get`/`Put` 可以並發調用
