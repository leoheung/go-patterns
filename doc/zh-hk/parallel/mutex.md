# Mutex

簡單的互斥鎖實現。

## 安裝

```go
import "github.com/leoxiang66/go-patterns/parallel/mutex"
```

## API 參考

### 建立 Mutex

```go
// 建立新的 Mutex
m := mutex.NewMutex()
```

### 鎖定及解鎖

```go
// 鎖定 Mutex
m.Lock()

// 解鎖 Mutex (為安全起見使用 defer)
defer m.Unlock()
```

## 完整範例

```go
package main

import (
    "fmt"
    "sync"
    "github.com/leoxiang66/go-patterns/parallel/mutex"
)

func main() {
    m := mutex.NewMutex()
    counter := 0
    var wg sync.WaitGroup

    // 啟動 10 個 Goroutine
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            // 存取共享資源前鎖定
            m.Lock()
            defer m.Unlock()
            
            // 臨界區
            counter++
            fmt.Printf("計數器: %d\n", counter)
        }()
    }

    wg.Wait()
    fmt.Printf("最終計數器: %d\n", counter)
}
```

## 輸出

```
計數器: 1
計數器: 2
計數器: 3
計數器: 4
計數器: 5
計數器: 6
計數器: 7
計數器: 8
計數器: 9
計數器: 10
最終計數器: 10
```

## 特性

- **基於 Channel**: 使用 Go Channel 實現
- **簡單 API**: 只有 Lock 及 Unlock
- **FIFO 排序**: 公平的鎖定獲取
