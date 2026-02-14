package wrapsocket

import (
	"context"
	"fmt"
	"sync"

	"github.com/coder/websocket"
)

type ConnManager struct {
	conns map[string]*Conn
	mu    sync.RWMutex
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		conns: make(map[string]*Conn),
	}
}

func (m *ConnManager) Add(conn *Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.conns[conn.ID] = conn
}

func (m *ConnManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.conns, id)
}

func (m *ConnManager) Get(id string) (*Conn, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.conns[id]
	return conn, ok
}

func (m *ConnManager) GetAll() []*Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conns := make([]*Conn, 0, len(m.conns))
	for _, conn := range m.conns {
		conns = append(conns, conn)
	}
	return conns
}

func (m *ConnManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.conns)
}

func (m *ConnManager) Broadcast(ctx context.Context, msgType websocket.MessageType, data []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, conn := range m.conns {
		if !conn.IsClosed() {
			if err := conn.ws.Write(ctx, msgType, data); err != nil {
				continue
			}
		}
	}
	return nil
}

func (m *ConnManager) SendTo(ctx context.Context, connID string, msgType websocket.MessageType, data []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, ok := m.conns[connID]
	if !ok {
		return fmt.Errorf("connection not found: %s", connID)
	}
	if conn.IsClosed() {
		return fmt.Errorf("connection is closed: %s", connID)
	}
	return conn.Write(ctx, msgType, data)
}

func (m *ConnManager) GetByGroup(group string) []*Conn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conns := make([]*Conn, 0)
	for _, conn := range m.conns {
		if conn.Group == group && !conn.IsClosed() {
			conns = append(conns, conn)
		}
	}
	return conns
}

func (m *ConnManager) SendToGroup(ctx context.Context, group string, msgType websocket.MessageType, data []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, conn := range m.conns {
		if conn.Group == group && !conn.IsClosed() {
			_ = conn.Write(ctx, msgType, data)
		}
	}
	return nil
}

func (m *ConnManager) CloseAll(code websocket.StatusCode, reason string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, conn := range m.conns {
		_ = conn.Close(code, reason)
	}
	m.conns = make(map[string]*Conn)
}
