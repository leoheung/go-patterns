# Worker Pool

A lightweight worker pool implementation based on semaphores to control concurrency.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/pool"
```

## API Reference

### Create a Worker Pool

```go
// Create a worker pool with a fixed number of workers
p := pool.NewWorkerPool(5)
```

### Submit Tasks

```go
// Blocking submit: waits for an available worker
p.Submit(func() {
    // Task logic
})

// Non-blocking submit: returns false immediately if no worker is available
success := p.TrySubmit(func() {
    // Task logic
})
```

## Complete Example

```go
package main

import (
    "fmt"
    "sync/atomic"
    "time"
    "github.com/leoheung/go-patterns/parallel/pool"
)

func main() {
    // Create a pool with 3 workers
    p := pool.NewWorkerPool(3)

    var counter int32 = 0

    // Submit 10 tasks
    for i := 0; i < 10; i++ {
        taskID := i
        p.Submit(func() {
            fmt.Printf("Task %d started\n", taskID)
            time.Sleep(100 * time.Millisecond)
            atomic.AddInt32(&counter, 1)
            fmt.Printf("Task %d completed\n", taskID)
        })
    }

    // Wait for tasks to complete
    time.Sleep(1 * time.Second)

    fmt.Printf("Total completed: %d\n", counter)
}
```

## Features

- **Concurrency Control**: Limits the number of goroutines running simultaneously.
- **Panic Protection**: Automatically recovers from panics within tasks and logs the error, preventing the entire pool from crashing.
- **Flexible Submission**: Supports both blocking (`Submit`) and non-blocking (`TrySubmit`) task dispatching.
- **Resource Efficient**: Built on top of `SemaphoreByCond` for minimal overhead.
