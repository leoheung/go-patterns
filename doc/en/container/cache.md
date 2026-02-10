# Cache

A cache implementation for storing and retrieving data with TTL support and permanent caching capabilities.

## Installation

```go
import "github.com/leoheung/go-patterns/container/cache"
import "github.com/leoheung/go-patterns/net" // For pointer helper functions
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
// Add a value with a specific TTL (requires *time.Duration)
duration := 5 * time.Minute
err := c.Add("key", "value", &duration)

// Add a permanent value (pass nil)
err := c.Add("permanent_key", "value", nil)
```

### Get

```go
// Get a value (returns nil if key doesn't exist or is expired)
// For non-permanent items, Get automatically resets the expiration time
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

    // 1. Add value with expiration
    duration := 5 * time.Minute
    err = c.Add("user:1", "Alice", &duration)
    if err != nil {
        fmt.Printf("Error adding user:1: %v\n", err)
    }

    // 2. Add permanent value (nil duration)
    err = c.Add("config:version", "v1.0.0", nil)
    if err != nil {
        fmt.Printf("Error adding config: %v\n", err)
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

- **Flexible Expiration**: Supports both automatic expiration (TTL) and permanent caching (`nil` duration).
- **Auto-Renewal**: Automatically resets the expiration timer on every `Get` access for items with TTL.
- **Thread-safe**: Uses RWMutex internally, supporting high-concurrency reads.
- **Priority-based Scheduling**: Uses `PriorityScheduledTaskManager` for precise management of expiration tasks.
