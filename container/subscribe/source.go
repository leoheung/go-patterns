package subscribe

import (
	"context"
	"time"
)

// SourceFn 生产策略：阻塞直到产出下一个值。
// 契约（两条，缺一不可）：
//  1. 仅当 ctx 结束（Done）时才返回 error；真实生产错误应在 SourceFn 内部自行重试/退避
//  2. fetch/循环内部必须 select ctx.Done()——暂停/取消时立即返回，
//     否则 provider 无法控制上游（协作式取消的前提）
//
// 注意：对 RenewableContext 适配时，退出路径不要用 ctx.Err() 表达原因——
// 阻塞解除后 ctx 可能已被 Reactivate 重置，Err() 会变回 nil 导致 (零值, nil) 假值
// 被下游广播。统一返回 context.Canceled。
type SourceFn[T any] func(ctx context.Context) (T, error)

// ChanSource 把现成 channel 适配为 Source（兼容旧用法）。
// 上游关闭后阻塞直到 ctx 结束（否则 worker 会空转）。
func ChanSource[T any](c <-chan T) SourceFn[T] {
	return func(ctx context.Context) (T, error) {
		var zero T
		select {
		case v, ok := <-c:
			if !ok {
				// 上游已关闭：永久阻塞，直到 ctx 结束
				<-ctx.Done()
				return zero, context.Canceled
			}
			return v, nil
		case <-ctx.Done():
			return zero, context.Canceled
		}
	}
}

// PollingSource 轮询式 Source：fetch 自身也应尊重 ctx；
// interval 等"生产侧配置"封装在这里，与 provider 解耦
func PollingSource[T any](fetch func(ctx context.Context) (T, error), interval time.Duration) SourceFn[T] {
	return func(ctx context.Context) (T, error) {
		var zero T
		for {
			v, err := fetch(ctx)
			if err == nil {
				return v, nil
			}
			if ctx.Err() != nil {
				return zero, context.Canceled
			}
			select {
			case <-time.After(interval):
			case <-ctx.Done():
				return zero, context.Canceled
			}
		}
	}
}
