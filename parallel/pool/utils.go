package pool

import "context"

type Task func() error
type OnError func(error)
type TaskWithCtx func(ctx context.Context) error


