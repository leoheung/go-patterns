# Number Utilities

## Overview

The `utils` package provides a `Number` type that simulates TypeScript/JavaScript's `number` type, handling both integers and floating-point numbers uniformly using `float64` internally.

## Number Type

```go
type Number struct {
    value float64
}
```

## API Reference

### `ParseNumber(raw string) (*Number, error)`

Parses a string into a `*Number`. Handles both integer ("100") and floating-point ("100.5") formats.

```go
num, err := utils.ParseNumber("42")
if err != nil {
    log.Fatal(err)
}
fmt.Println(num.Float()) // 42
```

### `(n *Number) Float() float64`

Returns the underlying `float64` value.

```go
num, _ := utils.ParseNumber("123.456")
fmt.Println(num.Float()) // 123.456
```

### `(n *Number) Int() int`

Returns the integer part of the number (truncates the decimal part).

```go
num, _ := utils.ParseNumber("99.9")
fmt.Println(num.Int()) // 99
```

### `(n *Number) Int64() int64`

Returns the integer part as `int64`.

```go
num, _ := utils.ParseNumber("9999999999.5")
fmt.Println(num.Int64()) // 9999999999
```

### `(n *Number) IsInteger() bool`

Checks if the number is an integer (has no decimal part).

```go
num1, _ := utils.ParseNumber("10.0")
num2, _ := utils.ParseNumber("10.5")

fmt.Println(num1.IsInteger()) // true
fmt.Println(num2.IsInteger()) // false
```

## Example

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// Parse various number formats
	examples := []string{"42", "3.14", "100.0", "99.99", "-5"}

	for _, s := range examples {
		num, err := utils.ParseNumber(s)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", s, err)
			continue
		}

		fmt.Printf("Input: %s\n", s)
		fmt.Printf("  Float: %f\n", num.Float())
		fmt.Printf("  Int: %d\n", num.Int())
		fmt.Printf("  Int64: %d\n", num.Int64())
		fmt.Printf("  IsInteger: %v\n", num.IsInteger())
		fmt.Println()
	}
}
```

## Notes

- Internally uses `float64` to match JavaScript/TypeScript behavior
- Integer conversion truncates the decimal part (doesn't round)
- `IsInteger()` compares with `math.Trunc()` for precision
- String parsing uses `strconv.ParseFloat` with 64-bit precision