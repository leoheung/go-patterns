# AsyncPool

## 概述

`AsyncPool` 是一個異步任務池，結合了工作池和任務緩衝區的功能，支持非阻塞任務提交。它允許在所有 workers 都忙時將任務排入緩衝區，而不是直接阻塞。

## 功能特點

- **工作池**: 固定數量的並發 workers
- **任務緩衝區**: 外部緩衝區用於排隊任務
- **優雅關閉**: 支持有序關閉
- **Panic 恢復**: 任務執行時內置 panic 處理
- **非阻塞提交**: 緩衝區滿時立即返回

## 架構

```
                    ┌─────────────────┐
                    │   任務緩衝區    │ (有容量限制)
                    │   (channel)     │
                    └────────▲────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────┴────┐        ┌────┴────┐        ┌────┴────┐
    │ Worker1 │        │ Worker2 │   ...  │ WorkerN │
    └─────────┘        └─────────┘        └─────────┘
```

## API 參考

### `NewAsyncPool(taskBufferCapacity int, numWorkers int) (*AsyncPool, error)`

創建具有指定緩衝區容量和 worker 數量的新 AsyncPool。

```go
pool, err := pool.NewAsyncPool(100, 10) // 100 緩衝區容量, 10 個 workers
if err != nil {
    log.Fatal(err)
}
```

### `(ap *AsyncPool) AsyncSubmit(task Task, onError OnError) error`

異步提交任務。任務類型為 `func() error`，onError 類型為 `func(error)`。

```go
err := pool.AsyncSubmit(
    func() error {
        // 執行一些工作
        return nil
    },
    func(err error) {
        fmt.Printf("任務失敗: %v\n", err)
    },
)
if err != nil {
    fmt.Printf("提交失敗: %v\n", err)
}
```

**行為：**
- 如果有可用的 worker，任務會立即在 goroutine 中執行
- 如果所有 workers 都忙，任務會被排入緩衝區
- 如果緩衝區滿了，返回錯誤
- 如果池已關閉，返回錯誤

### `(ap *AsyncPool) Shutdown()`

關閉池。調用 Shutdown 後：
- 新任務無法提交
- 緩衝區中的現有任務仍會執行
- Workers 在完成當前任務並清空緩衝區後退出

## 示例

```go
package main

import (
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/parallel/pool"
)

func main() {
	// 創建具有 100 任務緩衝區和 5 個 workers 的異步池
	ap, err := pool.NewAsyncPool(100, 5)
	if err != nil {
		panic(err)
	}

	// 提交 20 個任務
	for i := 0; i < 20; i++ {
		taskID := i
		err := ap.AsyncSubmit(
			func() error {
				fmt.Printf("任務 %d 開始\n", taskID)
				time.Sleep(500 * time.Millisecond)
				fmt.Printf("任務 %d 完成\n", taskID)
				return nil
			},
			func(err error) {
				fmt.Printf("任務錯誤: %v\n", err)
			},
		)
		if err != nil {
			fmt.Printf("提交任務 %d 失敗: %v\n", taskID, err)
		}
	}

	// 等待一段時間
	time.Sleep(3 * time.Second)

	// 關閉池
	ap.Shutdown()
	fmt.Println("池已關閉")
}
```

## 任務類型定義

```go
type Task func() error
type OnError func(error)
```

## 注意事項

- **緩衝區滿**: 當任務緩衝區滿時，`AsyncSubmit` 會立即返回錯誤。考慮增加緩衝區容量或實現重試邏輯。
- **Worker 效率**: Workers 高效地重用自己 - 完成任務後，如果緩衝區中有任務，它們會立即取下一個任務。
- **關閉順序**: `Shutdown()` 不會立即停止 workers。Workers 會在完成當前任務並清空緩衝區後退出。
- **Panic 處理**: 如果任務 panic，worker 會恢復並繼續處理下一個任務。
- **線程安全**: 所有公共方法都是並發安全的。