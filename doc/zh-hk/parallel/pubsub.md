# PubSub

發布-訂閱模式實現。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/pubsub"
```

## API 參考

### 初始化 PubSub 系統

```go
// 初始化 PubSub 系統
pubsub.InitPubSubSystem()

// 完成後關閉 PubSub 系統
// pubsub.Shutdown()
```

### 訂閱

```go
// 以緩衝區大小訂閱主題
id, ch, err := pubsub.Subscribe("topic-name", 10)
if err != nil {
    // 處理錯誤
}
```

### 發布

```go
// 發布消息到主題
err := pubsub.Publish("topic-name", message)
if err != nil {
    // 處理錯誤
}
```

### 取消訂閱

```go
// 取消訂閱主題
pubsub.Unsubscribe("topic-name", id)
```

訂閱通道由 Broker 關閉——唔好自己 close 佢。

### 關閉主題

```go
// 關閉某主題所有訂閱通道並移除該主題
err := pubsub.Close("topic-name")
```

### 關閉系統

```go
// 關閉所有訂閱通道並拆除整個系統
pubsub.Shutdown()
```

## 完整範例

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/parallel/pubsub"
)

func main() {
    // 初始化 PubSub 系統
    pubsub.InitPubSubSystem()
    defer pubsub.Shutdown()

    // 訂閱者 1
    id1, ch1, err := pubsub.Subscribe("news", 10)
    if err != nil {
        fmt.Printf("訂閱錯誤: %v\n", err)
        return
    }
    go func() {
        for msg := range ch1 {
            fmt.Printf("訂閱者 1 收到: %v\n", msg)
        }
    }()

    // 訂閱者 2
    id2, ch2, err := pubsub.Subscribe("news", 10)
    if err != nil {
        fmt.Printf("訂閱錯誤: %v\n", err)
        return
    }
    go func() {
        for msg := range ch2 {
            fmt.Printf("訂閱者 2 收到: %v\n", msg)
        }
    }()

    // 發布消息
    time.Sleep(100 * time.Millisecond)

    err = pubsub.Publish("news", "快訊: Go 1.22 發布了!")
    if err != nil {
        fmt.Printf("發布錯誤: %v\n", err)
    }

    err = pubsub.Publish("news", "快訊: 新增了模式!")
    if err != nil {
        fmt.Printf("發布錯誤: %v\n", err)
    }

    time.Sleep(100 * time.Millisecond)

    // 取消訂閱
    pubsub.Unsubscribe("news", id1)
    pubsub.Unsubscribe("news", id2)
}
```

## 輸出

```
訂閱者 1 收到: 快訊: Go 1.22 發布了!
訂閱者 2 收到: 快訊: Go 1.22 發布了!
訂閱者 1 收到: 快訊: 新增了模式!
訂閱者 2 收到: 快訊: 新增了模式!
```

## 特性

- **多個訂閱者**: 一對多消息分發
- **基於主題**: 按主題組織消息
- **異步傳遞**: 非阻塞發布——訂閱者緩衝區滿時消息會被丟棄
- **系統初始化**: 需要顯式初始化同關閉
- **緩衝支援**: 每個訂閱者可配置緩衝區大小(必須 `> 0`)
- **Broker 管理通道**: 訂閱通道只由 Broker 關閉(透過 `Unsubscribe`、`Close` 或 `Shutdown`)
- **錯誤處理**: 無效操作(例如冇訂閱者時發布)會返回錯誤
