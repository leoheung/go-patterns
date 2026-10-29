package subscribe

import (
	lctx "github.com/leoheung/go-patterns/container/context"
	"github.com/leoheung/go-patterns/container/safeslice"
	"github.com/leoheung/go-patterns/utils"
)

type ChanProvider[T any] struct {
	subscribers *safeslice.SafeSlice[chan T]
	src         SourceFn[T]
	ctx         *lctx.RenewableContext[any]
}

// NewChanProviderFromSource 共享总线版：pipeline 各节点传同一个 *RenewableContext，
// 任一节点 Cancel/Reactivate 全链联动（含上游 Source 的拉取）。
func NewChanProviderFromSource[T any](ctx *lctx.RenewableContext[any], src SourceFn[T]) *ChanProvider[T] {
	ret := &ChanProvider[T]{
		src:         src,
		subscribers: safeslice.NewSafeSlice[chan T](0, 0),
		ctx:         ctx,
	}

	go ret.notify_worker()

	return ret
}

// NewChanProvider 旧签名保留：内部自建 ctx + channel 适配（零破坏）
func NewChanProvider[T any](c <-chan T) *ChanProvider[T] {
	return NewChanProviderFromSource(
		lctx.NewRenewableContext[any](nil, nil),
		ChanSource[T](c),
	)
}

func (v *ChanProvider[T]) Subscribe(buffer int) (<-chan T, func(), error) {
	// 订阅不受 paused 影响：注册接收者与数据流状态无关，
	// 暂停期间订阅的 channel 在 Resume 后自动接收后续值
	if buffer < 0 {
		buffer = 0
	}
	ch := make(chan T, buffer)
	v.subscribers.Append(ch)

	unsubscribe := func() {
		removed_count := v.subscribers.RemoveIf(func(c chan T) bool { return c == ch })
		if removed_count == 1 {
			close(ch)
		}
	}

	return ch, unsubscribe, nil
}

func (v *ChanProvider[T]) Pause() {
	v.ctx.Cancel()
}

func (v *ChanProvider[T]) Resume() error {
	return v.ctx.Reactivate(nil)
}

func (v *ChanProvider[T]) notify_worker() {
	resume_ch := v.ctx.SubscribeReactivation()

	for {
		// ① 每轮先检查暂停（必须在 ③ 之前）——Cancel 后此处立即检出
		_, paused := utils.TryDequeue(v.ctx.Done())
		if paused {
			<-resume_ch // ② 暂停 → 阻塞等待恢复，绝不调 fetch
		}

		// ③ 只有 ctx 活着才调 Source；fetch 内部 select 的正是同一个 ctx.Done()，
		//    暂停瞬间 in-flight fetch 立即中断（前提：fetch 尊重 ctx，见 Source 契约）
		val, err := v.src(v.ctx)
		if err != nil {
			continue // ctx 已结束 → 回循环头，①检出暂停 → ②等待
		}

		v.subscribers.Range(func(index int, item chan T) bool {
			utils.TryEnqueue(item, val)
			return true
		})

	}
}
