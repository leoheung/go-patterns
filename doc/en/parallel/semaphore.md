# Semaphore

A semaphore implementation for limiting concurrent access to resources.

## Installation

```go
import "github.com/leoxiang66/go-patterns/parallel/semaphore"
```

## API Reference

### Create a Semaphore

```go
// Create a new semaphore with specified capacity
sem := semaphore.NewSemaphore(5) // 5 permits
```

### Acquire and Release

```go
// Acquire a permit
sem.Acquire()
defer sem.Release()
```

## Complete Example

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/leoxiang66/go-patterns/parallel/semaphore"
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

## Output

```
Goroutine 0: Working...
Goroutine 1: Working...
Goroutine 2: Working...
Goroutine 0: Done
Goroutine 3: Working...
...
```

## Semaphore with Condition Variable

Alternative implementation using condition variables:

```go
import "github.com/leoxiang66/go-patterns/parallel/semaphore"

// Create semaphore with condition variable
sem := semaphore.NewSemaphoreByCond(5)
sem.Acquire()
sem.Release()
```

## Features

- **Resource limiting**: Control concurrent access
- **Two implementations**: Channel-based and condition variable-based
- **Simple API**: Acquire and Release
