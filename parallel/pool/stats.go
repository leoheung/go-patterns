package pool

// Stats 保存调度器的统计信息
type Stats struct {
	Running   *int64 // 当前运行中的任务数
	Pending   *int64 // 等待中的任务数
	Completed *int64 // 已成功完成的任务数
	Failed    *int64 // 执行失败（包括超时）的任务数
}

func NewStats() *Stats {
	return &Stats{
		Running:   new(int64),
		Pending:   new(int64),
		Completed: new(int64),
		Failed:    new(int64),
	}
}
