# Context Merge

## Overview

The `utils` package provides `MergeContexts` for merging two `context.Context` values into one. The merged context is canceled when either of the original contexts is canceled, or when the returned `CancelFunc` is called. This is useful in scenarios where an operation must respect multiple cancellation sources (e.g. a request context and a server shutdown context).

## Installation

```go
import "github.com/leoheung/go-patterns/utils"
```

## API Reference

### `MergeContexts(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc)`

Merges `ctx1` and `ctx2` into a single context. The returned context is derived from `ctx1` and is canceled as soon as either `ctx1` or `ctx2` is done. Returns the merged context and a `CancelFunc` that stops the internal hook and cancels the merged context.

If either input context is already canceled, a context that is already canceled is returned immediately (fast path).

```go
ctx1, cancel1 := context.WithCancel(context.Background())
defer cancel1()

ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
defer cancel2()

merged, cancel := utils.MergeContexts(ctx1, ctx2)
defer cancel()

// merged is done when ctx1 is canceled, ctx2 times out, or cancel() is called
<-merged.Done()
```

## Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// Request context (e.g. canceled when the client disconnects)
	reqCtx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()

	// Server context with a deadline (e.g. shutdown timeout)
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelTimeout()

	// The merged context is done as soon as either one is done
	merged, cancel := utils.MergeContexts(reqCtx, timeoutCtx)
	defer cancel()

	select {
	case <-merged.Done():
		fmt.Println("Merged context done:", merged.Err())
	case <-time.After(time.Second):
		fmt.Println("Work completed")
	}
	// Output (after ~100ms):
	// Merged context done: context deadline exceeded
}
```

## Features

- **Bidirectional Cancellation**: The merged context is canceled when either input context is canceled
- **Fast Path**: If either input is already canceled, an already-canceled context is returned immediately
- **Hook-based Propagation**: Uses `context.AfterFunc` internally to propagate `ctx2`'s cancellation without spawning extra goroutines
- **Clean Shutdown**: The returned `CancelFunc` stops the internal hook and cancels the merged context, preventing resource leaks

## Notes

- Always call the returned `CancelFunc` (typically via `defer`) to release resources
- Values are inherited from `ctx1` only (the merged context is derived from `ctx1`)
- Requires Go 1.21+ (`context.AfterFunc`)
