package fluent

import "sync"

// ═══════════════════════════════════════════════════════
// Core Type
// ═══════════════════════════════════════════════════════

// Pipeline 流式管道，any 为当前流经的数据类型
type Pipeline struct {
	ch   chan any        // 当前阶段的输出通道
	quit <-chan struct{} // 全局终止信号（内部传播）
}

// ═══════════════════════════════════════════════════════
// Construction
// ═══════════════════════════════════════════════════════

// New 创建一个空管道，返回管道和可写的输入通道
//
//	p, in := New[int]()
//	result := p.Transform(...).Run()
//	in <- 42; close(in)

func NewPipeline(quit_ch <-chan struct{}, input_ch chan any) *Pipeline {
	return &Pipeline{
		ch:   input_ch,
		quit: quit_ch,
	}
}

// FromSlice 从 slice 创建管道（内部自动推送并关闭）
func NewPipelineFromSlice(quit_ch <-chan struct{}, items []any) *Pipeline {
	input_ch := make(chan any, len(items))
	for _, v := range items {
		input_ch <- v
	}
	close(input_ch)
	return &Pipeline{
		ch:   input_ch,
		quit: quit_ch,
	}
}

// ═══════════════════════════════════════════════════════
// Transform
// ═══════════════════════════════════════════════════════

// Transform 类型转换 any → any，返回 Pipeline
//
//	From[int](ch).
//	    Transform(func(x int) string { return strconv.Itoa(x) }).
//	    Transform(func(s string) int { n, _ := strconv.Atoi(s); return n })
func (p *Pipeline) Transform(fn func(any) any) *Pipeline {
	next := Pipeline{
		ch:   make(chan any),
		quit: p.quit,
	}
	go func() {
		defer close(next.ch)
		for {
			select {
			case <-next.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				y := fn(x)
				next.ch <- y
			}
		}
	}()
	return &next
}

// FlatMap 一对多展开 any → []any，展平为 Pipeline
func (p *Pipeline) FlatMap(fn func(any) []any) *Pipeline {
	next := Pipeline{
		ch:   make(chan any),
		quit: p.quit,
	}
	go func() {
		defer close(next.ch)
		for {
			select {
			case <-next.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				for _, item := range fn(x) {
					next.ch <- item
				}
			}
		}
	}()
	return &next
}

// ═══════════════════════════════════════════════════════
// Filter & Side Effect
// ═══════════════════════════════════════════════════════

// Filter 过滤，仅保留满足条件的元素
func (p *Pipeline) Filter(pred func(any) bool) *Pipeline {
	next := Pipeline{
		quit: p.quit,
		ch:   make(chan any),
	}

	go func() {
		defer close(next.ch)
		for {
			select {
			case <-next.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				if pred(x) {
					next.ch <- x
				}
			}
		}
	}()

	return &next
}

// Tap 副作用（如日志、metrics），不改变数据，原样传递
func (p *Pipeline) Tap(fn func(any)) *Pipeline {
	next := Pipeline{
		quit: p.quit,
		ch:   make(chan any),
	}

	go func() {
		defer close(next.ch)
		for {
			select {
			case <-next.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				fn(x)
				next.ch <- x
			}
		}
	}()

	return &next
}

// ═══════════════════════════════════════════════════════
// Flow Control
// ═══════════════════════════════════════════════════════

// Split 按条件分流为两个独立 Pipeline，各自可继续链式构建
//
//	trueBranch 走 pred == true 的数据
//	falseBranch 走 pred == false 的数据
func (p *Pipeline) Split(pred func(any) bool) (trueBranch, falseBranch *Pipeline) {
	nextTrue := Pipeline{
		quit: p.quit,
		ch:   make(chan any),
	}

	nextFalse := Pipeline{
		quit: p.quit,
		ch:   make(chan any),
	}

	go func() {
		defer close(nextTrue.ch)
		defer close(nextFalse.ch)
		for {
			select {
			case <-nextTrue.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				if pred(x) {
					nextTrue.ch <- x
				} else {
					nextFalse.ch <- x
				}
			}
		}
	}()

	return &nextTrue, &nextFalse
}

// Join 合并多个同类型 Pipeline 为一个 (fan in)
func (p *Pipeline) Join(others ...*Pipeline) *Pipeline {
	next := Pipeline{
		quit: p.quit,
		ch:   make(chan any),
	}

	go func() {
		defer close(next.ch)
		var wg sync.WaitGroup
		wg.Add(1 + len(others))

		// 处理 p
		go func() {
			defer wg.Done()
			for {
				select {
				case <-p.quit:
					return
				case x, ok := <-p.ch:
					if !ok {
						return
					}
					next.ch <- x
				}
			}
		}()

		// 处理 others
		for _, o := range others {
			o := o // 捕获变量
			go func() {
				defer wg.Done()
				for {
					select {
					case <-o.quit:
						return
					case x, ok := <-o.ch:
						if !ok {
							return
						}
						next.ch <- x
					}
				}
			}()
		}

		wg.Wait()
	}()

	return &next
}

// FanIn 合并多个 Pipeline 为一个（与 Join 等价，但接收者不参与）
func FanIn(quit_ch chan struct{}, pipelines ...*Pipeline) *Pipeline {
	if len(pipelines) == 0 {
		return &Pipeline{
			ch:   make(chan any),
			quit: quit_ch,
		}
	}


	next := Pipeline{
		quit: quit_ch,
		ch:   make(chan any),
	}

	go func() {
		defer close(next.ch)
		var wg sync.WaitGroup
		wg.Add(len(pipelines))

		for _, p := range pipelines {
			go func(pipe *Pipeline) {
				defer wg.Done()
				for {
					select {
					case <-quit_ch:
						return
					case x, ok := <-pipe.ch:
						if !ok {
							return
						}
						next.ch <- x
					}
				}
			}(p)
		}

		wg.Wait()
	}()

	return &next
}

// FanOut 争抢模式：每个 item 只给一个下游（负载均衡）
//
//	适合：并行处理，N 个 worker 分担工作量
func (p *Pipeline) FanOut(n int) []*Pipeline {
	ret := make([]*Pipeline, 0, n)
	for range n {
		ret = append(ret, &Pipeline{
			quit: p.quit,
			ch:   make(chan any),
		})
	}

	for _, o := range ret {
		go func(pipe *Pipeline) {
			defer close(pipe.ch)
			for {
				select {
				case <-pipe.quit:
					return

				case x, ok := <-p.ch:
					if !ok {
						return
					}
					pipe.ch <- x
				}
			}
		}(o)
	}
	return ret
}

// 广播：每个 item copy 给所有下游
func (p *Pipeline) Broadcast(n int) []*Pipeline {
	ret := make([]*Pipeline, 0, n)
	for range n {
		ret = append(ret, &Pipeline{
			quit: p.quit,
			ch:   make(chan any),
		})
	}

	go func() {
		defer func() {
			for _, o := range ret {
				close(o.ch)
			}
		}()

		for {
			select {
			case <-p.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				for _, o := range ret {
					o.ch <- x
				}
			}
		}
	}()
	return ret
}

// Parallel 并行处理：N 个 worker 同时执行 fn，结果自动合并
//
//	等价于 FanOut(n) → N × Transform(fn) → FanIn
func (p *Pipeline) Parallel(n int, fn func(any) any) *Pipeline {
	fanouts := p.FanOut(n)
	for i, o := range fanouts {
		fanouts[i] = o.Transform(fn)
	}
	return fanouts[0].Join(fanouts[1:]...)
}

// ═══════════════════════════════════════════════════════
// Rate Limiting
// ═══════════════════════════════════════════════════════

// Take 仅取前 n 个元素，之后自动终止
func (p *Pipeline) Take(n int) *Pipeline {
	next := Pipeline{
		quit: p.quit,
		ch:   make(chan any),
	}

	go func() {
		defer close(next.ch)
		for range n {
			select {
			case <-next.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				next.ch <- x
			}
		}
	}()

	return &next
}

// Skip 跳过前 n 个元素
func (p *Pipeline) Skip(n int) *Pipeline {
	next := Pipeline{
		quit: p.quit,
		ch:   make(chan any),
	}

	go func() {
		defer close(next.ch)
		// 跳过前 n 个
		for range n {
			select {
			case <-next.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				_ = x // 丢弃
			}
		}
		// 传递剩余的
		for {
			select {
			case <-next.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				next.ch <- x
			}
		}
	}()

	return &next
}

// Buffer 设置输出通道缓冲区大小
func (p *Pipeline) Buffer(size int) *Pipeline {
	next := Pipeline{
		quit: p.quit,
		ch:   make(chan any, size),
	}

	go func() {
		defer close(next.ch)
		for {
			select {
			case <-next.quit:
				return
			case x, ok := <-p.ch:
				if !ok {
					return
				}
				next.ch <- x
			}
		}
	}()

	return &next
}

// ═══════════════════════════════════════════════════════
// Terminal
// ═══════════════════════════════════════════════════════

// Run 启动管道，返回输出通道（用 range 消费）
func (p *Pipeline) Run() <-chan any {
	return p.ch
}

// Collect 消费所有输出，返回 slice（阻塞直到管道结束）
func (p *Pipeline) Collect() []any {
	ret := make([]any, 0)
	for x := range p.ch {
		ret = append(ret, x)
	}
	return ret
}

// ForEach 逐个消费输出，执行 fn（阻塞直到管道结束）
func (p *Pipeline) ForEach(fn func(any)) {
	for x := range p.ch {
		fn(x)
	}
}
