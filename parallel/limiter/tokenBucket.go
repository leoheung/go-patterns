/*
令牌桶算法是网络流量整形和限流的经典算法，广泛应用于 API 网关、微服务熔断、数
据库连接池控制等场景。其核心思想是：
• 以固定速率向桶中添加令牌
• 每个请求需要从桶中获取一个令牌才能通过
• 桶满时令牌溢出，桶空时请求被拒绝或等待
*/

package limiter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

var errClosed = errors.New("token bucket is closed")

type emptyT struct{}

type TokenBucket struct {
	capability int
	ticker     *time.Ticker
	tokens     chan emptyT
	closed     chan emptyT
}

// NewTokenBucket 创建令牌桶限流器
// capacity: 桶容量（最大令牌数）
// interval: 令牌生成時間interval（令牌/秒）
func NewTokenBucket(capacity int, interval time.Duration) *TokenBucket {
	var ret TokenBucket
	ret.capability = capacity
	ret.ticker = time.NewTicker(interval)
	ret.tokens = make(chan emptyT, capacity)
	ret.closed = make(chan emptyT, 1)

	for range capacity {
		ret.tokens <- emptyT{}
	}

	go func() {
		for {
			if ret.isClosed() {
				ret.ticker.Stop()
				return
			}

			<-ret.ticker.C
			ret.incToken()
		}
	}()

	return &ret
}

func (tb *TokenBucket) TryGrant() (bool, error) {
	select {
	case <-tb.closed:
		utils.TryEnqueue(tb.closed, emptyT{})
		return false, errClosed
	case <-tb.tokens:
		return true, nil
	default:
		return false, nil
	}
}

func (tb *TokenBucket) Grant(ctx context.Context) error {
	select {
	case <-tb.tokens:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("failed to grant token: context done: %s", ctx.Err().Error())
	case <-tb.closed:
		utils.TryEnqueue(tb.closed, emptyT{})
		return errClosed
	}
}

func (tb *TokenBucket) Close() {
	utils.TryEnqueue(tb.closed, emptyT{})
}

func (tb *TokenBucket) isClosed() bool {
	return len(tb.closed) == 1
}

func (tb *TokenBucket) incToken() {
	utils.TryEnqueue(tb.tokens, emptyT{})
}
