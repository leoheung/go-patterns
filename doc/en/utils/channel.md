# Channel Utilities

## Overview

The `utils` package provides non-blocking and timeout-based channel operations for safer concurrent programming.

## Features

- **Non-blocking Enqueue/Dequeue**: Send and receive without blocking
- **Timeout-based Operations**: Add time limits to channel operations
- **Generic Support**: Works with any data type

## API Reference

### `TryEnqueue[T any](c chan<- T, data T) bool`

Attempts to send data to a channel without blocking. Returns `true` if successful, `false` if the channel is full or blocked.

```go
ch := make(chan int, 5)

ok := utils.TryEnqueue(ch, 42)
if ok {
    fmt.Println("Enqueued successfully")
}
```

### `TryDequeue[T any](c <-chan T) (*T, bool)`

Attempts to receive data from a channel without blocking. Returns the data and `true` if successful, or `nil` and `false` if the channel is empty.

```go
ch := make(chan int, 5)
ch <- 10

val, ok := utils.TryDequeue(ch)
if ok {
    fmt.Printf("Dequeued: %d\n", *val)
}
```

### `EnqueueWithTimeout[T any](c chan<- T, data T, timeout time.Duration) bool`

Attempts to send data to a channel with a timeout. Returns `true` if successful, `false` if the timeout expires.

```go
ch := make(chan int, 2)

ok := utils.EnqueueWithTimeout(ch, 42, 5*time.Second)
if ok {
    fmt.Println("Enqueued within timeout")
} else {
    fmt.Println("Enqueue timed out")
}
```

### `DequeueWithTimeout[T any](c <-chan T, timeout time.Duration) (*T, bool)`

Attempts to receive data from a channel with a timeout. Returns the data and `true` if successful, or `nil` and `false` if the timeout expires.

```go
ch := make(chan int, 5)

val, ok := utils.DequeueWithTimeout(ch, 5*time.Second)
if ok {
    fmt.Printf("Dequeued within timeout: %d\n", *val)
} else {
    fmt.Println("Dequeue timed out")
}
```

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	ch := make(chan string, 3)

	// Non-blocking enqueue
	for _, msg := range []string{"a", "b", "c"} {
		if ok := utils.TryEnqueue(ch, msg); ok {
			fmt.Printf("Enqueued: %s\n", msg)
		} else {
			fmt.Printf("Failed to enqueue: %s\n", msg)
		}
	}

	// Non-blocking dequeue
	for i := 0; i < 4; i++ {
		if val, ok := utils.TryDequeue(ch); ok {
			fmt.Printf("Dequeued: %s\n", *val)
		} else {
			fmt.Println("Channel is empty")
		}
	}

	// Timeout-based operations
	largeCh := make(chan int, 1)
	largeCh <- 1

	// This will timeout because channel is full
	ok := utils.EnqueueWithTimeout(largeCh, 2, 100*time.Millisecond)
	fmt.Printf("Enqueue with timeout: %v\n", ok)
}
```

## Notes

- `TryEnqueue` and `TryDequeue` are non-blocking and return immediately
- Timeout functions use `time.Timer` for efficient timing
- All functions are thread-safe
- Generic type parameter `[T any]` works with any data type