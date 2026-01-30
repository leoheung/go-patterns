# Utils

The `utils` package provides utility functions for common operations.

## Modules

### Logging

Check if in development environment and log messages accordingly.

```go
import "github.com/leoheung/go-patterns/utils"

// Check if in development environment
isDev := utils.IsDev() // Returns true if env=dev

// Log message (uses fmt.Println in dev, log.Println in prod)
utils.LogMessage("Hello, world!")
```

### Retry

Retry a function with error/panic handling.

```go
// Retry a function with error/panic handling
// work: Function to execute
// retryTimes: Maximum retry attempts (excluding first try)
utils.RetryWork(
    func() error {
        // Operation that might fail
        return nil // or error
    },
    3, // Retry 3 times if failed
)
```

### Timeout

Execute a function with a timeout.

```go
// Execute function with timeout
err := utils.WithTimeout(5*time.Second, func() error {
    // Long running operation
    return nil
})
```

### Pretty Print

Pretty print objects for debugging.

```go
// Pretty print an object
utils.PrettyPrint(obj)

// Pretty print with label
utils.PrettyPrintWithLabel("User", user)
```

### Color Print

Print colored messages to the console.

```go
// Print in different colors
utils.Red("Error message")
utils.Green("Success message")
utils.Yellow("Warning message")
utils.Blue("Info message")
```

### Number Utils

Utility functions for number operations.

```go
// Min/Max functions
min := utils.Min(1, 2, 3) // 1
max := utils.Max(1, 2, 3) // 3

// Clamp function
clamped := utils.Clamp(10, 0, 5) // 5
```

## Complete Example

```go
package main

import (
    "github.com/leoheung/go-patterns/utils"
    "time"
)

func main() {
    // Set environment to development
    // os.Setenv("env", "dev")

    // Retry a potentially failing operation
    utils.RetryWork(func() error {
        utils.LogMessage("Attempting operation...")
        // Simulate failure
        if time.Now().Nanosecond()%2 == 0 {
            panic("simulated panic")
        }
        return nil
    }, 3)
}
```

## Features

- **Environment-aware**: Different behavior in dev/prod
- **Error handling**: Retry mechanisms with panic recovery
- **Debugging tools**: Pretty printing and color output
- **Common utilities**: Min, Max, Clamp functions
