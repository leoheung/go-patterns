# Go-Patterns

## Introduction

`go-patterns` is a collection of concurrency patterns and data structures implemented in Go, designed to help developers better understand and utilize Go's concurrency features.

- [doc](https://leoheung.github.io/go-patterns/en/)

## Installation

```bash
go get github.com/leoxiang66/go-patterns
```

## Quick Start

### Using List

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/container/list"
)

func main() {
    // Create a new list
    l := list.New[int]()
    
    // Add elements
    l.Append(1, 2, 3)
    l.Push(4)
    
    // Get element
    elem := l.Get(0)
    fmt.Println(elem) // Output: 1
    
    // Iterate
    for i := 0; i < l.Len(); i++ {
        fmt.Println(l.Get(i))
    }
}
```

### Using Pipeline

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/parallel/pipeline"
)

func main() {
    input := make(chan int)
    quit := make(chan struct{})
    defer close(quit)
    
    // Create pipeline stages
    square := func(x int) int { return x * x }
    double := func(x int) int { return x * 2 }
    
    stage1 := pipeline.AddOnPipe(quit, square, input)
    stage2 := pipeline.AddOnPipe(quit, double, stage1)
    
    // Send data
    go func() {
        for i := 1; i <= 5; i++ {
            input <- i
        }
        close(input)
    }()
    
    // Receive results
    for result := range stage2 {
        fmt.Println(result)
    }
}
```

### Using Semaphore

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/parallel/semaphore"
)

func main() {
    // Create a semaphore with capacity 3
    sem := semaphore.NewSemaphore(3)
    
    // Acquire and release
    sem.Acquire()
    defer sem.Release()
    
    // Perform operation
    fmt.Println("Performing operation...")
}
```

## Project Structure

```
go-patterns/
├── container/      # Data structures
│   ├── list/       # Generic dynamic array
│   ├── msgQueue/   # Channel-based message queue
│   ├── pq/         # Priority queue
│   └── cache/      # Cache implementation
├── parallel/       # Concurrency patterns
│   ├── barrier/    # Synchronization barrier
│   ├── limiter/    # Rate limiter
│   ├── mutex/      # Mutual exclusion lock
│   ├── pipeline/   # Pipeline patterns
│   ├── pool/       # Worker pool
│   ├── pubsub/     # Pub/Sub pattern
│   ├── rwlock/     # Read-write lock
│   └── semaphore/  # Semaphore
├── utils/          # Utility functions
├── cryptography/   # Cryptographic utilities
└── net/            # Network utilities
```

## Next Steps

- Explore the [Container](/en/container/) data structures
- Learn about [Parallel](/en/parallel/) concurrency patterns
- Check out [Utils](/en/utils/) for helper functions
