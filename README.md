# go-patterns

`go-patterns` is a collection of concurrency patterns and data structures implemented in Go, designed to help developers better understand and utilize Go's concurrency features.

## Project Structure

This repository includes the following modules, each implementing a common concurrency pattern or data structure:

- **container/list**: Implements a generic dynamic array, similar to Python's `list` and JavaScript's `Array`.
- **container/msgQueue**: Provides a channel-based message queue implementation with basic queue operations.
- **container/pq**: Implements a generic priority queue with customizable priority comparison.
- **parallel/barrier**: Provides implementations of Barrier for synchronizing multiple goroutines.
- **parallel/communication**: Implements communication patterns for goroutines.
- **parallel/limiter**: Implements a static limiter for controlling the rate of operations.
- **parallel/mutex**: Implements a simple mutex to ensure that only one goroutine accesses a shared resource at a time.
- **parallel/pipeline**: Implements various pipeline patterns for data processing.
- **parallel/rwlock**: Implements a read-write lock, supporting multiple readers or a single writer for concurrent access.
- **parallel/semaphore**: Implements semaphore patterns to limit the number of goroutines accessing shared resources simultaneously.
- **utils**: Provides utility functions for logging and retrying operations.

## API Documentation

### container/list

A generic dynamic array implementation supporting Python list and JavaScript Array operations.

#### Basic Operations
```go
// Create a new list
l := list.New[int]()

// Create a list from a slice
l := list.From([]int{1, 2, 3})

// Get the length and capacity
length := l.Len()
capacity := l.Cap()

// Convert to slice
slice := l.ToSlice()

// Clone the list
clone := l.Clone()
```

#### Element Access
```go
// Get element by index (supports negative indices)
elem := l.Get(0)
elem := l.Get(-1) // Last element

// Set element by index
l.Set(0, 10)

// Safe element access
if elem, ok := l.At(0); ok {
    // Element exists
}
```

#### Adding Elements
```go
// Append elements to the end
l.Append(4, 5)
l.Push(6) // Alias for Append

// Extend with a slice
l.Extend([]int{7, 8})

// Add elements to the beginning
l.Unshift(0, -1)
```

#### Removing Elements
```go
// Remove and return the first element
if elem, ok := l.Shift(); ok {
    // Handle element
}

// Remove and return the last element
if elem, ok := l.Pop(); ok {
    // Handle element
}

// Remove the first occurrence of a value
l.RemoveFirst(5, func(a, b int) bool { return a == b })

// Remove element at index
if elem, ok := l.RemoveAt(2); ok {
    // Handle element
}

// Clear the list
l.Clear()
```

#### Search and Query
```go
// Check if list contains an element
contains := l.Includes(5, func(a, b int) bool { return a == b })

// Find index of element
index := l.IndexOf(5, func(a, b int) bool { return a == b })
lastIndex := l.LastIndexOf(5, func(a, b int) bool { return a == b })

// Count occurrences
count := l.Count(5, func(a, b int) bool { return a == b })

// Find elements
if elem, ok := l.Find(func(v, i int) bool { return v > 10 }); ok {
    // Handle element
}
```

#### Transformation and Filtering
```go
// Map elements to new list
newList := list.Map(l, func(v, i int) string { return fmt.Sprintf("%d", v) })

// Filter elements
filtered := l.Filter(func(v, i int) bool { return v > 5 })

// Reduce elements
result := list.Reduce(l, 0, func(acc, v, i int) int { return acc + v })
```

#### Sorting and Reversing
```go
// Sort in place
l.Sort(func(a, b int) bool { return a < b })

// Get sorted copy
lSorted := l.ToSorted(func(a, b int) bool { return a < b })

// Reverse in place
l.Reverse()

// Get reversed copy
lReversed := l.ToReversed()
```

### container/msgQueue

A channel-based message queue implementation with basic queue operations.

#### API
```go
// Create a new message queue with specified capacity and device ID
mq := msgqueue.NewChanMQ(100, "device-1")

// Enqueue a message
err := mq.Enq([]byte("hello world"))

// Dequeue a message with context
ctx := context.Background()
msg, err := mq.Deq(ctx)

// Get current queue length
length := mq.Len()

// Clear all messages
err := mq.Clear()

// Check if queue is live
isLive := mq.IsLive()

// Renew a dead queue
mq.Renew()

// Destroy the queue
mq.Destroy()
```

### container/pq

A generic priority queue implementation with customizable priority comparison.

#### API
```go
// Create a new priority queue with specified capacity and comparison function
// better(a, b) returns true if a should come before b
pq, err := pq.NewPriorityQueue[int](10, func(a, b int) bool { return a < b })

// Enqueue an item
err := pq.Enqueue(5)

// Dequeue the highest priority item
item, err := pq.Dequeue()

// Get the highest priority item without removing it
item, err := pq.Peek()

// Get current queue length
length := pq.Len()
```

### parallel/barrier

Synchronization primitive that allows multiple goroutines to wait for each other to reach a certain point.

#### API
```go
// Create a new barrier for N goroutines
barrier := barrier.NewBarrier(5)

// In a goroutine:
barrier.Wait() // Wait for all goroutines to reach this point
```

### parallel/communication

Defines interfaces for communication patterns between goroutines.

#### MessageInterface
```go
// MessageInterface defines basic message operations
type MessageInterface[T any] interface {
    SetMsg(msg T)
    GetMsg() T
    SetTime(time int)
    GetTime() int
}
```

#### ClockInterface
```go
// ClockInterface defines basic clock operations
type ClockInterface interface {
    Tick()      // Advance clock by one unit
    Time() int  // Get current time
}
```

These interfaces provide contracts for implementing message-passing and clock synchronization in concurrent systems.

### parallel/limiter

A static limiter for controlling the rate of operations.

#### API
```go
// Create a new limiter with specified interval
limiter := limiter.NewStaticLimiter(100 * time.Millisecond) // 10 operations per second

// In a goroutine:
limiter.GrantNextToken() // Wait until next token is available
// Perform operation
```

### parallel/mutex

A simple mutual exclusion lock implementation.

#### API
```go
// Create a new mutex
m := mutex.NewMutex()

// Lock the mutex
m.Lock()

// Unlock the mutex
defer m.Unlock()

// Perform protected operations
```

### parallel/pipeline

Pipeline patterns for data processing.

#### API

```go
// AddOnPipe: Generic pipeline node that transforms data from X to Y
// q: quit channel
// f: transformation function
// in: input channel
out := pipeline.AddOnPipe(q, f, in)

// FanIn: Merge multiple input channels into a single output channel
out := pipeline.FanIn(q, input1, input2, input3)

// FanOut: Distribute data from a single input channel to multiple output channels
outs := pipeline.FanOut(q, in, 3) // 3 output channels

// Broadcast: Broadcast data to multiple subscribers
broadcast := pipeline.NewBroadcast(q, in)
subscriber1 := broadcast.Subscribe()
subscriber2 := broadcast.Subscribe()
go broadcast.Run()

// Take: Take the first n elements from input channel
out := pipeline.Take(q, 5, in) // Take first 5 elements
```

#### Example: Simple Pipeline

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/parallel/pipeline"
)

func main() {
    // Create channels
    input := make(chan int)
    quit := make(chan struct{})
    defer close(quit)
    
    // Create pipeline: Square -> Double
    square := func(x int) int { return x * x }
    double := func(x int) int { return x * 2 }
    
    stage1 := pipeline.AddOnPipe(quit, square, input)
    stage2 := pipeline.AddOnPipe(quit, double, stage1)
    
    // Send data
    go func() {
        for i := 1; i <= 5; i++ {
            input <- i
        }
        close(input)
    }()
    
    // Receive results
    for result := range stage2 {
        fmt.Println(result) // Output: 2, 8, 18, 32, 50
    }
}
```

#### Example: FanOut and FanIn

```go
package main

import (
    "fmt"
    "github.com/leoxiang66/go-patterns/parallel/pipeline"
)

func main() {
    // Create channels
    input := make(chan int)
    quit := make(chan struct{})
    defer close(quit)
    
    // FanOut: Distribute data to 3 workers
    workers := pipeline.FanOut(quit, input, 3)
    
    // Process data in parallel
    process := func(x int) int { return x * 2 }
    var processed []chan int
    for _, worker := range workers {
        processed = append(processed, pipeline.AddOnPipe(quit, process, worker))
    }
    
    // FanIn: Merge results from all workers
    output := pipeline.FanIn(quit, processed...)
    
    // Send data
    go func() {
        for i := 1; i <= 5; i++ {
            input <- i
        }
        close(input)
    }()
    
    // Receive results
    for result := range output {
        fmt.Println(result)
    }
}
```

### parallel/rwlock

A read-write lock implementation supporting multiple readers or a single writer.

#### API
```go
// Create a new read-write lock
rw := rwlock.NewRWLock()

// Read lock
rw.RLock()
defer rw.RUnlock()

// Write lock
rw.Lock()
defer rw.Unlock()
```

### parallel/semaphore

A semaphore implementation for limiting concurrent access to resources.

#### API
```go
// Create a new semaphore with specified capacity
sem := semaphore.NewSemaphore(5)

// Acquire a permit
sem.Acquire()
defer sem.Release()

// Perform operation
```

### utils

#### Logging
```go
// Check if in development environment
isDev := utils.IsDev() // Returns true if env=dev

// Log message (uses fmt.Println in dev, log.Println in prod)
utils.LogMessage("Hello, world!")
```

#### Retry
```go
// Retry a function with error/panic handling
// work: Function to execute
// retryTimes: Maximum retry attempts (excluding first try)
utils.RetryWork(
    func() error {
        // Operation that might fail
        return nil // or error
    },
    3, // Retry 3 times if failed
)
```

#### Example Usage
```go
package main

import (
    "github.com/leoxiang66/go-patterns/utils"
    "time"
)

func main() {
    // Set environment to development
    // os.Setenv("env", "dev")
    
    // Retry a potentially failing operation
    utils.RetryWork(func() error {
        utils.LogMessage("Attempting operation...")
        // Simulate failure
        if time.Now().Nanosecond()%2 == 0 {
            panic("simulated panic")
        }
        return nil
    }, 3)
}
```

## Usage Examples

### Example: Message Queue

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/leoxiang66/go-patterns/container/msgQueue"
)

func main() {
    // Create a message queue with capacity 10
    mq := msgqueue.NewChanMQ(10, "test-device")
    defer mq.Destroy()

    // Start a goroutine to consume messages
    go func() {
        ctx := context.Background()
        for i := 0; i < 5; i++ {
            msg, err := mq.Deq(ctx)
            if err != nil {
                fmt.Printf("Error dequeuing: %v\n", err)
                return
            }
            fmt.Printf("Received message: %s\n", string(msg))
            time.Sleep(500 * time.Millisecond)
        }
    }()

    // Enqueue messages
    for i := 0; i < 5; i++ {
        msg := fmt.Sprintf("Message %d", i)
        if err := mq.Enq([]byte(msg)); err != nil {
            fmt.Printf("Error enqueuing: %v\n", err)
            return
        }
        fmt.Printf("Sent message: %s\n", msg)
    }

    // Wait for all messages to be processed
    time.Sleep(3 * time.Second)
}
```

### Example: Priority Queue

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

## Contribution

Contributions are welcome! Feel free to suggest improvements or submit pull requests. If you have implementations of new concurrency patterns or data structures, we'd love to see them.

## License

This project is open-sourced under the MIT License. For more details, please refer to the [LICENSE](./LICENSE) file.

