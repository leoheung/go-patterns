# Net

The `net` package provides utilities for HTTP response handling, concurrent file downloads, and pointer helper functions.

## Installation

```go
import "github.com/leoheung/go-patterns/net"
```

## API Reference

### Concurrent Download

Download files using multiple goroutines with progress tracking.

```go
// Concurrent download with 4 workers
err := net.DownloadFileByConcurrent("https://example.com/file.zip", "./downloads/", 4)
```

### HTTP Response Helpers

Standardized JSON and CSV response helpers for web services.

```go
// Return a standardized JSON success response
net.ReturnJsonResponse(w, http.StatusOK, map[string]string{"message": "success"})

// Return a standardized JSON error response
net.ReturnErrorResponse(w, http.StatusBadRequest, "invalid input")

// Return a CSV file response
headers := []string{"ID", "Name"}
rows := [][]string{{"1", "Alice"}, {"2", "Bob"}}
net.ReturnCSVResponse(w, "users.csv", headers, rows)
```

### Chi Router Utilities

```go
// Print all registered routes in a Chi mux
net.PrintCHIRoutes(r)
```

### Pointer Helpers

Commonly used to create pointers for primitive types (useful for database models).

```go
s := net.PtrString("hello")
i := net.PtrInt(100)
b := net.PtrBool(true)
t := net.PtrTime(time.Now())
```

## Complete Example

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/leoheung/go-patterns/net"
)

func main() {
    r := chi.NewRouter()

    r.Get("/api/data", func(w http.ResponseWriter, r *http.Request) {
        data := struct {
            ID   int    `json:"id"`
            Name string `json:"name"`
        }{ID: 1, Name: "Pattern"}
        
        net.ReturnJsonResponse(w, http.StatusOK, data)
    })

    // Print routes for debugging
    net.PrintCHIRoutes(r)

    // Download a file concurrently
    go net.DownloadFileByConcurrent("https://example.com/large-file.bin", "./tmp/", 8)

    http.ListenAndServe(":8080", r)
}
```

## Features

- **Concurrent Download**: Multi-threaded downloading with automatic filename resolution.
- **Standardized Responses**: Consistent `UniversalResponse` structure for JSON APIs.
- **Route Debugging**: Easily visualize Chi router structures.
- **Pointer Utils**: Convenient helpers for handling optional fields in structs.
