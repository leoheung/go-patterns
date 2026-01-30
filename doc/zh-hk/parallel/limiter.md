# Limiter

靜態速率限制器，用於控制操作執行的頻率。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/limiter"
```

## API 參考

### 建立 Limiter

```go
// 建立指定時間間隔的限制器
// 100ms 間隔 = 每秒 10 次操作
lim := limiter.NewStaticLimiter(100 * time.Millisecond)
```

### 等待令牌

```go
// 阻塞調用：等待直到下一個令牌可用
lim.GrantNextToken()
```

### 控制操作

```go
// 在運行時重置限制器的時間間隔
lim.Reset(200 * time.Millisecond)

// 停止底層定時器以釋放資源
lim.Stop()
```

## 完整範例

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/parallel/limiter"
)

func main() {
    // 建立限制器：每 200ms 執行 1 次（每秒 5 次）
    lim := limiter.NewStaticLimiter(200 * time.Millisecond)
    defer lim.Stop()

    for i := 0; i < 5; i++ {
        start := time.Now()

        // 等待令牌
        lim.GrantNextToken()

        // 執行操作
        fmt.Printf("操作 %d 於 %v (耗時: %v)\n",
            i+1, time.Now().Format("15:04:05.000"),
            time.Since(start))
    }
}
```

## 特性

- **精確計時**: 基於 `time.Ticker` 實現，保證穩定的時間間隔。
- **線程安全**: 內部使用互斥鎖，支援多個 Goroutine 同時請求令牌。
- **動態配置**: 支援透過 `Reset` 方法隨時調整速率。
- **資源管理**: 提供 `Stop` 方法以便在不再需要時優雅地清理定時器資源。
