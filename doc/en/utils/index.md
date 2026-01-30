# Utils

The `utils` package provides a collection of utility functions for common operations, covering logging, retry mechanisms, timeout control, object beautification, and numeric processing.

## Modules

### 1. Logging and Environment Awareness

Automatically identifies development or production environments and adopts different logging methods.

```go
import "github.com/leoheung/go-patterns/utils"

// Check if in development environment (determined by env == "dev" environment variable)
isDev := utils.IsDev()

// Log message (uses fmt.Println in dev, standard library log in prod)
utils.LogMessage("Hello, world!")

// Dev-only colored logs (outputs only when IsDev() is true)
utils.DevLogError("This is an error message")
utils.DevLogInfo("This is an info message")
utils.DevLogSuccess("This is a success message")
```

### 2. Retry and Timeout Control

Provides protection mechanisms for unstable operations.

```go
// Retry a work function
// work: function to execute, returns (any, error)
// retryTimes: number of retries after initial failure
data, err := utils.RetryWork(func() (any, error) {
    // business logic
    return "result", nil
}, 3)

// Execution with timeout
// Returns fmt.Errorf("timeout") if timed out
res, err := utils.TimeoutWork(func() (any, error) {
    time.Sleep(2 * time.Second)
    return "done", nil
}, 1 * time.Second)
```

### 3. Object Beautification and JSON Processing

Advanced beautification tools based on `go-spew`.

```go
// Formatted printing of any object (common for debugging)
utils.PPrint(myStruct)
utils.PPrettyPrint(myStruct)

// Get a pretty string representation of an object (no direct printing)
str := utils.PrettyObjStr(myStruct)

// Serialize an object to a pretty JSON string (falls back to PrettyObjStr on failure)
jsonStr := utils.JSONalizeStr(myStruct)

// Deserialize a JSON string to an object (must pass a pointer)
err := utils.DeJSONalizeStr(jsonStr, &myTarget)
```

### 4. Colored Terminal Output

```go
// Output colored text using ANSI escape sequences
utils.PrintlnColor(utils.Red, "Red text")
utils.PrintlnColor(utils.Green, "Green text")
utils.PrintlnColor(utils.BrightBlue, "Bright blue text")

// Available color constants:
// utils.Red, utils.Green, utils.BrightBlue, utils.Magenta, utils.Cyan
```

### 5. Number Utils

Simulates numeric processing found in dynamic languages.

```go
// Parse a string into a Number object
n, err := utils.ParseNumber("100.5")

// Get values in different types
f := n.Float()   // 100.5
i := n.Int()     // 100
i64 := n.Int64() // 100

// Check if it's an integer (no fractional part)
isUint := n.IsInteger() // false
```

### 6. Common Helpers

```go
// Check if any value is nil (supports Interface, Slice, Map, Ptr, etc.)
isNull := utils.IsNil(someVar)

// Check if a string consists entirely of digits
allDigits := utils.IsDigits("12345")

// Delayed function execution (blocking)
utils.DelayDo(500 * time.Millisecond, func() {
    fmt.Println("Delayed execution")
})

// Block the current Goroutine indefinitely
utils.Hold()
```

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
