# 消息隊列

::: tip
MQ 實現位於 `parallel/communication/mq`(唔喺 `container/`)。本頁保留喺原位置以免斷開連結。
:::

輕量級基於 Topic 嘅消息隊列(`MQLite`),透過 HTTP 回調將消息推送畀訂閱者。佢建立喺 `parallel/stream`、`parallel/pool` 同 `container/safemap` 之上,並透過內部 `root_ctx` 管理整個生命週期。

## 安裝

```go
import "github.com/leoheung/go-patterns/parallel/communication/mq"
```

## API 參考

### 消息結構

```go
type MessageHeader struct {
    TraceID   string    `json:"traceID"`
    Source    string    `json:"source"`
    EventType string    `json:"eventType"`
    Timestamp time.Time `json:"timestamp"`
    Version   string    `json:"version"`
}

type Message struct {
    ID      string        `json:"id"`
    Topic   string        `json:"topic"`
    Payload []byte        `json:"payload"`
    Headers MessageHeader `json:"headers"`
}
```

`Message.ID` 由 `Publish` 自動分配(遞增計數器)。

### 建立消息隊列

```go
// 以消息緩衝區大小、工作者數量、
// 任務隊列大小及分片數建立新嘅 MQ
q := mq.NewMQLite(100, 4, 64, 16)
```

所有參數必須 `>= 0`,否則返回 `nil`。

### 發布

```go
// 將消息發布到 Topic,返回分配嘅消息 ID
id, err := q.Publish(ctx, &mq.Message{
    Topic:   "orders",
    Payload: []byte(`{"orderId":"42"}`),
})
```

- Topic 必須至少有一個訂閱者,否則返回錯誤。
- 發送係非阻塞:如果消息緩衝區已滿,會立即返回錯誤。

### 訂閱 / 取消訂閱

```go
// 以 HTTP 回調 URL 同最大重試次數註冊訂閱者
err := q.Subscribe(ctx, "orders", "sub-1", "http://localhost:8080/callback", 3)

// 從 Topic 移除訂閱者
err = q.Unsubscribe(ctx, "orders", "sub-1")
```

投遞係推送式:每條消息會以 JSON POST 到訂閱者嘅 `callbackURL`。

### 廣播與發送

```go
// 將消息投遞畀 msg.Topic 嘅每個訂閱者
err := q.Broadcast(ctx, msg)
```

`Send(ctx, sub, msg)` 負責將單條消息 POST 畀一個訂閱者,由 `Broadcast` 內部調用。兩者都係 `MQ` 接口嘅一部分。

### 關閉

```go
// 取消內部 root_ctx 以關閉隊列
err := q.Close(ctx)
```

## 完整範例

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/leoheung/go-patterns/parallel/communication/mq"
)

func main() {
    q := mq.NewMQLite(100, 4, 64, 16)
    defer q.Close(context.Background())

    // 註冊訂閱者
    err := q.Subscribe(context.Background(), "orders", "sub-1",
        "http://localhost:8080/callback", 3)
    if err != nil {
        fmt.Printf("訂閱錯誤: %v\n", err)
        return
    }

    // 發布消息
    id, err := q.Publish(context.Background(), &mq.Message{
        Topic:   "orders",
        Payload: []byte(`{"orderId":"42"}`),
        Headers: mq.MessageHeader{
            EventType: "order.created",
            Timestamp: time.Now(),
        },
    })
    if err != nil {
        fmt.Printf("發布錯誤: %v\n", err)
        return
    }
    fmt.Printf("已發布消息 %s\n", *id)
}
```

## 特性

- **基於 Topic**:訂閱者按 Topic 註冊,並提供 HTTP 回調 URL
- **HTTP 投遞**:消息以 JSON POST 到訂閱者端點
- **自動重試**:投遞失敗會按 `maxInflight` 次數重試,並採用遞增退避
- **非阻塞發布**:緩衝區滿時返回錯誤,唔會阻塞生產者
- **Root Context 生命週期**:`Close` 取消內部 `root_ctx`,停止所有流水線
- **並發投遞**:廣播喺 `parallel/pool.AsyncPoolV2` 工作者池上執行
