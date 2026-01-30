# 信號量

用於限制資源並發存取的信號量實現。

## 安裝

```go
import "github.com/leoxiang66/go-patterns/parallel/semaphore"
```

## API 參考

### 建立信號量

```go
// 以指定的容量建立新的信號量
sem := semaphore.NewSemaphore(5) // 5 個許可
```

### 取得及釋放

```go
// 取得許可
sem.Acquire()
defer sem.Release()
```

## 完整範例

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/leoxiang66/go-patterns/parallel/semaphore"
)

func main() {
    // 建立有 3 個許可的信號量
    sem := semaphore.NewSemaphore(3)
    var wg sync.WaitGroup
    
    // 啟動 10 個 Goroutine
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            // 取得許可
            sem.Acquire()
            defer sem.Release()
            
            fmt.Printf("Goroutine %d: 執行中...\n", id)
            time.Sleep(200 * time.Millisecond)
            fmt.Printf("Goroutine %d: 完成\n", id)
        }(i)
    }
    
    wg.Wait()
}
```

## 輸出

```
Goroutine 0: 執行中...
Goroutine 1: 執行中...
Goroutine 2: 執行中...
Goroutine 0: 完成
Goroutine 3: 執行中...
...
```

## 使用條件變數的信號量

使用條件變數的替代實現：

```go
import "github.com/leoxiang66/go-patterns/parallel/semaphore"

// 使用條件變數建立信號量
sem := semaphore.NewSemaphoreByCond(5)
sem.Acquire()
sem.Release()
```

## 特性

- **資源限制**: 控制並發存取
- **兩種實現**: 基於 Channel 及條件變數
- **簡單 API**: Acquire 及 Release
