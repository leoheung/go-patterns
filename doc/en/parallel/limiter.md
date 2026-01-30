# Limiter

A static limiter for controlling the rate of operations.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/limiter"
```

## API Reference

### Create a Limiter

```go
// Create a new limiter with specified interval
// 100ms interval = 10 operations per second
lim := limiter.NewStaticLimiter(100 * time.Millisecond)
```

### Wait for Token

```go
// Wait until next token is available
lim.GrantNextToken()
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

## Output

```
Operation 1 at 14:30:00.000 (elapsed: 0s)
Operation 2 at 14:30:00.200 (elapsed: 200ms)
Operation 3 at 14:30:00.400 (elapsed: 200ms)
Operation 4 at 14:30:00.600 (elapsed: 200ms)
Operation 5 at 14:30:00.800 (elapsed: 200ms)
```

## Use Cases

- API rate limiting
- Resource throttling
- Preventing overwhelming external services

## Features

- **Static rate**: Fixed interval between operations
- **Blocking**: Blocks until token is available
- **Simple**: Easy to use with defer
