package wrapsocket

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coder/websocket"
	"github.com/leoheung/go-patterns/cryptography"
	"github.com/leoheung/go-patterns/utils"
)

type Handler interface {
	http.Handler

	SetOnConnect(func(*Conn))
	SetOnDisconnect(func(*Conn))
	SetOnMessage(func(*Conn, *Message))
	SetOnError(func(*Conn, error))
	SetHeartbeatConfig(*HeartbeatConfig)
	Manager() *ConnManager
}

type DefaultHandler struct {
	ID              string
	manager         *ConnManager
	hooks           *Hooks
	opts            *websocket.AcceptOptions
	heartbeatConfig *HeartbeatConfig
}

func NewDefaultHandler(opts *websocket.AcceptOptions) Handler {
	return &DefaultHandler{
		ID:      cryptography.RandUUID(),
		manager: NewConnManager(),
		hooks:   &Hooks{},
		opts:    opts,
	}
}

func (h *DefaultHandler) Manager() *ConnManager {
	return h.manager
}

func (h *DefaultHandler) SetOnConnect(fn func(*Conn)) {
	h.hooks.OnConnect = fn
}

func (h *DefaultHandler) SetOnDisconnect(fn func(*Conn)) {
	h.hooks.OnDisconnect = fn
}

func (h *DefaultHandler) SetOnMessage(fn func(*Conn, *Message)) {
	h.hooks.OnMessage = fn
}

func (h *DefaultHandler) SetOnError(fn func(*Conn, error)) {
	h.hooks.OnError = fn
}

func (h *DefaultHandler) SetHeartbeatConfig(config *HeartbeatConfig) {
	h.heartbeatConfig = config
}

func (h *DefaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, h.opts)
	if err != nil {
		utils.DevLogError(fmt.Sprintf("[%s] WebSocket accept error: %v", h.ID, err))
		return
	}

	c := NewConn(conn)
	h.manager.Add(c)

	if h.hooks.OnConnect != nil {
		h.hooks.OnConnect(c)
	}

	welcomeMsg := fmt.Sprintf(`{"type":"connected","id":"%s"}`, c.ID)
	if err := c.ws.Write(r.Context(), websocket.MessageText, []byte(welcomeMsg)); err != nil {
		utils.DevLogError(fmt.Sprintf("[%s] Failed to send welcome message: %v", h.ID, err))
		h.manager.Remove(c.ID)
		c.Close(websocket.StatusInternalError, "failed to send welcome")
		return
	}

	utils.DevLogInfo(fmt.Sprintf("[%s] Client connected: %s", h.ID, c.ID))

	if h.heartbeatConfig != nil {
		c.StartHeartbeat(r.Context(), h.heartbeatConfig)
	}

	h.handleConn(r.Context(), c)
}

func (h *DefaultHandler) handleConn(ctx context.Context, c *Conn) {
	defer func() {
		h.manager.Remove(c.ID)
		if h.hooks.OnDisconnect != nil {
			h.hooks.OnDisconnect(c)
		}
		c.Close(websocket.StatusNormalClosure, "connection closed")
		utils.DevLogInfo(fmt.Sprintf("[%s] Client disconnected: %s", h.ID, c.ID))
	}()

	for {
		msgType, data, err := c.ws.Read(ctx)
		if err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
				websocket.CloseStatus(err) == websocket.StatusGoingAway {
				return
			}
			if h.hooks.OnError != nil {
				h.hooks.OnError(c, err)
			} else {
				utils.DevLogError(fmt.Sprintf("[%s] Read error: %v", h.ID, err))
			}
			return
		}

		c.UpdateLastSeen()

		if h.hooks.OnMessage != nil {
			msg := NewMessage(c.ID, msgType, data)
			h.hooks.OnMessage(c, msg)
		}
	}
}
