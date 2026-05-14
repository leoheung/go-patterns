# Utils

The `utils` package provides a collection of utility functions for common operations, covering logging, retry mechanisms, timeout control, object beautification, and numeric processing.

## Modules

### [Logging](./log.md)
Automatically identifies development or production environments and adopts different logging methods.

### [Retry](./retry.md)
Provides protection mechanisms with retry capabilities for unstable operations.

### [Timeout](./timeout.md)
Adds timeout limits to operations.

### [Pretty Print](./pretty.md)
Advanced beautification tools based on `go-spew`.

### [Color](./color.md)
Colored terminal output using ANSI escape sequences.

### [Number](./number.md)
Simulates numeric processing found in dynamic languages.

### [Common](./common.md)
Common helper functions.

### [Channel](./channel.md)
Non-blocking and timeout-based channel operations.

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/utils"
)

func main() {
    // 1. Formatted output
    user := struct {
        Name string
        Age  int
    }{"Leon", 25}

    fmt.Println("User data:")
    utils.PPrint(user)

    // 2. Retry logic
    count := 0
    utils.RetryWork(func() (any, error) {
        count++
        if count < 2 {
            return nil, fmt.Errorf("temporary error")
        }
        return "Success", nil
    }, 3)
}
```

## Features

- **Robustness**: Both retry and timeout functions include `recover()` internally to prevent business logic panics from crashing the program.
- **Ease of Use**: Simplifies tedious type pointer conversions and reflection checks in Go.
- **Debug Friendly**: Provides multiple levels of object serialization and printing tools.
