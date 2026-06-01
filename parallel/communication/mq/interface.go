package mq

import (
	"context"
	"time"
)

type MessageHeader struct {
	TraceID   string    `json:"traceID"`
	Source    string    `json:"source"`
	EventType string    `json:"eventType"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

type Message struct {
	ID      string        `json:"string"`
	Topic   string        `json:"topic"`
	Payload []byte        `json:"payload"`
	Headers MessageHeader `json:"headers"`
}

type subscriber struct {
	id           string
	callback_url string
	maxInflight  int
}

type HttpClient interface {
	Send(ctx context.Context, url string, msg Message)
}

type Broadcaster interface {
	Broadcast(ctx context.Context, topic string, msg Message) error
}

type Publisher interface {
	Publish(ctx context.Context, msg Message) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic, subscriberID, callbackURL string, maxInflight int) error
	Unsubscribe(ctx context.Context, topic, subscriberID string) error
}

type MQ interface {
	Publisher
	Subscriber
	Broadcaster
	HttpClient
	Close(ctx context.Context) error
}
