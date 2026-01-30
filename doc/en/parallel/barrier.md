# Barrier

Synchronization primitive that allows multiple goroutines to wait for each other to reach a certain point.

## Installation

```go
import "github.com/leoxiang66/go-patterns/parallel/barrier"
```

## API Reference

### Create a Barrier

```go
// Create a new barrier for N goroutines
b := barrier.NewBarrier(5)
```

### Wait

```go
// Wait for all goroutines to reach this point
b.Wait()
```

## Complete Example

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/leoxiang66/go-patterns/parallel/barrier"
)

func main() {
    const numWorkers = 3
    b := barrier.NewBarrier(numWorkers)
    var wg sync.WaitGroup

    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            fmt.Printf("Worker %d: Phase 1\n", id)
            time.Sleep(time.Duration(id*100) * time.Millisecond)
            
            // Wait for all workers to reach this point
            b.Wait()
            
            fmt.Printf("Worker %d: Phase 2 (all workers reached barrier)\n", id)
        }(i)
    }

    wg.Wait()
}
```

## Output

```
Worker 0: Phase 1
Worker 1: Phase 1
Worker 2: Phase 1
Worker 0: Phase 2 (all workers reached barrier)
Worker 1: Phase 2 (all workers reached barrier)
Worker 2: Phase 2 (all workers reached barrier)
```

## Barrier with Condition Variable

An alternative implementation using condition variables:

```go
import "github.com/leoxiang66/go-patterns/parallel/barrier"

// Create barrier with condition variable
b := barrier.NewBarrierWithCond(5)
b.Wait()
```

## Features

- **Cyclic**: Can be reused after all goroutines pass through
- **Thread-safe**: Safe for concurrent use
- **Two implementations**: Channel-based and condition variable-based
