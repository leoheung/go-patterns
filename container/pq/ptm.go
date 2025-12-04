package pq

// priorited task manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

type ScheduledTaskInterface interface {
	DoTask()
	ScheduledTime() time.Time
}

type PriorityScheduledTaskManager[T ScheduledTaskInterface] struct {
	pq       *PriorityQueue[T]
	mu       sync.Mutex
	cond     *sync.Cond
	canceled chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
}

func NewPriorityScheduledTaskManager[T ScheduledTaskInterface]() (*PriorityScheduledTaskManager[T], error) {
	pq, err := NewPriorityQueue(0, func(a, b T) bool {
		return a.ScheduledTime().Before(b.ScheduledTime())
	})
	if err != nil {
		return nil, err
	}

	ret := PriorityScheduledTaskManager[T]{
		pq:       pq,
		mu:       sync.Mutex{},
		canceled: make(chan struct{}, 1),
		stopped:  make(chan struct{}),
	}
	ret.cond = sync.NewCond(&ret.mu)

	go ret.watch()
	return &ret, nil
}

func (ptm *PriorityScheduledTaskManager[T]) watch() {
	for !ptm.isStopped() {
		ptm.mu.Lock()
		for ptm.pq.Len() == 0 {
			// 如果被 FinishAndQuit 唤醒，且队列为空，检查是否停止
			if ptm.isStopped() {
				ptm.mu.Unlock()
				return
			}
			ptm.cond.Wait()
		}

		t, err := ptm.pq.Peek()
		if err != nil || utils.IsNil(t) {
			ptm.mu.Unlock()
			continue
		}

		toSleep := t.ScheduledTime().UTC().Sub(time.Now().UTC())
		if toSleep > 0 {
			ptm.mu.Unlock()
			timer := time.NewTimer(toSleep)
			select {
			case <-timer.C:
				ptm.mu.Lock()
				t, err := ptm.pq.Dequeue()
				if err != nil || utils.IsNil(t) {
					ptm.mu.Unlock()
					continue
				}
				t.DoTask()
				if ptm.pq.Len() == 0 {
					ptm.cond.Broadcast()
				}
				ptm.mu.Unlock()
			case <-ptm.canceled:
				// 有新的更早任务来到, 中断timer
				if !timer.Stop() {
					// 非阻塞地吸干，避免阻塞
					select {
					case <-timer.C:
					default:
					}
				}
			}
			continue
		} else {
			// 补做遗留的任务
			t, err := ptm.pq.Dequeue()
			if err != nil || utils.IsNil(t) {
				ptm.mu.Unlock()
				continue
			}
			t.DoTask()
			if ptm.pq.Len() == 0 {
				ptm.cond.Broadcast()
			}
			ptm.mu.Unlock()
		}

	}
}

/* exposed public functions */

func (ptm *PriorityScheduledTaskManager[T]) PendNewTask(t T) error {
	if utils.IsNil(t) {
		return fmt.Errorf("t is nil")
	}

	// 第一次检查（快速失败，避免不必要的锁竞争）
	if ptm.isStopped() {
		return fmt.Errorf("PTM is already stopped")
	}

	ptm.mu.Lock()
	defer ptm.mu.Unlock()

	// 【新增】第二次检查（关键！防止在获取锁的间隙 stopped 被关闭）
	if ptm.isStopped() {
		return fmt.Errorf("PTM is already stopped")
	}

	// 记录入队前长度，判断是否为第一次入队
	lenBefore := ptm.pq.Len()
	if err := ptm.pq.Enqueue(t); err == nil {
		// 唤醒正在等待的 watch
		ptm.cond.Broadcast()

		// 只有当队列原本非空且新元素成为队首时，才发 canceled 中断正在等待的 timer
		if lenBefore > 0 {
			if head, err := ptm.pq.Peek(); err == nil && head.ScheduledTime().Equal(t.ScheduledTime()) {
				select {
				case ptm.canceled <- struct{}{}:
				default:
				}
			}
		}

		return nil
	} else {
		return err
	}
}

func (ptm *PriorityScheduledTaskManager[T]) FinishAndQuit() error {
	ptm.mu.Lock()
	defer ptm.mu.Unlock()

	// 等待队列清空
	for ptm.pq.Len() > 0 {
		ptm.cond.Wait()
	}

	// 只有第一次调用会执行，第二次调用直接跳过，不会 panic
	ptm.stopOnce.Do(func() {
		close(ptm.stopped)
		ptm.cond.Broadcast()
	})

	return nil
}

func (ptm *PriorityScheduledTaskManager[T]) isStopped() bool {
	select {
	case <-ptm.stopped:
		return true
	default:
		return false
	}
}
