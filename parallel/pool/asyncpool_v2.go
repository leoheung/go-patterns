package pool

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/leoheung/go-patterns/utils"
)

type combinedTask struct {
	task    TaskWithCtx
	ctx     context.Context
	onError OnError
}

type AsyncPoolV2 struct {
	maxWorkers    *int32
	taskqueue     chan combinedTask
	stats         *Stats
	ctx           context.Context
	cancelFn      context.CancelFunc
	allTasksCount *int64
}

func NewAsyncPoolV2(maxWorkers int32, queueSize int) *AsyncPoolV2 {
	if maxWorkers < 0 || queueSize < 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &AsyncPoolV2{
		maxWorkers:    &maxWorkers,
		taskqueue:     make(chan combinedTask, queueSize),
		stats:         NewStats(),
		ctx:           ctx,
		cancelFn:      cancel,
		allTasksCount: new(int64),
	}
}

func (ts *AsyncPoolV2) Shutdown() {
	ts.cancelFn()
}

func (ts *AsyncPoolV2) AsyncSubmit(ctx context.Context, task TaskWithCtx, onError OnError) error {

	if ts.isClosed() {
		return fmt.Errorf("TaskScheduler is closed")
	}

	if atomic.AddInt32(ts.maxWorkers, -1) >= 0 {
		atomic.AddInt64(ts.allTasksCount, 1)
		go ts.work(ctx, task, onError)
	} else {
		queueFull := !utils.TryEnqueue(ts.taskqueue, combinedTask{
			task:    task,
			ctx:     ctx,
			onError: onError,
		})
		if queueFull {
			return fmt.Errorf("task queue if full")
		} else {
			atomic.AddInt64(ts.allTasksCount, 1)
		}
	}
	return nil
}

func (ts *AsyncPoolV2) Stats() any {
	type stats struct {
		Running   int64 `json:"running"`
		Pending   int   `json:"pending"`
		Completed int64 `json:"completed"`
		Failed    int64 `json:"failed"`
	}
	pending := len(ts.taskqueue)
	completed := atomic.LoadInt64(ts.stats.Completed)
	failed := atomic.LoadInt64(ts.stats.Failed)
	running := atomic.LoadInt64(ts.allTasksCount) - int64(pending) - completed - failed

	return stats{
		Running:   running,
		Pending:   pending,
		Completed: completed,
		Failed:    failed,
	}
}

func (ts *AsyncPoolV2) work(ctx context.Context, task TaskWithCtx, onError OnError) {
	mergedCtx, cancel := utils.MergeContexts(ctx, ts.ctx)
	err := ts.execTask(mergedCtx, task, onError)
	cancel()
	ts.updateStatsAtomicly(err)

	for {
		select {
		case <-ts.ctx.Done():
			return
		case newTask := <-ts.taskqueue:
			mergedCtx, cancel = utils.MergeContexts(newTask.ctx, ts.ctx)
			err := ts.execTask(mergedCtx, newTask.task, newTask.onError)
			cancel()
			ts.updateStatsAtomicly(err)
		}
	}
}

func (ts *AsyncPoolV2) updateStatsAtomicly(err error) {
	if err != nil {
		atomic.AddInt64(ts.stats.Failed, 1)
	} else {
		atomic.AddInt64(ts.stats.Completed, 1)
	}
}

func (ts *AsyncPoolV2) isClosed() bool {
	_, closed := utils.TryDequeue(ts.ctx.Done())
	return closed
}

func (ts *AsyncPoolV2) execTask(ctx context.Context, task TaskWithCtx, onError OnError) error {
	err := task(ctx)
	if err != nil && onError != nil {
		onError(err)
	}
	return err
}
