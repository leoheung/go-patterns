package pubsub

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type broker struct {
	topics map[string]map[uuid.UUID]chan any
}

var b *broker
var mu sync.Mutex

func InitPubSubSystem() {
	mu.Lock()
	defer mu.Unlock()

	if b == nil {
		b = &broker{
			topics: make(map[string]map[uuid.UUID]chan any),
		}
	}
}

func Shutdown() {
	mu.Lock()
	defer mu.Unlock()

	if b == nil {
		return
	}

	// 关闭所有订阅通道并清理
	for _, subs := range b.topics {
		for _, ch := range subs {
			if ch != nil {
				close(ch)
			}
		}
	}
	b.topics = nil
	b = nil
}

func Subscribe(topic string, buffer int) (*uuid.UUID, <-chan any, error) {
	mu.Lock()
	defer mu.Unlock()

	if b == nil {
		return nil, nil, fmt.Errorf("PubSub system is not initialized")
	}

	if buffer <= 0 {
		return nil, nil, fmt.Errorf("buffer should be > 0")
	}

	nextId := uuid.New()
	if b.topics[topic] == nil {
		b.topics[topic] = make(map[uuid.UUID]chan any)
	}
	ch := make(chan any, buffer)
	b.topics[topic][nextId] = ch
	return &nextId, ch, nil
}

func Unsubscribe(topic string, id *uuid.UUID) {
	mu.Lock()
	defer mu.Unlock()

	if b == nil || id == nil {
		return
	}

	subscribers, ok := b.topics[topic]
	if !ok {
		return
	}

	ch, ok2 := subscribers[*id]
	if !ok2 {
		return
	}

	// 关闭由 broker 管理的订阅通道，防止向已关闭通道写入请确保仅由 broker 关闭
	if ch != nil {
		close(ch)
	}

	delete(subscribers, *id)
	if len(subscribers) == 0 {
		delete(b.topics, topic)
	}
}

func Publish(topic string, data any) error {
	mu.Lock()
	defer mu.Unlock()

	if b == nil {
		return fmt.Errorf("PubSub is not initialized")
	}

	if subscribers, ok := b.topics[topic]; !ok {
		return fmt.Errorf("no subscribers")
	} else {
		for _, c := range subscribers {
			select {
			case c <- data:
			default:
			}
		}
		return nil
	}
}
