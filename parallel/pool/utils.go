package pool

type Task func() error
type OnError func(error)