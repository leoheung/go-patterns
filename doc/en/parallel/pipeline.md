# Pipeline

Pipeline patterns for data processing.

## Installation

```go
import "github.com/leoxiang66/go-patterns/parallel/pipeline"
```

## API Reference

### AddOnPipe

Generic pipeline node that transforms data from X to Y.

```go
// q: quit channel
// f: transformation function
// in: input channel
out := pipeline.AddOnPipe(q, f, in)
```

### FanIn

Merge multiple input channels into a single output channel.

```go
out := pipeline.FanIn(q, input1, input2, input3)
```

### FanOut

Distribute data from a single input channel to multiple output channels.

```go
outs := pipeline.FanOut(q, in, 3) // 3 output channels
```

### Broadcast

Broadcast data to multiple subscribers.

```go
broadcast := pipeline.NewBroadcast(q, in)
subscriber1 := broadcast.Subscribe()
subscriber2 := broadcast.Subscribe()
go broadcast.Run()
```

### Take

Take the first n elements from input channel.

```go
out := pipeline.Take(q, 5, in) // Take first 5 elements
```

## Example: Simple Pipeline

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/parallel/pipeline"
)

func main() {
    // Create channels
    input := make(chan int)
    quit := make(chan struct{})
    defer close(quit)
    
    // Create pipeline: Square -> Double
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
        fmt.Println(result) // Output: 2, 8, 18, 32, 50
    }
}
```

## Example: FanOut and FanIn

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/parallel/pipeline"
)

func main() {
    // Create channels
    input := make(chan int)
    quit := make(chan struct{})
    defer close(quit)
    
    // FanOut: Distribute data to 3 workers
    workers := pipeline.FanOut(quit, input, 3)
    
    // Process data in parallel
    process := func(x int) int { return x * 2 }
    var processed []chan int
    for _, worker := range workers {
        processed = append(processed, pipeline.AddOnPipe(quit, process, worker))
    }
    
    // FanIn: Merge results from all workers
    output := pipeline.FanIn(quit, processed...)
    
    // Send data
    go func() {
        for i := 1; i <= 5; i++ {
            input <- i
        }
        close(input)
    }()
    
    // Receive results
    for result := range output {
        fmt.Println(result)
    }
}
```

## Features

- **Composable**: Chain multiple pipeline stages
- **Type-safe**: Generic functions
- **Concurrent**: Leverages Go channels for concurrency
- **Cancellable**: Quit channel for graceful shutdown
