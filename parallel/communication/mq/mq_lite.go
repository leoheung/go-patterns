package mq

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/leoheung/go-patterns/container/list"
	"github.com/leoheung/go-patterns/container/safemap"
	"github.com/leoheung/go-patterns/net/clients"
	"github.com/leoheung/go-patterns/parallel/pool"
	"github.com/leoheung/go-patterns/parallel/stream"
	"github.com/leoheung/go-patterns/utils"
)

var _ MQ = new(MQLite)

type topic = string

type retryStruct struct {
	sub *subscriber
	msg *Message
}

type MQLite struct {
	root_ctx        context.Context
	root_ctx_cancel context.CancelFunc

	msgs        chan *Message
	retry_ch    chan *retryStruct
	subscribers *safemap.ShardedMap[topic, []*subscriber]
	wokerpool   *pool.AsyncPoolV2
	httpClient  *clients.SharedHTTPClient
	msg_count   *int64
}

func NewMQLite(msg_buffer_size, num_workers, job_queue_size, shardCount int) *MQLite {
	if msg_buffer_size < 0 || num_workers < 0 || job_queue_size < 0 || shardCount < 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	ret := MQLite{
		root_ctx:        ctx,
		root_ctx_cancel: cancel,
		msgs:            make(chan *Message, msg_buffer_size),
		retry_ch:        make(chan *retryStruct),
		subscribers:     safemap.NewShardedMap[topic, []*subscriber](shardCount),
		httpClient:      clients.NewDefaultSharedHTTPClient(),
		wokerpool:       pool.NewAsyncPoolV2(int32(num_workers), job_queue_size),
		msg_count:       new(int64),
	}

	go stream.NewPipeline(ret.root_ctx.Done(), ret.msgs).
		ForEach(func(m *Message) {
			ret.Broadcast(ret.root_ctx, m)
		})
	go stream.NewPipeline(ret.root_ctx.Done(), ret.retry_ch).
		Buffer(32).
		Parallel(3, func(r *retryStruct) error {
			for i := range r.sub.maxInflight {
				if ret.Send(ret.root_ctx, r.sub, r.msg) == nil {
					return nil
				}
				time.Sleep(time.Duration(i+1) * time.Second)
			}
			return fmt.Errorf("failed to resend msg %s to subscriber %s after retried %d times", r.msg.ID, r.sub.id, r.sub.maxInflight)
		}).
		ForEach(func(e error) {
			if e != nil {
				utils.DevLogError(e.Error())
			}
		})

	return &ret
}

// Send implements [MQ].
func (c *MQLite) Send(ctx context.Context, sub *subscriber, msg *Message) error {
	sub.mu.Lock()
	defer sub.mu.Unlock()

	payload, err := clients.AnyToBody(msg)
	if err != nil {
		return fmt.Errorf("failed to jsonlize the message: %s", err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sub.callback_url, payload)

	if err != nil {
		return fmt.Errorf("failed to create a request: %s", err.Error())
	}

	_, _, code, err := c.httpClient.Request(req)
	if err != nil {
		return fmt.Errorf("failed to send notification request to subscriber: %w", err)
	}
	if code != http.StatusOK {
		return fmt.Errorf("[%d] failed to send notification request to subscriber", code)
	}

	return nil
}

// Broadcast implements [MQ].
func (c *MQLite) Broadcast(ctx context.Context, msg *Message) error {
	subs, ok := c.subscribers.Get(msg.Topic)
	if ok && len(subs) > 0 {
		for _, s := range subs {
			job := func(ctx context.Context) error {
				return c.Send(ctx, s, msg)
			}

			onError := func(err error) {
				c.retry_ch <- &retryStruct{
					sub: s,
					msg: msg,
				}
			}

			err := c.wokerpool.AsyncSubmit(c.root_ctx, job, onError)
			if err != nil {
				utils.DevLogError(fmt.Sprintf("failed to submit notifiying task into the pool: %s for message id: %s", err.Error(), msg.ID))
			}
		}
	}
	return nil
}

// Close implements [MQ].
func (c *MQLite) Close(ctx context.Context) error {
	c.root_ctx_cancel()
	return nil
}

// Publish implements [MQ].
func (c *MQLite) Publish(ctx context.Context, msg *Message) (*string, error) {
	if _, ok := c.subscribers.Get(msg.Topic); !ok {
		return nil, fmt.Errorf("Message is not published as there are no subscribers to this topic")
	}

	msg.ID = strconv.FormatInt(atomic.AddInt64(c.msg_count, 1), 10)

	select {
	case <-c.root_ctx.Done():
		return nil, fmt.Errorf("MQ closed")
	case <-ctx.Done():
		return nil, fmt.Errorf("context done")
	case c.msgs <- msg:
		return &msg.ID, nil
	default:
		return nil, fmt.Errorf("msg buffer is full")
	}
}

// Subscribe implements [MQ].
func (c *MQLite) Subscribe(ctx context.Context, topic string, subscriberID string, callbackURL string, maxInflight int) error {
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
func (c *MQLite) Unsubscribe(ctx context.Context, topic, subscriberID string) error {
	c.subscribers.Compute(topic, nil, func(s []*subscriber) []*subscriber {
		return list.From(s).Filter(func(v *subscriber, i int) bool {
			return v.id != subscriberID
		}).ToSlice()
	})
	return nil
}
