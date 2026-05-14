# Common Helpers

## Overview

The `utils` package provides common helper functions for type checking, string manipulation, and goroutine control.

## API Reference

### `IsNil[T any](v T) bool`

Checks if a value is nil. Supports interface, pointer, slice, map, channel, function, and chan types.

```go
var s *string = nil
if utils.IsNil(s) {
    fmt.Println("Value is nil")
}

// Works with generics
var arr []int = nil
utils.IsNil(arr) // true
```

### `IsDigits(s string) bool`

Checks if a string consists entirely of digits (0-9).

```go
utils.IsDigits("12345") // true
utils.IsDigits("12a45") // false
utils.IsDigits("")      // true (empty string)
```

### `DelayDo(d time.Duration, fn func())`

Executes a function after a delay. This is a blocking operation.

```go
utils.DelayDo(500*time.Millisecond, func() {
    fmt.Println("Delayed execution")
})
```

### `PPrint(obj interface{})`

Prints an object in a formatted way. Alias for `PPrettyPrint`.

```go
utils.PPrint(myStruct)
```

### `Hold()`

Blocks the current goroutine indefinitely. Use for debugging or keeping the program running.

```go
utils.Hold() // Blocks forever
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
	// Check for nil values
	var ptr *int = nil
	fmt.Printf("ptr is nil: %v\n", utils.IsNil(ptr))

	var slice []string = nil
	fmt.Printf("slice is nil: %v\n", utils.IsNil(slice))

	// Check if string is all digits
	fmt.Printf("\"123\" is digits: %v\n", utils.IsDigits("123"))
	fmt.Printf("\"12a\" is digits: %v\n", utils.IsDigits("12a"))

	// Delayed execution
	fmt.Println("Starting delayed task...")
	utils.DelayDo(1*time.Second, func() {
		fmt.Println("Delayed task executed!")
	})
	fmt.Println("After delay")
}
```

## Notes

- `IsNil` uses reflection and handles various types correctly
- `IsDigits` returns `true` for empty strings
- `DelayDo` blocks until the delay completes
- `Hold()` creates a deadlock intentionally, use only for debugging