package wrapsocket

import (
	"time"

	"github.com/coder/websocket"
)

type Message struct {
	ID        string                `json:"id"`
	Type      websocket.MessageType `json:"type"`
	Data      []byte                `json:"data"`
	Timestamp time.Time             `json:"timestamp"`
}

func NewMessage(id string, msgType websocket.MessageType, data []byte) *Message {
	return &Message{
		ID:        id,
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
	}
}
