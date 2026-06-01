package mq

import (
	"context"
	"fmt"

	"github.com/leoheung/go-patterns/container/list"
	"github.com/leoheung/go-patterns/container/safemap"
	"github.com/leoheung/go-patterns/parallel/pool"
	"github.com/leoheung/go-patterns/parallel/stream"
)

var _ MQ = (*ClusterMQ)(nil)

type topic = string

type ClusterMQ struct {
	root_ctx        context.Context
	root_ctx_cancel context.CancelFunc

	msgs        chan *Message
	subscribers *safemap.ShardedMap[topic, []*subscriber]
	wokerpool   *pool.AsyncPoolV2
}

func NewClusterMQ(msg_buffer int) *ClusterMQ {
	if msg_buffer < 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	ret := ClusterMQ{
		root_ctx:        ctx,
		root_ctx_cancel: cancel,
		msgs:            make(chan *Message, msg_buffer),
		subscribers:     safemap.NewShardedMap[topic, []*subscriber](32),
	}

	pipeline := stream.NewPipeline(ret.root_ctx.Done(), ret.msgs)

	return &ret
}

// Send implements [MQ].
func (c *ClusterMQ) Send(ctx context.Context, url string, msg Message) {
	panic("unimplemented")
}

// Broadcast implements [MQ].
func (c *ClusterMQ) Broadcast(ctx context.Context, topic string, msg Message) error {
	panic("unimplemented")
}

// Close implements [MQ].
func (c *ClusterMQ) Close(ctx context.Context) error {
	c.root_ctx_cancel()
	return nil
}

// Publish implements [MQ].
func (c *ClusterMQ) Publish(ctx context.Context, msg Message) error {
	select {
	case <-c.root_ctx.Done():
		return fmt.Errorf("MQ closed")
	case <-ctx.Done():
		return fmt.Errorf("context done")
	case c.msgs <- &msg:
		return nil
	default:
		return fmt.Errorf("msg buffer is full")
	}
}

// Subscribe implements [MQ].
func (c *ClusterMQ) Subscribe(ctx context.Context, topic string, subscriberID string, callbackURL string, maxInflight int) error {
	newS := subscriber{
		id:           subscriberID,
		callback_url: callbackURL,
		maxInflight:  maxInflight,
	}

	c.subscribers.Compute(topic,
		func() []*subscriber { return []*subscriber{&newS} },
		func(s []*subscriber) []*subscriber { s = append(s, &newS); return s })

	return nil
}

// Unsubscribe implements [MQ].
func (c *ClusterMQ) Unsubscribe(ctx context.Context, topic, subscriberID string) error {
	c.subscribers.Compute(topic, nil, func(s []*subscriber) []*subscriber {
		return list.From(s).Filter(func(v *subscriber, i int) bool {
			return v.id != subscriberID
		}).ToSlice()
	})
	return nil
}
