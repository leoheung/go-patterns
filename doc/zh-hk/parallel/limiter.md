# Limiter

用於控制操作速率的靜態限制器。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/limiter"
```

## API 參考

### 建立 Limiter

```go
// 以指定的時間間隔建立新的限制器
// 100ms 間隔 = 每秒 10 次操作
lim := limiter.NewStaticLimiter(100 * time.Millisecond)
```

### 等待令牌

```go
// 等待直到下一個令牌可用
lim.GrantNextToken()
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
    // 建立限制器: 每 200ms 1 次操作 (每秒 5 次)
    lim := limiter.NewStaticLimiter(200 * time.Millisecond)

    for i := 0; i < 5; i++ {
        start := time.Now()

        // 等待令牌
        lim.GrantNextToken()

        // 執行操作
        fmt.Printf("操作 %d 於 %v (經過: %v)\n",
            i+1, time.Now().Format("15:04:05.000"),
            time.Since(start))
    }
}
```

## 輸出

```
操作 1 於 14:30:00.000 (經過: 0s)
操作 2 於 14:30:00.200 (經過: 200ms)
操作 3 於 14:30:00.400 (經過: 200ms)
操作 4 於 14:30:00.600 (經過: 200ms)
操作 5 於 14:30:00.800 (經過: 200ms)
```

## 使用場景

- API 速率限制
- 資源節流
- 防止壓垮外部服務

## 特性

- **靜態速率**: 操作之間的固定間隔
- **阻塞**: 阻塞直到令牌可用
- **簡單**: 易於與 defer 一起使用
