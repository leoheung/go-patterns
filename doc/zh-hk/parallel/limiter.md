# Limiter

速率限制器,用於控制操作執行嘅頻率。套件提供三種實現:

- **StaticLimiter** — 固定間隔嘅令牌節流
- **TokenBucket** — 經典令牌桶,支援 Context 感知嘅阻塞式授權
- **SlidingWindowLimiter** — 滑動窗口請求計數器

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

- **精確計時**: 基於 `time.Ticker` 實現,保證穩定嘅時間間隔
- **線程安全**: 內部使用互斥鎖,支援多個 Goroutine 同時請求令牌
- **動態配置**: 支援透過 `Reset` 方法隨時調整速率
- **資源管理**: 提供 `Stop` 方法以便喺唔再需要時優雅地清理定時器資源

## 令牌桶(Token Bucket)

經典令牌桶限流器:以固定速率向桶中添加令牌,每個請求必須獲取一個令牌先至通過。桶滿時令牌溢出,桶空時請求被拒絕或等待。

### 建立令牌桶

```go
// capacity: 桶容量(最大令牌數)
// interval: 令牌補充間隔
tb := limiter.NewTokenBucket(100, 10*time.Millisecond)
```

桶初始係滿嘅(`capacity` 個令牌),每個 `interval` 補充一個令牌。

### TryGrant(非阻塞)

```go
ok, err := tb.TryGrant()
// ok == true  -> 攞到令牌
// ok == false, err == nil  -> 桶空
// err != nil  -> 桶已關閉
```

### Grant(阻塞,支援 Context)

```go
// 阻塞直到有令牌、ctx 完成或者桶被關閉
err := tb.Grant(ctx)
```

### 關閉

```go
tb.Close()  // 停止補充 Goroutine;之後嘅 Grant 會返回錯誤
```

### 令牌桶範例

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/leoheung/go-patterns/parallel/limiter"
)

func main() {
    // 10 個令牌嘅桶,每 100ms 補充一個
    tb := limiter.NewTokenBucket(10, 100*time.Millisecond)
    defer tb.Close()

    for i := 0; i < 15; i++ {
        if ok, err := tb.TryGrant(); ok {
            fmt.Printf("請求 %d 通過\n", i)
        } else {
            fmt.Printf("請求 %d 被拒: %v\n", i, err)
        }
    }
}
```

### 令牌桶特性

- **Context 感知**: `Grant` 尊重 Context 取消
- **非阻塞選項**: `TryGrant` 永不阻塞
- **突發支援**: 初始滿桶,按固定速率補充
- **可關閉**: `Close` 停止補充 Goroutine,並拒絕後續授權

## 滑動窗口限流器(Sliding Window Limiter)

滑動窗口請求計數器:每秒最多允許 `frequency_hz` 個請求,以 `buckets_count` 個時間桶劃分窗口計算。以低開銷近似實現滑動窗口限流。

### 建立滑動窗口限流器

```go
// frequency_hz: 每秒最大請求數
// buckets_count: 一秒窗口內嘅子桶數量
sw, err := limiter.NewSlidingWindowLimiter(100, 10)
if err != nil {
    // frequency_hz 或 buckets_count <= 0
}
```

### TryGrant

```go
ok, err := sw.TryGrant()
// ok == true  -> 請求允許
// ok == false, err == nil  -> 超出限流
// err != nil  -> 限流器已關閉
```

`TryGrant` 非阻塞,直接喺調用者 Goroutine 執行。

### 關閉

```go
sw.Close()  // 停止清理 Goroutine;之後嘅 TryGrant 會返回錯誤
```

### 滑動窗口範例

```go
package main

import (
    "fmt"

    "github.com/leoheung/go-patterns/parallel/limiter"
)

func main() {
    // 每秒 100 個請求,分成 10 個 100ms 嘅桶
    sw, err := limiter.NewSlidingWindowLimiter(100, 10)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer sw.Close()

    for i := 0; i < 150; i++ {
        ok, err := sw.TryGrant()
        if err != nil {
            fmt.Println(err)
            return
        }
        if !ok {
            fmt.Printf("請求 %d 被限流\n", i)
        }
    }
}
```

### 滑動窗口特性

- **近似滑動窗口**: 統計最近 `buckets_count` 個桶內嘅請求數
- **非阻塞**: `TryGrant` 直接執行,唔會等待
- **原子計數**: 桶計數器以 `sync/atomic` 更新
- **自動清理**: 過期桶由後台 Goroutine 定期移除
