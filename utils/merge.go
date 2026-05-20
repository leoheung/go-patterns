package utils

import "context"

func MergeContexts(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc) {
	// Fast path: ctx1 已取消
	select {
	case <-ctx1.Done():
		// 返回一个已取消的 context，cancel 是空操作
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消
		return ctx, cancel
	default:
	}

	// Fast path: ctx2 已取消
	select {
	case <-ctx2.Done():
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		return ctx, cancel
	default:
	}

	ctx, cancel := context.WithCancel(ctx1)

	stop := context.AfterFunc(ctx2, func() {
		cancel()
	})

	return ctx, func() {
		stop()
		cancel()
	}
}
