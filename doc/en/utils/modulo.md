# Modulo Utilities

## Overview

The `utils` package provides `ModEuclid`, a Euclidean modulo operation. Unlike Go's built-in `%` operator (which is a truncated remainder and takes the sign of the dividend), `ModEuclid` always returns a non-negative result when `base` is positive. This is useful for cyclic indexing, hash bucketing, and wrapping negative numbers.

## Installation

```go
import "github.com/leoheung/go-patterns/utils"
```

## API Reference

### `ModEuclid(num, base int) int`

Returns the Euclidean modulo of `num` by `base`. If the built-in remainder is negative, `base` is added to make the result non-negative (for positive `base`).

```go
r := utils.ModEuclid(-1, 5) // 4  (built-in -1 % 5 would be -1)
r = utils.ModEuclid(7, 5)   // 2
```

## Example

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// Wrap a possibly negative index into a valid range
	items := []string{"a", "b", "c", "d", "e"}
	for _, i := range []int{-1, -6, 3, 8} {
		idx := utils.ModEuclid(i, len(items))
		fmt.Printf("items[ModEuclid(%d, 5)] = %s\n", i, items[idx])
	}
	// Output:
	// items[ModEuclid(-1, 5)] = e
	// items[ModEuclid(-6, 5)] = e
	// items[ModEuclid(3, 5)] = d
	// items[ModEuclid(8, 5)] = d

	// Compare with the built-in operator
	fmt.Println(-1 % 5)              // -1
	fmt.Println(utils.ModEuclid(-1, 5)) // 4
}
```

## Features

- **Always Non-negative Result**: For positive `base`, the result is always in `[0, base)`
- **One Extra Branch**: Only adds `base` when the built-in remainder is negative, minimal overhead
- **Cyclic-safe**: Ideal for wrapping array indices or cyclic counters across zero

## Notes

- `base` must not be `0` (same restriction as the built-in `%` operator)
- For positive `num` and `base`, `ModEuclid` behaves identically to `num % base`
- The result convention follows the Euclidean definition: `0 <= result < base` when `base > 0`
