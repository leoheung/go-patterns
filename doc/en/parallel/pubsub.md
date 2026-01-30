# PubSub

Publish-Subscribe pattern implementation.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/pubsub"
```

## API Reference

### Initialize PubSub System

```go
// Initialize the PubSub system
pubsub.InitPubSubSystem()

// Shutdown the PubSub system when done
// pubsub.Shutdown()
```

### Subscribe

```go
// Subscribe to a topic with buffer size
id, ch, err := pubsub.Subscribe("topic-name", 10)
if err != nil {
    // Handle error
}
```

### Publish

```go
// Publish a message to a topic
err := pubsub.Publish("topic-name", message)
if err != nil {
    // Handle error
}
```

### Unsubscribe

```go
// Unsubscribe from a topic
pubsub.Unsubscribe("topic-name", id)
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/parallel/pubsub"
)

func main() {
    // Initialize PubSub system
    pubsub.InitPubSubSystem()
    defer pubsub.Shutdown()

    // Subscriber 1
    id1, ch1, err := pubsub.Subscribe("news", 10)
    if err != nil {
        fmt.Printf("Error subscribing: %v\n", err)
        return
    }
    go func() {
        for msg := range ch1 {
            fmt.Printf("Subscriber 1 received: %v\n", msg)
        }
    }()

    // Subscriber 2
    id2, ch2, err := pubsub.Subscribe("news", 10)
    if err != nil {
        fmt.Printf("Error subscribing: %v\n", err)
        return
    }
    go func() {
        for msg := range ch2 {
            fmt.Printf("Subscriber 2 received: %v\n", msg)
        }
    }()

    // Publish messages
    time.Sleep(100 * time.Millisecond)

    err = pubsub.Publish("news", "Breaking: Go 1.22 released!")
    if err != nil {
        fmt.Printf("Error publishing: %v\n", err)
    }

    err = pubsub.Publish("news", "Breaking: New patterns added!")
    if err != nil {
        fmt.Printf("Error publishing: %v\n", err)
    }

    time.Sleep(100 * time.Millisecond)

    // Unsubscribe
    pubsub.Unsubscribe("news", id1)
    pubsub.Unsubscribe("news", id2)
}
```

## Output

```
Subscriber 1 received: Breaking: Go 1.22 released!
Subscriber 2 received: Breaking: Go 1.22 released!
Subscriber 1 received: Breaking: New patterns added!
Subscriber 2 received: Breaking: New patterns added!
```

## Features

- **Multiple subscribers**: One-to-many message distribution
- **Topic-based**: Organize messages by topic
- **Async delivery**: Non-blocking publish
- **System initialization**: Requires explicit initialization and shutdown
- **Buffer support**: Configurable buffer size for subscribers
- **Error handling**: Returns errors for invalid operations
