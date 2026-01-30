# Cache

A cache implementation for storing and retrieving data with TTL support.

## Installation

```go
import "github.com/leoheung/go-patterns/container/cache"
```

## API Reference

### Create a Cache

```go
// Create a new cache
c, err := cache.NewCache()
if err != nil {
    // Handle error
}
```

### Add

```go
// Add a value with a specific TTL
err := c.Add("key", "value", 5*time.Minute)
if err != nil {
    // Handle error
}
```

### Get

```go
// Get a value (returns nil if key doesn't exist or is expired)
value := c.Get("key")
if value != nil {
    // Use value
}
```

### Delete

```go
// Delete a specific key
c.Delete("key")
```

### Cache Status

```go
// Get cache status as string (includes item count and scheduler status)
status := c.String()
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/container/cache"
)

func main() {
    // Create a cache
    c, err := cache.NewCache()
    if err != nil {
        fmt.Printf("Error creating cache: %v\n", err)
        return
    }

    // Add values
    err = c.Add("user:1", "Alice", 5*time.Minute)
    if err != nil {
        fmt.Printf("Error adding user:1: %v\n", err)
    }

    err = c.Add("user:2", "Bob", 10*time.Minute)
    if err != nil {
        fmt.Printf("Error adding user:2: %v\n", err)
    }

    // Get value
    value := c.Get("user:1")
    if value != nil {
        fmt.Printf("User 1: %v\n", value)
    }

    // Delete a key
    c.Delete("user:1")

    // Print cache status
    fmt.Println("Cache status:")
    fmt.Println(c.String())
}
```

## Features

- **TTL Support**: Entries automatically expire and are cleaned up based on the specified duration.
- **Thread-safe**: Uses RWMutex internally, supporting high-concurrency reads.
- **Simple API**: Easy-to-use Add/Get interface.
- **Priority-based Scheduling**: Uses `PriorityScheduledTaskManager` for precise management of expiration tasks.
