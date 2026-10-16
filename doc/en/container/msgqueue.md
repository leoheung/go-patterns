# Message Queue

::: tip
The MQ implementation lives in `parallel/communication/mq` (not in `container/`). This page is kept at its original location to avoid breaking links.
:::

A lightweight topic-based message queue (`MQLite`) that delivers messages to subscribers over HTTP callbacks. It builds on `parallel/stream`, `parallel/pool` and `container/safemap`, and manages its whole lifecycle through an internal `root_ctx`.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/communication/mq"
```

## API Reference

### Message Structure

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

`Message.ID` is assigned automatically by `Publish` (monotonically increasing counter).

### Create a Message Queue

```go
// Create a new MQ with message buffer size, worker count,
// job queue size and shard count
q := mq.NewMQLite(100, 4, 64, 16)
```

All parameters must be `>= 0`, otherwise `nil` is returned.

### Publish

```go
// Publish a message to a topic. Returns the assigned message ID.
id, err := q.Publish(ctx, &mq.Message{
    Topic:   "orders",
    Payload: []byte(`{"orderId":"42"}`),
})
```

- The topic must have at least one subscriber, otherwise an error is returned.
- The send is non-blocking: if the message buffer is full, an error is returned immediately.

### Subscribe / Unsubscribe

```go
// Register a subscriber with an HTTP callback URL and a max retry count
err := q.Subscribe(ctx, "orders", "sub-1", "http://localhost:8080/callback", 3)

// Remove a subscriber from a topic
err = q.Unsubscribe(ctx, "orders", "sub-1")
```

Delivery is push-based: each message is POSTed as JSON to the subscriber's `callbackURL`.

### Broadcast & Send

```go
// Deliver a message to every subscriber of msg.Topic
err := q.Broadcast(ctx, msg)
```

`Send(ctx, sub, msg)` posts a single message to one subscriber and is used internally by `Broadcast`. Both are part of the `MQ` interface.

### Close

```go
// Cancel the internal root_ctx to shut the queue down
err := q.Close(ctx)
```

## Complete Example

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

    // Register a subscriber
    err := q.Subscribe(context.Background(), "orders", "sub-1",
        "http://localhost:8080/callback", 3)
    if err != nil {
        fmt.Printf("subscribe error: %v\n", err)
        return
    }

    // Publish a message
    id, err := q.Publish(context.Background(), &mq.Message{
        Topic:   "orders",
        Payload: []byte(`{"orderId":"42"}`),
        Headers: mq.MessageHeader{
            EventType: "order.created",
            Timestamp: time.Now(),
        },
    })
    if err != nil {
        fmt.Printf("publish error: %v\n", err)
        return
    }
    fmt.Printf("published message %s\n", *id)
}
```

## Features

- **Topic-based**: subscribers register per topic with an HTTP callback URL
- **HTTP delivery**: messages are POSTed as JSON to subscriber endpoints
- **Automatic retry**: failed deliveries are retried up to `maxInflight` times with increasing backoff
- **Non-blocking publish**: full buffers return an error instead of blocking the producer
- **Root context lifecycle**: `Close` cancels the internal `root_ctx`, stopping all pipelines
- **Concurrent delivery**: broadcasts are executed on a `parallel/pool.AsyncPoolV2` worker pool
