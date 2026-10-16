# Communication

並行程式入面消息傳遞同時間抽象嘅接口定義。本頁係 `parallel/communication` 底下接口嘅概覽:泛型 `MessageInterface`、`ClockInterface`,同 MQ 接口家族(`Publisher`、`Subscriber`、`Broadcaster`、`HttpClient`、`MQ`)。

::: tip
`MQ` 接口嘅 `MQLite` 實現詳情,請睇[消息隊列](../container/msgqueue.md)頁。
:::

## 安裝

```go
import (
    "github.com/leoheung/go-patterns/parallel/communication"
    "github.com/leoheung/go-patterns/parallel/communication/clock"
    "github.com/leoheung/go-patterns/parallel/communication/mq"
)
```

## API 參考

### MessageInterface

泛型消息抽象,帶住一個 `T` 類型嘅負載同整數時間戳。定義喺 `parallel/communication/msg.go`。

```go
type MessageInterface[T any] interface {
    SetMsg(msg T)
    GetMsg() T
    SetTime(time int)
    GetTime() int
}
```

任何有負載欄位同時間欄位嘅結構都可以實作佢——睇下面嘅範例。

### ClockInterface

最小化嘅時間抽象,畀需要「打拍」同讀取邏輯時間嘅組件使用。定義喺 `parallel/communication/clock/base.go`。

```go
type ClockInterface interface {
    Tick()
    Time() int
}
```

- `Tick()` 將時鐘推進一步。
- `Time()` 以 `int` 返回當前邏輯時間。

### MQ 接口

定義喺 `parallel/communication/mq/interface.go`。佢哋一齊定義咗基於 Topic 嘅消息隊列契約;現成實現係 `MQLite`。

#### 消息結構

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

#### Publisher

發布消息到所屬 Topic,並返回分配嘅消息 ID。

```go
type Publisher interface {
    Publish(ctx context.Context, msg *Message) (*string, error)
}
```

#### Subscriber

註冊同註銷訂閱者。一個訂閱會將 Topic 綁定到訂閱者,包括回調 URL 同 `maxInflight` 重試次數上限。

```go
type Subscriber interface {
    Subscribe(ctx context.Context, topic, subscriberID, callbackURL string, maxInflight int) error
    Unsubscribe(ctx context.Context, topic, subscriberID string) error
}
```

#### Broadcaster

將消息投遞畀該 Topic 嘅所有訂閱者。

```go
type Broadcaster interface {
    Broadcast(ctx context.Context, msg *Message) error
}
```

#### HttpClient

將單條消息發送畀單個訂閱者。`subscriber` 參數係內部(未匯出)類型,所以呢個接口係畀套件內部實作——實際上由 `MQLite` 實現。

```go
type HttpClient interface {
    Send(ctx context.Context, sub *subscriber, msg *Message) error
}
```

#### MQ

完整契約:發布、訂閱、廣播同發送,加上 `Close` 做生命週期管理。

```go
type MQ interface {
    Publisher
    Subscriber
    Broadcaster
    HttpClient
    Close(ctx context.Context) error
}
```

`MQLite` 實現詳情(構造函數、重試策略、`root_ctx` 生命週期),請睇[消息隊列](../container/msgqueue.md)。

## 範例:實作 MessageInterface 同 ClockInterface

```go
package main

import (
    "fmt"

    "github.com/leoheung/go-patterns/parallel/communication"
    "github.com/leoheung/go-patterns/parallel/communication/clock"
)

// Msg 帶 string 負載同邏輯時間戳嘅消息
type Msg struct {
    msg  string
    time int
}

// 實作 communication.MessageInterface[string]
func (m *Msg) SetMsg(msg string) { m.msg = msg }
func (m *Msg) GetMsg() string    { return m.msg }
func (m *Msg) SetTime(time int)  { m.time = time }
func (m *Msg) GetTime() int      { return m.time }

// SecondClock 每 tick 推進 1 嘅邏輯時鐘
type SecondClock struct{ t int }

// 實作 clock.ClockInterface
func (c *SecondClock) Tick()    { c.t++ }
func (c *SecondClock) Time() int { return c.t }

func main() {
    // 編譯期接口檢查
    var _ communication.MessageInterface[string] = (*Msg)(nil)
    var _ clock.ClockInterface = (*SecondClock)(nil)

    m := &Msg{msg: "hello"}
    c := &SecondClock{}

    for i := 0; i < 3; i++ {
        c.Tick()
        m.SetTime(c.Time())
    }

    fmt.Println(m.GetMsg(), m.GetTime()) // 輸出: hello 3
}
```

## 範例:使用 MQ 接口

```go
package main

import (
    "context"

    "github.com/leoheung/go-patterns/parallel/communication/mq"
)

func main() {
    ctx := context.Background()

    // MQLite 滿足 MQ 接口(詳情睇消息隊列頁)
    var q mq.MQ = mq.NewMQLite(100, 4, 64, 16)
    defer q.Close(ctx)

    _ = q.Subscribe(ctx, "orders", "sub-1", "http://localhost:8080/callback", 3)

    payload := []byte(`{"event":"created"}`)
    id, err := q.Publish(ctx, &mq.Message{Topic: "orders", Payload: payload})
    _ = id
    _ = err
}
```

## 特性

- **泛型消息抽象**:`MessageInterface[T]` 適用於任何負載類型
- **邏輯時間**:`ClockInterface` 將時間同真實時鐘解耦
- **MQ 契約**:`Publisher` / `Subscriber` / `Broadcaster` / `HttpClient` 組合成 `MQ`
- **支援 Context**:所有 MQ 操作都接收 `context.Context`
- **由 MQLite 實現**:睇[消息隊列](../container/msgqueue.md)
