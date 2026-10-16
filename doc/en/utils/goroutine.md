# Goroutine Utilities

## Overview

The `utils` package provides `GetGoid` for retrieving the current goroutine's ID. Since the Go runtime does not officially expose goroutine IDs, this function parses them from the stack trace output. It is intended for debugging and logging purposes only.

## Installation

```go
import "github.com/leoheung/go-patterns/utils"
```

## API Reference

### `GetGoid() uint64`

Returns the ID of the current goroutine by parsing `runtime.Stack` output. Returns `0` if the stack output cannot be parsed.

```go
id := utils.GetGoid()
fmt.Printf("Running on goroutine %d\n", id)
```

## Example

```go
package main

import (
	"fmt"
	"sync"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	fmt.Println("main goroutine:", utils.GetGoid())

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			// Each goroutine prints its own ID
			fmt.Printf("worker %d runs on goroutine %d\n", n, utils.GetGoid())
		}(i)
	}
	wg.Wait()
}
```

## Features

- **Zero Allocation Parsing**: Reuses a fixed 64-byte buffer to read the stack output
- **Robust Parsing**: Validates the `"goroutine "` prefix and field format before parsing
- **Safe Fallback**: Returns `0` instead of panicking when the output cannot be parsed

## Notes

- Go officially discourages reliance on goroutine IDs; use this function for debugging and logging only
- Goroutine IDs may be reused after a goroutine exits
- The stack buffer is limited to 64 bytes, which is sufficient for the `"goroutine N [status]:"` header line
