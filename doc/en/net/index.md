# Net

The `net` package provides network utilities for HTTP handling, file downloads, and common networking tasks.

## Modules

### Chi Router Utils

Utilities for working with the Chi router.

```go
import "github.com/leoxiang66/go-patterns/net"

// Wrap handler with middleware
handler := net.WrapHandler(myHandler)

// Common middleware setup
r := chi.NewRouter()
net.SetupCommonMiddleware(r)
```

### Download

File download utilities with progress tracking.

```go
// Download file with progress
err := net.DownloadFile("https://example.com/file.zip", "/path/to/save", func(progress float64) {
    fmt.Printf("Download progress: %.2f%%\n", progress*100)
})

// Simple download
err := net.DownloadFileSimple("https://example.com/file.zip", "/path/to/save")
```

### Common Network Utils

Common networking helper functions.

```go
// Check if URL is reachable
reachable := net.IsReachable("https://example.com")

// Get free port
port, err := net.GetFreePort()

// Parse URL
parsed, err := net.ParseURL("https://example.com/path?query=value")
```

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/net"
)

func main() {
    // Check if website is reachable
    if net.IsReachable("https://google.com") {
        fmt.Println("Google is reachable")
    }
    
    // Get a free port
    port, err := net.GetFreePort()
    if err != nil {
        panic(err)
    }
    fmt.Printf("Free port available: %d\n", port)
    
    // Download file with progress
    err = net.DownloadFile(
        "https://example.com/file.zip",
        "/tmp/file.zip",
        func(progress float64) {
            fmt.Printf("Progress: %.0f%%\n", progress*100)
        },
    )
    if err != nil {
        fmt.Printf("Download failed: %v\n", err)
    }
}
```

## Features

- **HTTP utilities**: Common HTTP helper functions
- **File download**: Download with progress tracking
- **Chi integration**: Utilities for Chi router
- **Port management**: Find available ports
