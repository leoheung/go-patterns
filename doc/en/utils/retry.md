# Retry

## Overview

The `utils` package provides a retry mechanism for executing functions that may fail, with automatic panic recovery and configurable retry attempts.

## API Reference

### `RetryWork(work func() (any, error), retryTimes int) (any, error)`

Executes a work function, catching panics or errors, and retries up to `retryTimes` times.

**Parameters:**
- `work`: The function to execute
- `retryTimes`: Maximum number of retries (not including the first attempt)

**Returns:**
- The data returned by the work function
- Any error encountered (returns after exhausting all retries)

```go
func RetryWork(work func() (any, error), retryTimes int) (any, error)
```

## Example

```go
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// Simulate a function that may fail
	attempt := 0
	work := func() (any, error) {
		attempt++
		// Simulate random failure (70% chance of failure)
		if rand.Float32() < 0.7 {
			return nil, fmt.Errorf("operation failed on attempt %d", attempt)
		}
		return fmt.Sprintf("Success on attempt %d", attempt), nil
	}

	// Retry up to 3 times (4 total attempts)
	result, err := utils.RetryWork(work, 3)

	if err != nil {
		fmt.Printf("Final error: %v\n", err)
	} else {
		fmt.Printf("Result: %v\n", result)
	}
}
```

## Retry Behavior

1. Executes the work function
2. If successful, returns immediately
3. If failed (error or panic):
   - Logs the failure
   - If retries remaining, sleeps 500ms and retries
   - If no retries remaining, returns the error
4. Automatically recovers from panics and treats them as errors

## Notes

- Panics are automatically recovered and converted to errors
- A 500ms delay is inserted between retries
- All attempts (including failures) are logged
- The function is thread-safe when used with proper synchronization