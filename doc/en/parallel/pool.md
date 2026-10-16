# Pools

The `pool` package provides three concurrency-oriented pools:

- **WorkerPool** — semaphore-based worker pool for precise concurrency control
- **AsyncPoolV2** — context-aware async task pool with a bounded task queue and stats
- **ObjPool** — generic object pool to avoid frequent allocation/GC

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/pool"
```

## Task Types

```go
type Task func() error            // for WorkerPool
type OnError func(error)          // error callback (invoked when a task returns an error)
type TaskWithCtx func(ctx context.Context) error  // for AsyncPoolV2
```

## WorkerPool

A lightweight worker pool based on a semaphore, for limiting concurrent goroutines.

### Create a Worker Pool

```go
// Create a worker pool with a fixed number of workers (must be > 0)
p := pool.NewWorkerPool(5)
```

### Submit Tasks

```go
// Blocking submit: waits for an available worker
p.Submit(func() error {
    // Task logic
    return nil
}, func(err error) {
    // Optional error callback
})

// Non-blocking submit: returns false immediately if no worker is available
success := p.TrySubmit(func() error {
    // Task logic
    return nil
}, nil)
```

### Complete Example

```go
package main

import (
    "fmt"
    "sync/atomic"
    "time"

    "github.com/leoheung/go-patterns/parallel/pool"
)

func main() {
    // Create a pool with 3 workers
    p := pool.NewWorkerPool(3)

    var counter int32 = 0

    // Submit 10 tasks
    for i := 0; i < 10; i++ {
        taskID := i
        p.Submit(func() error {
            fmt.Printf("Task %d started\n", taskID)
            time.Sleep(100 * time.Millisecond)
            atomic.AddInt32(&counter, 1)
            fmt.Printf("Task %d completed\n", taskID)
            return nil
        }, nil)
    }

    // Wait for tasks to complete
    time.Sleep(1 * time.Second)

    fmt.Printf("Total completed: %d\n", counter)
}
```

### WorkerPool Features

- **Concurrency Control**: Limits the number of goroutines running simultaneously.
- **Panic Protection**: Automatically recovers from panics within tasks and logs the error.
- **Error Handling**: Tasks return `error`; optional `onError` callback is invoked on failure.
- **Flexible Submission**: Supports both blocking (`Submit`) and non-blocking (`TrySubmit`).
- **Resource Efficient**: Built on `SemaphoreByCond` for minimal overhead.

## AsyncPoolV2

A context-aware async task pool with a bounded task queue. Workers merge the caller's context with the pool's own context via `utils.MergeContexts`, so both task-level cancellation and pool shutdown are honored.

### Create a Pool

```go
// maxWorkers >= 0, queueSize >= 0 (otherwise nil is returned)
p := pool.NewAsyncPoolV2(4, 64)
```

### Submit a Task

```go
err := p.AsyncSubmit(ctx, func(ctx context.Context) error {
    // Task logic with context support
    return nil
}, func(err error) {
    // Optional error callback
})
```

- If a worker is free, the task starts immediately; otherwise it is enqueued.
- If the queue is full, an error is returned (`task queue if full`).
- Submitting after `Shutdown` returns an error.

### Stats & Shutdown

```go
s := p.Stats()   // { Running, Pending, Completed, Failed }

p.Shutdown()     // cancel the pool context; workers stop after current tasks
```

### AsyncPoolV2 Example

```go
package main

import (
    "context"
    "fmt"

    "github.com/leoheung/go-patterns/parallel/pool"
)

func main() {
    p := pool.NewAsyncPoolV2(4, 64)
    defer p.Shutdown()

    for i := 0; i < 10; i++ {
        taskID := i
        err := p.AsyncSubmit(context.Background(), func(ctx context.Context) error {
            fmt.Printf("Task %d running\n", taskID)
            return nil
        }, func(err error) {
            fmt.Printf("Task error: %v\n", err)
        })
        if err != nil {
            fmt.Printf("submit error: %v\n", err)
        }
    }

    fmt.Printf("stats: %+v\n", p.Stats())
}
```

### AsyncPoolV2 Features

- **Context propagation**: each task runs with a merged context (task ctx + pool ctx).
- **Bounded queue**: surplus tasks are queued; a full queue rejects submissions with an error.
- **Stats**: live counts of running/pending/completed/failed tasks.
- **Graceful shutdown**: `Shutdown` cancels the pool context.

## ObjPool

A generic object pool with a fixed size, to avoid frequent allocation and GC clearing. Unlike `sync.Pool`, objects are **not** cleared by GC.

### Create an Object Pool

```go
// poolSize must be > 0
p, err := pool.NewObjPool(10, func() *Buffer {
    return &Buffer{}
})
```

The pool is pre-filled with `poolSize` objects created by `newFn`.

### Get / Put

```go
// Get an object from the pool (creates a new one if empty)
buf := p.Get()

// Return the object (dropped silently if the pool is full)
p.Put(buf)
```

### ObjPool Example

```go
pool, err := pool.NewObjPool(10, func() *Connection {
    return OpenConnection()
})
if err != nil {
    log.Fatal(err)
}

// Get object from pool
conn := pool.Get()
defer pool.Put(conn)

// Use conn...
```

### ObjPool Features

- **Generic**: works with any type `T`.
- **Pre-filled**: initialized with `poolSize` pre-created objects.
- **Non-blocking**: `Get` creates a new object when empty; `Put` drops when full.
- **GC-friendly**: objects are never cleared by GC — ideal for reuse hot paths.
- **Thread-safe**: `Get`/`Put` can be called concurrently.
