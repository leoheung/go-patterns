# Priority Queue

A generic priority queue implementation with customizable priority comparison.

## Installation

```go
import "github.com/leoxiang66/go-patterns/container/pq"
```

## API Reference

### Create a Priority Queue

```go
// Create a new priority queue with specified capacity and comparison function
// better(a, b) returns true if a should come before b
pq, err := pq.NewPriorityQueue[int](10, func(a, b int) bool { return a < b })
```

### Enqueue

```go
// Enqueue an item
err := pq.Enqueue(5)
```

### Dequeue

```go
// Dequeue the highest priority item
item, err := pq.Dequeue()
```

### Peek

```go
// Get the highest priority item without removing it
item, err := pq.Peek()
```

### Queue Length

```go
// Get current queue length
length := pq.Len()
```

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/container/pq"
)

type Task struct {
    ID       int
    Priority int
    Name     string
}

func main() {
    // Create a priority queue where higher priority tasks come first
    pq, err := pq.NewPriorityQueue[Task](5, func(a, b Task) bool {
        return a.Priority > b.Priority
    })
    if err != nil {
        fmt.Printf("Error creating priority queue: %v\n", err)
        return
    }

    // Add tasks with different priorities
    tasks := []Task{
        {ID: 1, Priority: 3, Name: "Task 1"},
        {ID: 2, Priority: 1, Name: "Task 2"},
        {ID: 3, Priority: 5, Name: "Task 3"},
        {ID: 4, Priority: 2, Name: "Task 4"},
        {ID: 5, Priority: 4, Name: "Task 5"},
    }

    for _, task := range tasks {
        if err := pq.Enqueue(task); err != nil {
            fmt.Printf("Error enqueueing task %d: %v\n", task.ID, err)
        }
    }

    // Process tasks in priority order
    for pq.Len() > 0 {
        task, err := pq.Dequeue()
        if err != nil {
            fmt.Printf("Error dequeuing task: %v\n", err)
            continue
        }
        fmt.Printf("Processing Task %d: %s (Priority: %d)\n", task.ID, task.Name, task.Priority)
    }
}
```

## Output

```
Processing Task 3: Task 3 (Priority: 5)
Processing Task 5: Task 5 (Priority: 4)
Processing Task 1: Task 1 (Priority: 3)
Processing Task 4: Task 4 (Priority: 2)
Processing Task 2: Task 2 (Priority: 1)
```

## Features

- **Generic support**: Works with any type
- **Custom comparison**: Define your own priority logic
- **Binary heap**: Efficient O(log n) enqueue/dequeue operations
- **Type-safe**: Compile-time type checking
