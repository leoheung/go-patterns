# Type Utilities

## Overview

The `utils` package provides `IsInstanceOf`, a generic helper for checking whether a value implements or is of a specific type. It simplifies the verbose type assertion syntax in Go, especially when working with `any` values.

## Installation

```go
import "github.com/leoheung/go-patterns/utils"
```

## API Reference

### `IsInstanceOf[T any](obj any) bool`

Returns `true` if `obj` is of type `T` (or implements `T` when `T` is an interface), otherwise `false`. The type to check is specified as the generic parameter.

```go
var obj any = "hello"

ok := utils.IsInstanceOf[string](obj) // true
ok = utils.IsInstanceOf[int](obj)     // false
```

## Example

```go
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/leoheung/go-patterns/utils"
)

func describe(obj any) {
	switch {
	case utils.IsInstanceOf[string](obj):
		fmt.Println("It is a string:", obj)
	case utils.IsInstanceOf[int](obj):
		fmt.Println("It is an int:", obj)
	case utils.IsInstanceOf[io.Reader](obj):
		fmt.Println("It is an io.Reader")
	default:
		fmt.Println("Unknown type")
	}
}

func main() {
	describe("hello")        // It is a string: hello
	describe(42)             // It is an int: 42
	describe(os.Stdin)       // It is an io.Reader
}
```

## Features

- **Generic Type Parameter**: Pass the target type directly as `[T]`, no reflection boilerplate
- **Interface Support**: Works with interface types (e.g. `IsInstanceOf[io.Reader](obj)`)
- **Non-panicking**: Uses the comma-ok form of type assertion internally, never panics on mismatch

## Notes

- Equivalent to `_, ok := obj.(T)` with the type supplied via a generic parameter
- For pointer types, specify the pointer explicitly (e.g. `IsInstanceOf[*os.File](obj)`)
- A `nil` `any` value always returns `false`
