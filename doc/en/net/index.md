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

### Stream Download

Stream any `io.Reader` to HTTP response with proper headers. Supports both fixed-size and chunked transfer.

```go
// Stream download with known file size
fileData := bytes.NewReader(fileBytes)
size := int64(len(fileBytes))
net.StreamDownloadHandler(w, fileData, "report.pdf", "application/pdf", &size)

// Stream download with unknown size (uses chunked transfer)
s3Reader := getS3ObjectReader(key)
net.StreamDownloadHandler(w, s3Reader, "backup.zip", "application/zip", nil)
```

**Parameters:**

- `w`: HTTP ResponseWriter
- `reader`: Any io.Reader (file, memory buffer, S3 object, etc.)
- `filename`: Download filename shown to user
- `contentType`: MIME type (e.g., "application/pdf", "application/octet-stream")
- `size`: File size pointer (optional, pass nil for chunked transfer)

### WebSocket Manager

A high-level WebSocket connection management framework. See [WebSocket Documentation](./wrapsocket) for details.

```go
import "github.com/leoheung/go-patterns/net/wrapsocket"

// Create WebSocket handler
handler := wrapsocket.NewDefaultHandler(nil)

// Set callbacks
handler.SetOnConnect(func(conn *wrapsocket.Conn) {
    fmt.Printf("Client connected: %s\n", conn.ID)
})

handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
    // Echo message back
    conn.Write(ctx, msg.Type, msg.Data)
})

http.ListenAndServe(":8080", handler)
```

**Features:**

- Connection lifecycle management
- Heartbeat detection
- Group broadcasting
- Metadata storage

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
