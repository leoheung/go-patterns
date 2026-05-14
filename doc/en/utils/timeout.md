# Timeout

## Overview

The `utils` package provides a timeout mechanism for executing functions that may hang or take too long to complete.

## API Reference

### `TimeoutWork(work func() (any, error), timeout time.Duration) (any, error)`

Executes a work function with a timeout. If the function doesn't complete within the specified duration, it returns a timeout error.

**Parameters:**
- `work`: The function to execute
- `timeout`: Maximum duration to wait for completion

**Returns:**
- The data returned by the work function
- A timeout error if the function didn't complete in time
- Any error returned by the work function

```go
func TimeoutWork(work func() (any, error), timeout time.Duration) (any, error)
```

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// Example 1: Function that completes in time
	fastWork := func() (any, error) {
		time.Sleep(100 * time.Millisecond)
		return "Task completed", nil
	}

	result, err := utils.TimeoutWork(fastWork, 1*time.Second)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}

	// Example 2: Function that times out
	slowWork := func() (any, error) {
		time.Sleep(5 * time.Second) // Simulating slow operation
		return "Task completed", nil
	}

	result, err = utils.TimeoutWork(slowWork, 1*time.Second)
	if err != nil {
		fmt.Printf("Timeout error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}
}
```

## Behavior

1. Starts a timer for the specified timeout duration
2. Executes the work function in a separate goroutine
3. Waits for either:
   - The work function to complete → returns the result
   - The timeout to expire → returns timeout error
4. Automatically recovers from panics in the work function

## Notes

- The work function always runs to completion in its goroutine, even if the timeout expires
- Panic recovery is built-in to prevent crashes
- The timeout is precise and doesn't block the calling goroutine
- Suitable for network requests, database queries, and other potentially slow operations