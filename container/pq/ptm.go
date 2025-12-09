package pq

import (
	"fmt"
	"sync"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

// 内部使用的包装结构体，不再暴露给外部
type scheduledTask struct {
	action func()
	runAt  time.Time
}

// 移除泛型 [T]
type PriorityScheduledTaskManager struct {
	pq       *PriorityQueue[*scheduledTask] // 内部存储具体的包装类型
	mu       sync.Mutex
	cond     *sync.Cond
	canceled chan struct{}
	stopped  chan struct{}
	stopOnce sync.Once
}

// 构造函数不再需要类型参数
func NewPriorityScheduledTaskManager() (*PriorityScheduledTaskManager, error) {
	// 比较逻辑改为比较内部结构体的 runAt
	pq, err := NewPriorityQueue(0, func(a, b *scheduledTask) bool {
		return a.runAt.Before(b.runAt)
	})
	if err != nil {
		return nil, err
	}

	ret := PriorityScheduledTaskManager{
		pq:       pq,
		mu:       sync.Mutex{},
		canceled: make(chan struct{}, 1),
		stopped:  make(chan struct{}),
	}
	ret.cond = sync.NewCond(&ret.mu)

	go ret.watch()
	return &ret, nil
}

func (ptm *PriorityScheduledTaskManager) watch() {
	for !ptm.isStopped() {
		ptm.mu.Lock()
		for ptm.pq.Len() == 0 {
			if ptm.isStopped() {
				ptm.mu.Unlock()
				return
			}
			ptm.cond.Wait()
		}

		t, err := ptm.pq.Peek()
		if err != nil || t == nil { // t 是 *scheduledTask，直接判空即可
			ptm.mu.Unlock()
			continue
		}

		// 使用 t.runAt
		toSleep := t.runAt.UTC().Sub(time.Now().UTC())
		if toSleep > 0 {
			ptm.mu.Unlock()
			timer := time.NewTimer(toSleep)
			select {
			case <-timer.C:
				ptm.mu.Lock()
				t, err := ptm.pq.Dequeue()
				if err != nil || t == nil {
					ptm.mu.Unlock()
					continue
				}
				// 执行闭包
				t.action()
				if ptm.pq.Len() == 0 {
					ptm.cond.Broadcast()
				}
				ptm.mu.Unlock()
			case <-ptm.canceled:
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
			}
			continue
		} else {
			t, err := ptm.pq.Dequeue()
			if err != nil || t == nil {
				ptm.mu.Unlock()
				continue
			}
			// 执行闭包
			t.action()
			if ptm.pq.Len() == 0 {
				ptm.cond.Broadcast()
			}
			ptm.mu.Unlock()
		}
	}
}

/* exposed public functions */

// API 变更：直接传入函数和时间
func (ptm *PriorityScheduledTaskManager) PendNewTask(action func(), runAt time.Time) error {
	if action == nil {
		return fmt.Errorf("action is nil")
	}

	if ptm.isStopped() {
		return fmt.Errorf("PTM is already stopped")
	}

	ptm.mu.Lock()
	defer ptm.mu.Unlock()

	if ptm.isStopped() {
		return fmt.Errorf("PTM is already stopped")
	}

	// 内部包装
	task := &scheduledTask{
		action: action,
		runAt:  runAt,
	}

	lenBefore := ptm.pq.Len()
	if err := ptm.pq.Enqueue(task); err == nil {
		ptm.cond.Broadcast()

		if lenBefore > 0 {
			// 比较逻辑变更
			if head, err := ptm.pq.Peek(); err == nil && head.runAt.Equal(task.runAt) {
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

func (ptm *PriorityScheduledTaskManager) FinishAndQuit() error {
	ptm.mu.Lock()
	defer ptm.mu.Unlock()

	for ptm.pq.Len() > 0 {
		ptm.cond.Wait()
	}

	ptm.stopOnce.Do(func() {
		close(ptm.stopped)
		ptm.cond.Broadcast()
	})

	return nil
}

func (ptm *PriorityScheduledTaskManager) isStopped() bool {
	select {
	case <-ptm.stopped:
		return true
	default:
		return false
	}
}
