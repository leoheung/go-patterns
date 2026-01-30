# Priority Queue & Scheduler

Generic priority queue implementation and time-based task scheduling manager.

## Installation

```go
import "github.com/leoheung/go-patterns/container/pq"
```

## 1. Priority Queue

### Create a Queue

```go
// Create a new priority queue with specified capacity and comparison function
// better(a, b) returns true if a should come before b
pq, err := pq.NewPriorityQueue[int](10, func(a, b int) bool { return a < b })
```

### Basic Operations

```go
// Enqueue an item
err := pq.Enqueue(5)

// Dequeue (get the highest priority item)
item, err := pq.Dequeue()

// Peek (get the highest priority item without removing it)
item, err := pq.Peek()

// Get current queue length
length := pq.Len()
```

## 2. Priority Scheduled Task Manager (PTM)

A scheduler used to execute specific tasks at a designated time.

### Create a Manager

```go
ptm, err := pq.NewPriorityScheduledTaskManager()
if err != nil {
    // Handle error
}
```

### Submit a Task

```go
// Submit a task to be executed in 5 seconds
cancel, err := ptm.PendNewTask(func() {
    fmt.Println("Task executing...")
}, time.Now().Add(5 * time.Second))
```

### Stop the Manager

```go
// Wait for all tasks to complete and stop gracefully
err := ptm.FinishAndQuit()
```

## 3. Cancelable Object

The `Cancelable` object returned by `PendNewTask` is used to manage task status.

```go
// Cancel the task
success := cancel.Cancel()

// Recover the task (if not yet executed)
success := cancel.Recover()

// Check if canceled
isCanceled := cancel.IsCanceled()
```

## Complete Example

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/container/pq"
)

func main() {
    // 1. Using the Scheduler
    ptm, _ := pq.NewPriorityScheduledTaskManager()

    // Submit a task
    cancel, _ := ptm.PendNewTask(func() {
        fmt.Println("This is a delayed task")
    }, time.Now().Add(1 * time.Second))

    // Decide to cancel it later
    if cancel.Cancel() {
        fmt.Println("Task successfully canceled")
    }

    // 2. Using the Priority Queue
    queue, _ := pq.NewPriorityQueue[string](5, func(a, b string) bool {
        return len(a) < len(b) // Shortest string first
    })

    queue.Enqueue("apple")
    queue.Enqueue("go")
    queue.Enqueue("banana")

    item, _ := queue.Dequeue()
    fmt.Printf("Highest priority (shortest): %s\n", item) // Output: go

    // Stop the scheduler
    ptm.FinishAndQuit()
}
```

## Features

- **Generic Support**: `PriorityQueue` works with any data type.
- **Precise Scheduling**: `PriorityScheduledTaskManager` uses a priority queue internally to manage execution times, ensuring the next task is always triggered on time.
- **Task Control**: Each scheduled task has an independent `Cancelable` controller.
- **Thread-safe**: All operations are protected by mutexes.
