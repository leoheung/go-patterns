# Limiter

A static rate limiter used to control the frequency of operations.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/limiter"
```

## API Reference

### Create a Limiter

```go
// Create a limiter with a specific interval
// 100ms interval = 10 operations per second
lim := limiter.NewStaticLimiter(100 * time.Millisecond)
```

### Wait for Token

```go
// Blocking call: waits until the next token is available
lim.GrantNextToken()
```

### Control

```go
// Change the limiting interval at runtime
lim.Reset(200 * time.Millisecond)

// Stop the underlying ticker to release resources
lim.Stop()
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/parallel/limiter"
)

func main() {
    // Create a limiter: 1 operation per 200ms (5 ops/sec)
    lim := limiter.NewStaticLimiter(200 * time.Millisecond)
    defer lim.Stop()

    for i := 0; i < 5; i++ {
        start := time.Now()

        // Wait for token
        lim.GrantNextToken()

        // Perform operation
        fmt.Printf("Operation %d at %v (elapsed: %v)\n",
            i+1, time.Now().Format("15:04:05.000"),
            time.Since(start))
    }
}
```

## Features

- **Precise Timing**: Built on `time.Ticker` for accurate interval control.
- **Thread-safe**: Safe for concurrent access from multiple goroutines.
- **Dynamic Configuration**: Supports updating the rate on the fly using `Reset`.
- **Resource Management**: Includes a `Stop` method to clean up background tickers.
