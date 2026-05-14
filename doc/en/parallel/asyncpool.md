# AsyncPool

## Overview

`AsyncPool` is an asynchronous task pool that manages concurrent task execution with a buffer and worker pool. It combines the benefits of worker pools with an external task buffer, allowing for non-blocking task submission.

## Features

- **Worker Pool**: Fixed number of concurrent workers
- **Task Buffer**: External buffer for queuing tasks when all workers are busy
- **Graceful Shutdown**: Supports orderly shutdown when pool is closed
- **Panic Recovery**: Built-in panic handling for task execution
- **Non-blocking Submit**: Returns immediately if buffer is full

## Architecture

```
                    ┌─────────────────┐
                    │   Task Buffer   │ (with capacity limit)
                    │   (channel)     │
                    └────────▲────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
    ┌────┴────┐        ┌────┴────┐        ┌────┴────┐
    │ Worker1 │        │ Worker2 │   ...  │ WorkerN │
    └─────────┘        └─────────┘        └─────────┘
```

## API Reference

### `NewAsyncPool(taskBufferCapacity int, numWorkers int) (*AsyncPool, error)`

Creates a new AsyncPool with the specified buffer capacity and number of workers.

```go
pool, err := pool.NewAsyncPool(100, 10) // 100 buffer capacity, 10 workers
if err != nil {
    log.Fatal(err)
}
```

### `(ap *AsyncPool) AsyncSubmit(task Task, onError OnError) error`

Submits a task asynchronously. The task type is `func() error`, and onError type is `func(error)`.

```go
err := pool.AsyncSubmit(
    func() error {
        // Do some work
        return nil
    },
    func(err error) {
        fmt.Printf("Task failed: %v\n", err)
    },
)
if err != nil {
    fmt.Printf("Submit failed: %v\n", err)
}
```

**Behavior:**
- If a worker is available, the task is executed immediately in a goroutine
- If all workers are busy, the task is queued in the buffer
- If the buffer is full, an error is returned
- If the pool is closed, an error is returned

### `(ap *AsyncPool) Shutdown()`

Closes the pool. After calling Shutdown:
- New tasks cannot be submitted
- Existing tasks in the buffer will still be executed
- Workers will exit after completing their current tasks and draining the buffer

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/parallel/pool"
)

func main() {
	// Create an async pool with 100 task buffer and 5 workers
	ap, err := pool.NewAsyncPool(100, 5)
	if err != nil {
		panic(err)
	}

	// Submit 20 tasks
	for i := 0; i < 20; i++ {
		taskID := i
		err := ap.AsyncSubmit(
			func() error {
				fmt.Printf("Task %d started\n", taskID)
				time.Sleep(500 * time.Millisecond)
				fmt.Printf("Task %d completed\n", taskID)
				return nil
			},
			func(err error) {
				fmt.Printf("Task error: %v\n", err)
			},
		)
		if err != nil {
			fmt.Printf("Failed to submit task %d: %v\n", taskID, err)
		}
	}

	// Wait for some time
	time.Sleep(3 * time.Second)

	// Shutdown the pool
	ap.Shutdown()
	fmt.Println("Pool shutdown")
}
```

## Task Type Definitions

```go
type Task func() error
type OnError func(error)
```

## Notes

- **Buffer Full**: When the task buffer is full, `AsyncSubmit` returns an error immediately. Consider increasing buffer capacity or implementing retry logic.
- **Worker Efficiency**: Workers efficiently reuse themselves - after completing a task, they immediately pick up the next task from the buffer if available.
- **Shutdown Order**: `Shutdown()` does not immediately stop workers. Workers will finish their current task and drain the buffer before exiting.
- **Panic Handling**: If a task panics, the worker recovers and continues processing the next task.
- **Thread Safety**: All public methods are safe for concurrent use.