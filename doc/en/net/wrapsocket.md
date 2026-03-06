# WebSocket Manager (wrapsocket)

A high-level WebSocket connection management framework built on top of `coder/websocket`. It provides connection lifecycle management, heartbeat detection, group broadcasting, and metadata storage.

## Installation

```go
import "github.com/leoheung/go-patterns/net/wrapsocket"
```

**Dependencies:**
- `github.com/coder/websocket` - WebSocket implementation

## Quick Start

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/leoheung/go-patterns/net/wrapsocket"
)

func main() {
    // Create handler with default options
    handler := wrapsocket.NewDefaultHandler(nil)

    // Set connection callback
    handler.SetOnConnect(func(conn *wrapsocket.Conn) {
        fmt.Printf("Client connected: %s\n", conn.ID)
    })

    // Set message callback
    handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
        fmt.Printf("Received from %s: %s\n", conn.ID, string(msg.Data))
    })

    // Set disconnect callback
    handler.SetOnDisconnect(func(conn *wrapsocket.Conn) {
        fmt.Printf("Client disconnected: %s\n", conn.ID)
    })

    http.ListenAndServe(":8080", handler)
}
```

## API Reference

### Handler

The Handler interface manages WebSocket upgrades and connection lifecycle.

```go
// Create handler with custom WebSocket accept options
opts := &websocket.AcceptOptions{
    OriginPatterns: []string{"localhost", "*.example.com"},
}
handler := wrapsocket.NewDefaultHandler(opts)
```

#### Lifecycle Hooks

```go
// Called when a new client connects
handler.SetOnConnect(func(conn *wrapsocket.Conn) {
    // Connection is already added to manager
    // Send welcome message or authenticate
})

// Called when a client disconnects
handler.SetOnDisconnect(func(conn *wrapsocket.Conn) {
    // Cleanup resources
})

// Called when a message is received
handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
    // Handle incoming message
})

// Called when an error occurs
handler.SetOnError(func(conn *wrapsocket.Conn, err error) {
    // Log or handle error
})
```

#### Heartbeat Configuration

```go
config := &wrapsocket.HeartbeatConfig{
    Interval:  30 * time.Second,  // Ping interval
    Timeout:   10 * time.Second,  // Ping timeout
    MaxMissed: 3,                 // Max missed pings before disconnect
}
handler.SetHeartbeatConfig(config)
```

### Conn (Connection)

Each WebSocket connection is wrapped in a `Conn` struct with additional metadata.

```go
// Connection properties
type Conn struct {
    ID       string                 // Auto-generated UUID
    Group    string                 // Optional group name
    metadata map[string]interface{} // Custom metadata storage
}
```

#### Metadata Operations

```go
// Store custom data
conn.SetMetadata("user_id", "12345")
conn.SetMetadata("room", "lobby")

// Retrieve data
if val, ok := conn.GetMetadata("user_id"); ok {
    userID := val.(string)
}
```

#### Send Message

```go
// Send to specific connection
ctx := context.Background()
err := conn.Write(ctx, websocket.MessageText, []byte("hello"))

// Check connection status
if !conn.IsClosed() {
    // Safe to send
}
```

### ConnManager

Manages all active connections with thread-safe operations.

```go
manager := handler.Manager()
```

#### Connection Operations

```go
// Get connection by ID
conn, ok := manager.Get("conn-uuid")

// Get all connections
conns := manager.GetAll()

// Get connection count
count := manager.Count()
```

#### Broadcasting

```go
// Broadcast to all connections
ctx := context.Background()
manager.Broadcast(ctx, websocket.MessageText, []byte("announcement"))

// Send to specific connection
manager.SendTo(ctx, "conn-uuid", websocket.MessageText, []byte("private"))
```

#### Group Operations

```go
// Assign group (in OnConnect callback)
handler.SetOnConnect(func(conn *wrapsocket.Conn) {
    conn.Group = "room-1"
})

// Get connections by group
roomConns := manager.GetByGroup("room-1")

// Broadcast to group
manager.SendToGroup(ctx, "room-1", websocket.MessageText, []byte("room message"))
```

#### Cleanup

```go
// Close all connections
manager.CloseAll(websocket.StatusGoingAway, "server shutdown")
```

### Message

Message structure received from clients.

```go
type Message struct {
    ID        string                // Connection ID
    Type      websocket.MessageType // Message type (Text/Binary)
    Data      []byte                // Raw message data
    Timestamp time.Time             // Receive time
}
```

## Complete Examples

### Chat Server with Rooms

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/coder/websocket"
    "github.com/leoheung/go-patterns/net/wrapsocket"
)

type ChatMessage struct {
    Room    string `json:"room"`
    Content string `json:"content"`
}

func main() {
    handler := wrapsocket.NewDefaultHandler(nil)
    
    // Configure heartbeat
    handler.SetHeartbeatConfig(&wrapsocket.HeartbeatConfig{
        Interval:  30 * time.Second,
        Timeout:   10 * time.Second,
        MaxMissed: 3,
    })
    
    manager := handler.Manager()
    
    handler.SetOnConnect(func(conn *wrapsocket.Conn) {
        fmt.Printf("User %s joined\n", conn.ID)
    })
    
    handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
        var chatMsg ChatMessage
        if err := json.Unmarshal(msg.Data, &chatMsg); err != nil {
            return
        }
        
        // Join room
        conn.Group = chatMsg.Room
        
        // Broadcast to room
        response, _ := json.Marshal(map[string]string{
            "from":    conn.ID,
            "content": chatMsg.Content,
        })
        manager.SendToGroup(context.Background(), chatMsg.Room, 
            websocket.MessageText, response)
    })
    
    handler.SetOnDisconnect(func(conn *wrapsocket.Conn) {
        fmt.Printf("User %s left room %s\n", conn.ID, conn.Group)
    })
    
    http.ListenAndServe(":8080", handler)
}
```

### Connection with Authentication

```go
handler.SetOnConnect(func(conn *wrapsocket.Conn) {
    // Store auth info in metadata
    conn.SetMetadata("authenticated", false)
    conn.SetMetadata("auth_time", time.Now())
})

handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
    // Check authentication
    if auth, _ := conn.GetMetadata("authenticated"); !auth.(bool) {
        // Handle auth message
        if isValidToken(string(msg.Data)) {
            conn.SetMetadata("authenticated", true)
            conn.Write(context.Background(), websocket.MessageText, 
                []byte(`{"type":"auth_success"}`))
        }
        return
    }
    
    // Process authenticated message
})
```

## Features

- **Thread-Safe**: All connection and manager operations are protected by mutex
- **Auto UUID**: Each connection gets a unique ID automatically
- **Heartbeat**: Configurable ping/pong with timeout detection
- **Grouping**: Built-in support for connection groups/rooms
- **Metadata**: Attach custom data to connections
- **Lifecycle Hooks**: Connect, disconnect, message, and error callbacks
- **Graceful Shutdown**: Close all connections cleanly

## Notes

- The handler automatically sends a welcome message with connection ID upon connection
- Heartbeat is optional; if not configured, no ping/pong will be performed
- All callbacks are executed synchronously; use goroutines for long operations
- The manager's Broadcast skips closed connections automatically
