# Semaphore (信號量)

用於限制對共享資源並發訪問數量的信號量實現。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/semaphore"
```

## API 參考

### 建立信號量

```go
// 建立基於通道（Channel）的信號量
sem := semaphore.NewSemaphore(5)

// 建立基於條件變量（Condition Variable）的信號量
semCond := semaphore.NewSemaphoreByCond(5)
```

### 獲取與釋放

```go
// 獲取許可（阻塞）
sem.Acquire()

// 嘗試獲取許可（非阻塞，若無可用許可則立即返回 false）
success := sem.TryAcquire()

// 釋放許可
sem.Release()
```

## 完整範例

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/leoheung/go-patterns/parallel/semaphore"
)

func main() {
    // 建立一個只有 3 個許可的信號量
    sem := semaphore.NewSemaphore(3)
    var wg sync.WaitGroup

    // 啟動 10 個 Goroutine
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            // 獲取許可
            sem.Acquire()
            defer sem.Release()

            fmt.Printf("Goroutine %d: 正在工作...\n", id)
            time.Sleep(200 * time.Millisecond)
            fmt.Printf("Goroutine %d: 完成\n", id)
        }(i)
    }

    wg.Wait()
}
```

## 特性

- **資源限制**: 嚴格控制並發操作的數量。
- **兩種實現方式**:
  - `Semaphore`: 基於緩衝通道，簡潔且符合 Go 慣用法。
  - `SemaphoreByCond`: 基於互斥鎖與條件變量，適用於特定的同步場景。
- **非阻塞支持**: 提供 `TryAcquire` 接口，允許在不等待的情況下檢查資源可用性。
