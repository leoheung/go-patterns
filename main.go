package main

import (
	"time"

	"github.com/leoheung/go-patterns/container/pq"
	"github.com/leoheung/go-patterns/utils"
)

type Task struct{
	st time.Time
}

func (t *Task) ScheduledTime() time.Time {
	return  t.st
}

func (t *Task) DoTask()  {
	utils.PrintlnColor(utils.BrightBlue, time.Now().String())
}

func main()  {
	ptm,_ := pq.NewPriorityScheduledTaskManager[*Task]()

	ptm.PendNewTask(&Task{
		st: time.Now().Add(5*time.Second),
	})
	

	ptm.PendNewTask(&Task{
		st: time.Now().Add(time.Second),
	})

	

	select{}
}