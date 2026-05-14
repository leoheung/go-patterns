# Logging

## Overview

The `utils` package provides logging utilities with environment awareness, automatically switching between development and production logging modes.

## API Reference

### `IsDev() bool`

Checks if the application is running in development environment (when `env` environment variable equals `"dev"`).

```go
if utils.IsDev() {
    fmt.Println("Running in development mode")
}
```

### `LogMessage(message string)`

Logs a message. Uses `fmt.Println` in development mode and `log.Println` in production mode.

```go
utils.LogMessage("Application started")
```

### `DevLogError(errMsg string)`

Logs an error message in development mode with timestamp and red color.

```go
utils.DevLogError("Failed to connect to database")
// Output: [Dev Logs] - 2024-01-01 12:00:00: Failed to connect to database
```

### `DevLogInfo(infoMsg string)`

Logs an info message in development mode with timestamp and bright blue color.

```go
utils.DevLogInfo("Processing request")
```

### `DevLogSuccess(successMsg string)`

Logs a success message in development mode with timestamp and green color.

```go
utils.DevLogSuccess("User logged in successfully")
```

## Example

```go
package main

import (
	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// Check environment
	if utils.IsDev() {
		utils.LogMessage("Running in development mode")
	}

	// Log messages based on environment
	utils.LogMessage("Application started")

	// Development-only colored logs
	utils.DevLogError("This is an error")
	utils.DevLogInfo("This is info")
	utils.DevLogSuccess("This is success")
}
```

## Notes

- Set `env=dev` environment variable to enable development mode
- Development logs are colored and include timestamps
- Production logs use standard library `log` package