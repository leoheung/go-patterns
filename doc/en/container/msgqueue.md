# Message Queue

A channel-based message queue implementation with basic queue operations.

## Installation

```go
import "github.com/leoxiang66/go-patterns/container/msgQueue"
```

## API Reference

### Create a Message Queue

```go
// Create a new message queue with specified capacity and device ID
mq := msgqueue.NewChanMQ(100, "device-1")
```

### Enqueue

```go
// Enqueue a message
err := mq.Enq([]byte("hello world"))
```

### Dequeue

```go
// Dequeue a message with context
ctx := context.Background()
msg, err := mq.Deq(ctx)
```

### Queue Operations

```go
// Get current queue length
length := mq.Len()

// Clear all messages
err := mq.Clear()

// Check if queue is live
isLive := mq.IsLive()

// Renew a dead queue
mq.Renew()

// Destroy the queue
mq.Destroy()
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/leoxiang66/go-patterns/container/msgQueue"
)

func main() {
    // Create a message queue with capacity 10
    mq := msgqueue.NewChanMQ(10, "test-device")
    defer mq.Destroy()

    // Start a goroutine to consume messages
    go func() {
        ctx := context.Background()
        for i := 0; i < 5; i++ {
            msg, err := mq.Deq(ctx)
            if err != nil {
                fmt.Printf("Error dequeuing: %v\n", err)
                return
            }
            fmt.Printf("Received message: %s\n", string(msg))
            time.Sleep(500 * time.Millisecond)
        }
    }()

    // Enqueue messages
    for i := 0; i < 5; i++ {
        msg := fmt.Sprintf("Message %d", i)
        if err := mq.Enq([]byte(msg)); err != nil {
            fmt.Printf("Error enqueuing: %v\n", err)
            return
        }
        fmt.Printf("Sent message: %s\n", msg)
    }

    // Wait for all messages to be processed
    time.Sleep(3 * time.Second)
}
```

## Features

- **Channel-based**: Built on Go channels for efficient concurrent operations
- **Context support**: Supports cancellation via context
- **Lifecycle management**: Create, renew, and destroy queues
- **Thread-safe**: Safe for concurrent use by multiple goroutines
