package wrapsocket

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/leoheung/go-patterns/cryptography"

	"github.com/coder/websocket"
)

type Conn struct {
	ID          string
	ws          *websocket.Conn
	mu          sync.Mutex
	metadata    map[string]interface{}
	lastSeen    time.Time
	closed      bool
	closeCode   websocket.StatusCode
	closeReason string
	heartbeat   *Heartbeat
	Group       string
}

func NewConn(ws *websocket.Conn) *Conn {
	return &Conn{
		ID:          cryptography.RandUUID(),
		ws:          ws,
		lastSeen:    time.Now(),
		closed:      false,
		closeCode:   0,
		closeReason: "",
	}
}

func (c *Conn) UpdateLastSeen() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastSeen = time.Now()
}

func (c *Conn) GetLastSeen() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastSeen
}

func (c *Conn) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.metadata == nil {
		c.metadata = make(map[string]interface{})
	}
	c.metadata[key] = value
}

func (c *Conn) GetMetadata(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.metadata == nil {
		return nil, false
	}
	val, ok := c.metadata[key]
	return val, ok
}

func (c *Conn) Close(code websocket.StatusCode, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	c.closeCode = code
	c.closeReason = reason
	if c.heartbeat != nil {
		c.heartbeat.Stop()
	}
	return c.ws.Close(code, reason)
}

func (c *Conn) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closed
}

func (c *Conn) StartHeartbeat(ctx context.Context, config *HeartbeatConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.heartbeat != nil {
		return
	}
	c.heartbeat = NewHeartbeat(c, config)
	go c.heartbeat.Start(ctx)
}

func (c *Conn) Write(ctx context.Context, msgType websocket.MessageType, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("connection is closed")
	}
	return c.ws.Write(ctx, msgType, data)
}
