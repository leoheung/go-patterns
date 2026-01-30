# Barrier

Synchronization primitive that allows multiple goroutines to wait for each other to reach a certain point.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/barrier"
```

## API Reference

### Create a Barrier

```go
// Create a new barrier for N goroutines
b := barrier.NewEasyBarrier(5)
```

### Done

```go
// Signal that a worker has completed
b.Done()
```

### Sync

```go
// Wait for all workers to complete
b.Sync()
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/parallel/barrier"
)

func main() {
    const numWorkers = 3
    b := barrier.NewEasyBarrier(numWorkers)

    for i := 0; i < numWorkers; i++ {
        go func(id int) {
            fmt.Printf("Worker %d: Phase 1\n", id)
            time.Sleep(time.Duration(id*100) * time.Millisecond)

            // Signal completion
            b.Done()
        }(i)
    }

    // Wait for all workers to complete
    b.Sync()
    fmt.Println("All workers reached barrier")
}
```

## Output

```
Worker 0: Phase 1
Worker 1: Phase 1
Worker 2: Phase 1
All workers reached barrier
```

## Features

- **Simple**: Easy to use synchronization mechanism
- **Thread-safe**: Safe for concurrent use
- **Channel-based**: Uses channels for synchronization
- **Single implementation**: EasyBarrier implementation
