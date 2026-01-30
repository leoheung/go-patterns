# 消息隊列

基於 Channel 的消息隊列實現，具備基本隊列操作。

## 安裝

```go
import "github.com/leoxiang66/go-patterns/container/msgQueue"
```

## API 參考

### 建立消息隊列

```go
// 以指定容量及裝置 ID 建立新的消息隊列
mq := msgqueue.NewChanMQ(100, "device-1")
```

### 入隊

```go
// 將消息加入隊列
err := mq.Enq([]byte("hello world"))
```

### 出隊

```go
// 以 Context 從隊列取出消息
ctx := context.Background()
msg, err := mq.Deq(ctx)
```

### 隊列操作

```go
// 取得目前隊列長度
length := mq.Len()

// 清空所有消息
err := mq.Clear()

// 檢查隊列是否運作中
isLive := mq.IsLive()

// 重新啟動已停止的隊列
mq.Renew()

// 銷毀隊列
mq.Destroy()
```

## 完整範例

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/leoxiang66/go-patterns/container/msgQueue"
)

func main() {
    // 建立容量為 10 的消息隊列
    mq := msgqueue.NewChanMQ(10, "test-device")
    defer mq.Destroy()

    // 啟動 Goroutine 消費消息
    go func() {
        ctx := context.Background()
        for i := 0; i < 5; i++ {
            msg, err := mq.Deq(ctx)
            if err != nil {
                fmt.Printf("出隊錯誤: %v\n", err)
                return
            }
            fmt.Printf("收到消息: %s\n", string(msg))
            time.Sleep(500 * time.Millisecond)
        }
    }()

    // 入隊消息
    for i := 0; i < 5; i++ {
        msg := fmt.Sprintf("消息 %d", i)
        if err := mq.Enq([]byte(msg)); err != nil {
            fmt.Printf("入隊錯誤: %v\n", err)
            return
        }
        fmt.Printf("發送消息: %s\n", msg)
    }

    // 等待所有消息處理完成
    time.Sleep(3 * time.Second)
}
```

## 特性

- **基於 Channel**: 以 Go Channel 建立，實現高效並發操作
- **支援 Context**: 透過 Context 支援取消操作
- **生命週期管理**: 建立、重新啟動及銷毀隊列
- **線程安全**: 可安全地由多個 Goroutine 並發使用
