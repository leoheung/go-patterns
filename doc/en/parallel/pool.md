# Worker Pool

Worker pool pattern for managing concurrent tasks.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/pool"
```

## API Reference

### Create a Worker Pool

```go
// Create a worker pool with specified number of workers
p := pool.NewWorkerPool(5) // 5 workers
```

### Submit Tasks

```go
// Submit a task to the pool
p.Submit(func() {
    // Task logic
})
```

### Stop the Pool

```go
// Stop the pool gracefully
p.Stop()
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

    // Stop the pool
    p.Stop()

    fmt.Printf("Total completed: %d\n", counter)
}
```

## Output

```
Task 0 started
Task 1 started
Task 2 started
Task 0 completed
Task 3 started
Task 1 completed
Task 4 started
...
Total completed: 10
```

## Features

- **Fixed-size pool**: Controls concurrency level
- **Task queue**: Buffered channel for pending tasks
- **Graceful shutdown**: Waits for running tasks to complete
