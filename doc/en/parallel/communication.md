# Communication

Interface definitions for message passing and time abstraction in concurrent programs. This page is an overview of the interfaces declared under `parallel/communication`: the generic `MessageInterface`, the `ClockInterface`, and the MQ interface family (`Publisher`, `Subscriber`, `Broadcaster`, `HttpClient`, `MQ`).

::: tip
The `MQLite` implementation of the `MQ` interface is documented on the [Message Queue](../container/msgqueue.md) page.
:::

## Installation

```go
import (
    "github.com/leoheung/go-patterns/parallel/communication"
    "github.com/leoheung/go-patterns/parallel/communication/clock"
    "github.com/leoheung/go-patterns/parallel/communication/mq"
)
```

## API Reference

### MessageInterface

A generic message abstraction that carries a payload of type `T` plus an integer timestamp. Declared in `parallel/communication/msg.go`.

```go
type MessageInterface[T any] interface {
    SetMsg(msg T)
    GetMsg() T
    SetTime(time int)
    GetTime() int
}
```

Any struct with a payload field and a time field can implement it — see the example below.

### ClockInterface

A minimal time abstraction used by components that need to tick and read a logical time. Declared in `parallel/communication/clock/base.go`.

```go
type ClockInterface interface {
    Tick()
    Time() int
}
```

- `Tick()` advances the clock by one step.
- `Time()` returns the current logical time as an `int`.

### MQ Interfaces

Declared in `parallel/communication/mq/interface.go`. Together they define the contract of a topic-based message queue; the shipped implementation is `MQLite`.

#### Message Structures

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

Publishes a message to its topic and returns the assigned message ID.

```go
type Publisher interface {
    Publish(ctx context.Context, msg *Message) (*string, error)
}
```

#### Subscriber

Registers and unregisters subscribers. A subscription binds a topic to a subscriber with a callback URL and a `maxInflight` retry budget.

```go
type Subscriber interface {
    Subscribe(ctx context.Context, topic, subscriberID, callbackURL string, maxInflight int) error
    Unsubscribe(ctx context.Context, topic, subscriberID string) error
}
```

#### Broadcaster

Delivers a message to every subscriber of its topic.

```go
type Broadcaster interface {
    Broadcast(ctx context.Context, msg *Message) error
}
```

#### HttpClient

Sends one message to one subscriber. The `subscriber` parameter is an internal (unexported) type, so this interface is meant to be implemented inside this package — in practice it is implemented by `MQLite`.

```go
type HttpClient interface {
    Send(ctx context.Context, sub *subscriber, msg *Message) error
}
```

#### MQ

The full contract: publish, subscribe, broadcast and send, plus `Close` for lifecycle management.

```go
type MQ interface {
    Publisher
    Subscriber
    Broadcaster
    HttpClient
    Close(ctx context.Context) error
}
```

For the `MQLite` implementation details (constructor, retry policy, `root_ctx` lifecycle), see [Message Queue](../container/msgqueue.md).

## Example: Implementing MessageInterface and ClockInterface

```go
package main

import (
    "fmt"

    "github.com/leoheung/go-patterns/parallel/communication"
    "github.com/leoheung/go-patterns/parallel/communication/clock"
)

// Msg is a typed message carrying a string payload and a logical timestamp
type Msg struct {
    msg  string
    time int
}

// Implement communication.MessageInterface[string]
func (m *Msg) SetMsg(msg string) { m.msg = msg }
func (m *Msg) GetMsg() string    { return m.msg }
func (m *Msg) SetTime(time int)  { m.time = time }
func (m *Msg) GetTime() int      { return m.time }

// SecondClock is a logical clock that advances by 1 on every tick
type SecondClock struct{ t int }

// Implement clock.ClockInterface
func (c *SecondClock) Tick()    { c.t++ }
func (c *SecondClock) Time() int { return c.t }

func main() {
    // Compile-time interface checks
    var _ communication.MessageInterface[string] = (*Msg)(nil)
    var _ clock.ClockInterface = (*SecondClock)(nil)

    m := &Msg{msg: "hello"}
    c := &SecondClock{}

    for i := 0; i < 3; i++ {
        c.Tick()
        m.SetTime(c.Time())
    }

    fmt.Println(m.GetMsg(), m.GetTime()) // Output: hello 3
}
```

## Example: Using the MQ Interface

```go
package main

import (
    "context"

    "github.com/leoheung/go-patterns/parallel/communication/mq"
)

func main() {
    ctx := context.Background()

    // MQLite satisfies the MQ interface (see the Message Queue page for details)
    var q mq.MQ = mq.NewMQLite(100, 4, 64, 16)
    defer q.Close(ctx)

    _ = q.Subscribe(ctx, "orders", "sub-1", "http://localhost:8080/callback", 3)

    payload := []byte(`{"event":"created"}`)
    id, err := q.Publish(ctx, &mq.Message{Topic: "orders", Payload: payload})
    _ = id
    _ = err
}
```

## Features

- **Generic message abstraction**: `MessageInterface[T]` works with any payload type
- **Logical time**: `ClockInterface` decouples time from the wall clock
- **MQ contract**: `Publisher` / `Subscriber` / `Broadcaster` / `HttpClient` composed into `MQ`
- **Context-aware**: all MQ operations take a `context.Context`
- **Implemented by MQLite**: see [Message Queue](../container/msgqueue.md)
