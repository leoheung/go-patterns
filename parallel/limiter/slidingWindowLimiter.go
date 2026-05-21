package limiter

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/leoheung/go-patterns/container/safelist"
	"github.com/leoheung/go-patterns/container/safemap"
)

type SlidingWindowLimiter struct {
	frequency_hz  int
	buckets_count int
	buckets       *safemap.ShardedMap[int, *int64]
	start_time    time.Time     // 值类型，不用指针
	quit_ch       chan struct{} // 用 struct{} 比 emptyT 更惯用
	keys          *safelist.SafeSlice[int]
}

func NewSlidingWindowLimiter(frequency_hz, buckets_count int) (*SlidingWindowLimiter, error) {
	if frequency_hz <= 0 || buckets_count <= 0 {
		return nil, fmt.Errorf("frequency_hz or buckets_count <= 0")
	}

	ret := &SlidingWindowLimiter{
		frequency_hz:  frequency_hz,
		buckets_count: buckets_count,
		buckets:       safemap.NewShardedMap[int, *int64](buckets_count),
		start_time:    time.Now(),
		quit_ch:       make(chan struct{}),
		keys:          safelist.NewSafeSlice[int](0, 0),
	}

	go ret.cron_clean()
	return ret, nil
}

// TryGrant 非阻塞，直接在当前 goroutine 执行
func (sw *SlidingWindowLimiter) TryGrant() (bool, error) {
	select {
	case <-sw.quit_ch:
		return false, fmt.Errorf("limiter closed")
	default:
	}

	idx := sw.getCurrBucketIndex()

	// 原子读取并判断
	if sw.countRequestsWithinWindow(idx) >= sw.frequency_hz {
		return false, nil
	}

	sw.incRequestCount(idx)
	return true, nil
}

func (sw *SlidingWindowLimiter) Close() {
	close(sw.quit_ch)
}

// 修复：响应退出信号
func (sw *SlidingWindowLimiter) cron_clean() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sw.quit_ch:
			return
		case <-ticker.C:
			sw.cleanupExpired()
		}
	}
}

func (sw *SlidingWindowLimiter) cleanupExpired() {
	currentIdx := sw.getCurrBucketIndex()
	cutoff := currentIdx - sw.buckets_count

	// 清理 keys 和 buckets
	sw.keys.RemoveIf(func(i int) bool {
		if i <= cutoff {
			sw.buckets.Delete(i) // 同时清理 map
			return true
		}
		return false
	})
}

func (sw *SlidingWindowLimiter) getCurrBucketIndex() int {
	elapsed := time.Since(sw.start_time)
	interval := time.Second / time.Duration(sw.buckets_count)
	return int(elapsed / interval)
}

// 修复：原子递增，避免 read-then-write 竞态
func (sw *SlidingWindowLimiter) incRequestCount(idx int) {
	// 用 safemap 的原子操作（如果支持），否则用 sync/atomic
	// 方案：存储 *int64 指针，用 atomic 操作
	valObj, loaded := sw.buckets.GetOrStore(idx, new(int64))
	atomic.AddInt64(valObj, 1)

	if !loaded {
		sw.keys.Append(idx)
	}
}

// countRequestsWithinWindow 仍然有弱一致性，但对于限流场景可以接受
// 限流本身就是近似值，不需要强一致性
func (sw *SlidingWindowLimiter) countRequestsWithinWindow(currentIdx int) int {
	ret := 0
	for i := currentIdx; i >= 0 && i > currentIdx-sw.buckets_count; i-- {
		if countPtr, ok := sw.buckets.Get(i); ok {
			ret += int(atomic.LoadInt64(countPtr))
		}
	}
	return ret
}
