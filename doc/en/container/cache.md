# Cache

A cache implementation for storing and retrieving data with TTL support.

## Installation

```go
import "github.com/leoxiang66/go-patterns/container/cache"
```

## API Reference

### Create a Cache

```go
// Create a new cache with default settings
c := cache.NewCache()

// Create a cache with custom TTL
c := cache.NewCacheWithTTL(5 * time.Minute)
```

### Set

```go
// Set a value with default TTL
c.Set("key", "value")

// Set a value with custom TTL
c.SetWithTTL("key", "value", 10*time.Minute)
```

### Get

```go
// Get a value
if value, ok := c.Get("key"); ok {
    // Use value
}
```

### Delete

```go
// Delete a key
c.Delete("key")

// Clear all entries
c.Clear()
```

### Cache Operations

```go
// Check if key exists
exists := c.Has("key")

// Get cache size
size := c.Len()
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoxiang66/go-patterns/container/cache"
)

func main() {
    // Create a cache with 5 minute TTL
    c := cache.NewCacheWithTTL(5 * time.Minute)

    // Set values
    c.Set("user:1", "Alice")
    c.Set("user:2", "Bob")

    // Get value
    if user, ok := c.Get("user:1"); ok {
        fmt.Printf("User 1: %s\n", user)
    }

    // Set with custom TTL
    c.SetWithTTL("session:abc", "active", 30*time.Minute)

    // Check existence
    if c.Has("user:2") {
        fmt.Println("User 2 exists")
    }

    // Get cache size
    fmt.Printf("Cache size: %d\n", c.Len())

    // Delete a key
    c.Delete("user:1")

    // Clear all
    c.Clear()
}
```

## Features

- **TTL support**: Automatic expiration of entries
- **Thread-safe**: Safe for concurrent use
- **Simple API**: Easy to use Get/Set interface
- **Memory efficient**: Automatic cleanup of expired entries
