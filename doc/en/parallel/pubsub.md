# PubSub

Publish-Subscribe pattern implementation.

## Installation

```go
import "github.com/leoxiang66/go-patterns/parallel/pubsub"
```

## API Reference

### Create a PubSub

```go
// Create a new PubSub instance
ps := pubsub.NewPubSub()
```

### Subscribe

```go
// Subscribe to a topic
ch := ps.Subscribe("topic-name")
```

### Publish

```go
// Publish a message to a topic
ps.Publish("topic-name", message)
```

### Unsubscribe

```go
// Unsubscribe from a topic
ps.Unsubscribe("topic-name", ch)
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoxiang66/go-patterns/parallel/pubsub"
)

func main() {
    ps := pubsub.NewPubSub()
    
    // Subscriber 1
    ch1 := ps.Subscribe("news")
    go func() {
        for msg := range ch1 {
            fmt.Printf("Subscriber 1 received: %v\n", msg)
        }
    }()
    
    // Subscriber 2
    ch2 := ps.Subscribe("news")
    go func() {
        for msg := range ch2 {
            fmt.Printf("Subscriber 2 received: %v\n", msg)
        }
    }()
    
    // Publish messages
    time.Sleep(100 * time.Millisecond)
    ps.Publish("news", "Breaking: Go 1.22 released!")
    ps.Publish("news", "Breaking: New patterns added!")
    
    time.Sleep(100 * time.Millisecond)
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
