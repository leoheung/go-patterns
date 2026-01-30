# Mutex

A simple mutual exclusion lock implementation.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/mutex"
```

## API Reference

### Create a Mutex

```go
// Create a new mutex
m := mutex.NewMutex()
```

### Lock and Unlock

```go
// Lock the mutex
m.Lock()

// Unlock the mutex (use defer for safety)
defer m.Unlock()
```

## Complete Example

```go
package main

import (
    "fmt"
    "sync"
    "github.com/leoheung/go-patterns/parallel/mutex"
)

func main() {
    m := mutex.NewMutex()
    counter := 0
    var wg sync.WaitGroup

    // Launch 10 goroutines
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()

            // Lock before accessing shared resource
            m.Lock()
            defer m.Unlock()

            // Critical section
            counter++
            fmt.Printf("Counter: %d\n", counter)
        }()
    }

    wg.Wait()
    fmt.Printf("Final counter: %d\n", counter)
}
```

## Output

```
Counter: 1
Counter: 2
Counter: 3
Counter: 4
Counter: 5
Counter: 6
Counter: 7
Counter: 8
Counter: 9
Counter: 10
Final counter: 10
```

## Features

- **Channel-based**: Implemented using Go channels
- **Simple API**: Just Lock and Unlock
- **FIFO ordering**: Fair lock acquisition
