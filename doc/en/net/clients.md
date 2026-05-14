# Shared HTTP Client

## Overview

The `clients` package provides a thread-safe, reusable HTTP client with configurable connection pooling and connection management. It implements the singleton pattern to ensure only one HTTP client instance is shared across the application.

## Features

- **Connection Pooling**: Configurable max idle connections, max connections per host
- **TLS Configuration**: Custom TLS settings support
- **Proxy Support**: Configurable HTTP proxy
- **Panic Recovery**: Built-in panic handling to prevent crashes
- **Response Parsing**: Generic JSON response parsing

## HTTPClientConfig

Configuration struct for the shared HTTP client.

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| MaxIdleConns | `int` | `100` | Maximum number of idle connections in the pool |
| MaxIdleConnsPerHost | `int` | `10` | Maximum idle connections per host |
| MaxConnsPerHost | `int` | `100` | Maximum connections per host |
| IdleConnTimeout | `time.Duration` | `90s` | Idle connection timeout |
| TLSConfig | `*tls.Config` | `nil` | TLS configuration |
| CheckRedirect | `func(req *http.Request, via []*http.Request) error` | `nil` | Redirect policy |
| Proxy | `func(*http.Request) (*url.URL, error)` | `nil` | Proxy configuration |
| DisableKeepAlives | `bool` | `false` | Disable HTTP keep-alives |
| DisableCompression | `bool` | `false` | Disable compression |
| ForceAttemptHTTP2 | `bool` | `true` | Force HTTP/2 |

## API Reference

### `InitDefaultSharedHTTPClient()`

Initializes the shared HTTP client with default configuration.

```go
clients.InitDefaultSharedHTTPClient()
```

### `InitSharedHTTPClientWithConfig(config *HTTPClientConfig)`

Initializes the shared HTTP client with custom configuration. This function can only be called once due to the singleton pattern.

```go
config := &clients.HTTPClientConfig{
    MaxIdleConns:        200,
    MaxIdleConnsPerHost: 20,
    MaxConnsPerHost:     200,
    IdleConnTimeout:     60 * time.Second,
}
clients.InitSharedHTTPClientWithConfig(config)
```

### `Request(req *http.Request) (data []byte, headers http.Header, httpCode int, err error)`

Executes an HTTP request using the shared client. Returns the response body, headers, status code, and any error.

```go
req, _ := http.NewRequest("GET", "https://api.example.com/data", nil)
data, headers, code, err := clients.Request(req)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %d, Body: %s\n", code, string(data))
```

### `ParseResponse[T any](data []byte, dest *T) error`

Parses JSON response data into a typed struct.

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

var user User
err := clients.ParseResponse(data, &user)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User: %s <%s>\n", user.Name, user.Email)
```

## Example

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leoheung/go-patterns/net/clients"
)

func main() {
	// 1. Initialize the shared HTTP client
	clients.InitDefaultSharedHTTPClient()

	// 2. Create a request with timeout via context
	req, _ := http.NewRequest("GET", "https://api.example.com/users/1", nil)
	req = req.WithContext(req.Context())

	// 3. Execute the request
	data, headers, statusCode, err := clients.Request(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// 4. Process response
	fmt.Printf("Status: %d\n", statusCode)
	fmt.Printf("Content-Type: %s\n", headers.Get("Content-Type"))
	fmt.Printf("Body: %s\n", string(data))
}
```

## Notes

- The shared HTTP client uses `sync.Once` to ensure it's only initialized once
- Timeout should be controlled via request-level context, not client-level timeout
- Response bodies are automatically closed after reading
- Panic recovery is built-in to prevent request failures from crashing the application
- Connection pooling significantly improves performance for multiple requests to the same host