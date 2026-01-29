package pq

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// 内部使用的包装结构体，不再暴露给外部
type scheduledTask struct {
	Action      func()
	RunAt       time.Time
	TaskCanceld chan struct{}
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
		return a.RunAt.Before(b.RunAt)
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
		toSleep := t.RunAt.UTC().Sub(time.Now().UTC())
		if toSleep > 0 {
			ptm.mu.Unlock()
			timer := time.NewTimer(toSleep)
			select {
			case <-t.TaskCanceld:
				ptm.mu.Lock()
				// 再次檢查隊頭是否仍是 t，防止因為插入了更早的任務導致 Dequeue 錯誤的任務
				head, _ := ptm.pq.Peek()
				if head == t {
					_, err := ptm.pq.Dequeue()
					if err != nil {
						ptm.mu.Unlock()
						continue
					}
					if ptm.pq.Len() == 0 {
						ptm.cond.Broadcast()
					}
				} else {
					// 隊頭變了，說明有新任務插入。
					// 我們無法從 PQ 中間移除 t，只能把取消訊號放回（因為 buffer 為 1，非阻塞）
					// 等 t 再次浮到隊頭時再處理
					select {
					case t.TaskCanceld <- struct{}{}:
					default:
					}
				}
				ptm.mu.Unlock()

			case <-timer.C:
				ptm.mu.Lock()
				t, err := ptm.pq.Dequeue()
				if err != nil || t == nil {
					ptm.mu.Unlock()
					continue
				}
				// 执行闭包
				t.Action()
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
			// 執行閉包
			// 最後一次檢查是否被取消
			select {
			case <-t.TaskCanceld:
				// 被取消了，不執行
			default:
				t.Action()
			}
			ptm.mu.Unlock()
		}
	}
}

/* exposed public functions */

// API 变更：直接传入函数和时间
func (ptm *PriorityScheduledTaskManager) PendNewTask(action func(), runAt time.Time) (chan struct{}, error) {
	if action == nil {
		return nil, fmt.Errorf("action is nil")
	}

	if ptm.isStopped() {
		return nil, fmt.Errorf("PTM is already stopped")
	}

	ptm.mu.Lock()
	defer ptm.mu.Unlock()

	if ptm.isStopped() {
		return nil, fmt.Errorf("PTM is already stopped")
	}

	// 内部包装
	task := &scheduledTask{
		Action:      action,
		RunAt:       runAt,
		TaskCanceld: make(chan struct{}, 1),
	}

	lenBefore := ptm.pq.Len()
	if err := ptm.pq.Enqueue(task); err == nil {
		ptm.cond.Broadcast()

		if lenBefore > 0 {
			// 比较逻辑变更
			if head, err := ptm.pq.Peek(); err == nil && head.RunAt.Equal(task.RunAt) {
				select {
				case ptm.canceled <- struct{}{}:
				default:
				}
			}
		}

		return task.TaskCanceld, nil
	} else {
		return nil, err
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

func (ptm *PriorityScheduledTaskManager) GetAllTasks() []scheduledTask {
	ptm.mu.Lock()
	defer ptm.mu.Unlock()

	// 复制内部数据为值副本，避免泄露内部状态
	src := ptm.pq.data // type: []*scheduledTask
	out := make([]scheduledTask, len(src))
	for i, p := range src {
		if p != nil {
			out[i] = *p
		}
	}
	return out
}

func (ptm *PriorityScheduledTaskManager) isStopped() bool {
	select {
	case <-ptm.stopped:
		return true
	default:
		return false
	}
}

func (ptm *PriorityScheduledTaskManager) String() string {
	var ret strings.Builder
	tasks := ptm.GetAllTasks()
	fmt.Fprintf(&ret, "total %d scheduled tasks", len(tasks))
	for idx, t := range tasks {
		isCanceled := false
		select {
		case <-t.TaskCanceld:
			isCanceled = true
		default:
		}
		fmt.Fprintf(&ret, "scheduled task %d: runAt: %s, isCanceled: %v\n", (idx + 1), t.RunAt.String(), isCanceled)
	}
	return ret.String()
}
