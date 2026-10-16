# Limiter

Rate limiters used to control the frequency of operations. The package provides three implementations:

- **StaticLimiter** — fixed-interval token pacing
- **TokenBucket** — classic token bucket with context-aware blocking grants
- **SlidingWindowLimiter** — sliding window request counter

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

## Token Bucket

A classic token bucket rate limiter: tokens are added at a fixed interval, and every request must acquire a token to pass. The bucket overflows when full and rejects or blocks requests when empty.

### Create a Token Bucket

```go
// capacity: bucket size (max tokens)
// interval: token refill interval
tb := limiter.NewTokenBucket(100, 10*time.Millisecond)
```

The bucket starts full (`capacity` tokens) and refills one token per `interval`.

### TryGrant (non-blocking)

```go
ok, err := tb.TryGrant()
// ok == true  -> token acquired
// ok == false, err == nil  -> bucket empty
// err != nil  -> bucket closed
```

### Grant (blocking, context-aware)

```go
// Blocks until a token is available, ctx is done, or the bucket is closed
err := tb.Grant(ctx)
```

### Close

```go
tb.Close()  // stop the refiller; pending/future grants return an error
```

### Token Bucket Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/leoheung/go-patterns/parallel/limiter"
)

func main() {
    // Bucket of 10 tokens, refill every 100ms
    tb := limiter.NewTokenBucket(10, 100*time.Millisecond)
    defer tb.Close()

    for i := 0; i < 15; i++ {
        if ok, err := tb.TryGrant(); ok {
            fmt.Printf("request %d granted\n", i)
        } else {
            fmt.Printf("request %d rejected: %v\n", i, err)
        }
    }
}
```

### Token Bucket Features

- **Context-aware**: `Grant` respects context cancellation.
- **Non-blocking option**: `TryGrant` never blocks.
- **Burst support**: starts full and refills at a fixed rate.
- **Closable**: `Close` stops the refill goroutine and rejects further grants.

## Sliding Window Limiter

A sliding window request counter: allows up to `frequency_hz` requests per second, measured over a window split into `buckets_count` time buckets. This approximates a sliding window rate limit with low overhead.

### Create a Sliding Window Limiter

```go
// frequency_hz: max requests per second
// buckets_count: number of sub-buckets per second window
sw, err := limiter.NewSlidingWindowLimiter(100, 10)
if err != nil {
    // frequency_hz or buckets_count <= 0
}
```

### TryGrant

```go
ok, err := sw.TryGrant()
// ok == true  -> request allowed
// ok == false, err == nil  -> rate limit exceeded
// err != nil  -> limiter closed
```

`TryGrant` is non-blocking and runs directly in the calling goroutine.

### Close

```go
sw.Close()  // stop the cleanup goroutine; further TryGrant calls return an error
```

### Sliding Window Example

```go
package main

import (
    "fmt"

    "github.com/leoheung/go-patterns/parallel/limiter"
)

func main() {
    // 100 requests per second, 10 buckets of 100ms each
    sw, err := limiter.NewSlidingWindowLimiter(100, 10)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer sw.Close()

    for i := 0; i < 150; i++ {
        ok, err := sw.TryGrant()
        if err != nil {
            fmt.Println(err)
            return
        }
        if !ok {
            fmt.Printf("request %d throttled\n", i)
        }
    }
}
```

### Sliding Window Features

- **Approximate sliding window**: counts requests over the last `buckets_count` buckets.
- **Non-blocking**: `TryGrant` executes inline without waiting.
- **Atomic counting**: bucket counters are updated with `sync/atomic`.
- **Automatic cleanup**: expired buckets are removed periodically by a background goroutine.
