# Semaphore

A semaphore implementation for limiting concurrent access to resources.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/semaphore"
```

## API Reference

### Create a Semaphore

```go
// Create a new semaphore with specified capacity (channel-based)
sem := semaphore.NewSemaphore(5)

// Create a new semaphore with specified capacity (condition variable-based)
semCond := semaphore.NewSemaphoreByCond(5)
```

### Acquire and Release

```go
// Acquire a permit (blocking)
sem.Acquire()

// Try to acquire a permit (non-blocking, returns false if no permits available)
success := sem.TryAcquire()

// Release a permit
sem.Release()
```

## Complete Example

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/leoheung/go-patterns/parallel/semaphore"
)

func main() {
    // Create a semaphore with 3 permits
    sem := semaphore.NewSemaphore(3)
    var wg sync.WaitGroup

    // Launch 10 goroutines
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            // Acquire permit
            sem.Acquire()
            defer sem.Release()

            fmt.Printf("Goroutine %d: Working...\n", id)
            time.Sleep(200 * time.Millisecond)
            fmt.Printf("Goroutine %d: Done\n", id)
        }(i)
    }

    wg.Wait()
}
```

## Features

- **Resource Limiting**: Controls the number of concurrent operations.
- **Two Implementations**:
  - `Semaphore`: Channel-based, simple and idiomatic Go.
  - `SemaphoreByCond`: Mutex and Condition Variable-based, useful for specific synchronization needs.
- **Non-blocking Support**: `TryAcquire` allows checking for resource availability without waiting.
