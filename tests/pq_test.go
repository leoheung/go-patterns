package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/leoheung/go-patterns/container/pq"
)

type TaskA struct {
	ts   time.Time
	done chan struct{}
}

func (t *TaskA) DoTask()                  { close(t.done) }
func (t *TaskA) ScheduledTime() time.Time { return t.ts }

type TaskB struct {
	ts   time.Time
	done chan struct{}
}

func (t *TaskB) DoTask()                  { close(t.done) }
func (t *TaskB) ScheduledTime() time.Time { return t.ts }

func TestMixedTypes(t *testing.T) {
	ptm, err := pq.NewPriorityScheduledTaskManager[pq.ScheduledTaskInterface]()
	if err != nil {
		t.Fatal(err)
	}

	// 等待两个任务完成
	var wg sync.WaitGroup
	wg.Add(2)

	aDone := make(chan struct{})
	bDone := make(chan struct{})

	_ = ptm.PendNewTask(&TaskA{ts: time.Now().Add(100 * time.Millisecond), done: aDone})
	_ = ptm.PendNewTask(&TaskB{ts: time.Now().Add(200 * time.Millisecond), done: bDone})

	// 通过通道或 WaitGroup 等待任务完成
	go func() { <-aDone; wg.Done() }()
	go func() { <-bDone; wg.Done() }()

	// 等待最多 1s，避免死等
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		// success
	case <-time.After(time.Second):
		t.Fatal("tasks did not run in time")
	}
}
