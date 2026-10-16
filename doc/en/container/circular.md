# CircularStack

A fixed-capacity, last-in-first-out (LIFO/FILO) circular buffer with change-notification subscriptions.

Designed for producer/consumer scenarios where a producer pushes data at high frequency while a consumer reads the latest item more slowly. With a plain stack, a fast producer and a slow consumer cause unbounded growth. `CircularStack` bounds memory by fixing its size: once full, every new `Push` overwrites the oldest element, so the buffer always retains at most `size` of the most recently pushed items.

Consumers subscribe to change notifications and retrieve the newest value via `Peek` or `Pop`. The stack can be paused and resumed: pausing stops notifications and terminates the background notification goroutine, while resuming restarts it (powered by `RenewableContext`).

## Installation

```go
import "github.com/leoheung/go-patterns/container/circular"
```

## API Reference

### Create

```go
// Create a fixed-size circular stack. If size <= 0, it is clamped to 1.
func NewCircularStack[T any](size int) *CircularStack[T]
```

### Push / Pop / Peek

```go
// Push appends an element at the top.
// Once the stack is full (count == size), the oldest element is overwritten
// and the top pointer advances.
// Returns an error if the stack is paused.
func (cs *CircularStack[T]) Push(data T) error

// Peek returns the newest element without removing it.
// Returns an error if the stack is empty. (Works even while paused.)
func (cs *CircularStack[T]) Peek() (T, error)

// Pop removes and returns the newest element.
// Returns an error if the stack is paused or empty.
func (cs *CircularStack[T]) Pop() (T, error)
```

### Pause / Resume

```go
// Pause stops notifications and terminates the background notification
// goroutine. Push / Pop / Subscribe return errors while paused.
func (cs *CircularStack[T]) Pause()

// Resume revives the stack and restarts the background notification
// goroutine. Returns an error if the stack is already alive.
func (cs *CircularStack[T]) Resume() error
```

### Subscribe

```go
// Subscribe registers a change-notification channel.
// Every successful Push triggers a best-effort (non-blocking) notification
// on each subscriber channel.
//
//   - buffer: subscriber channel buffer size (clamped to >= 0)
//   - returns the notification channel, an unsubscribe function, and an error
//     (error if the stack is paused)
//
// Calling the unsubscribe function removes the subscription and closes the
// channel (safe to call multiple times — only the first call closes it).
func (cs *CircularStack[T]) Subscribe(buffer int) (<-chan struct{}, func(), error)
```

> Note: notifications are coalesced — the internal notify channel has a buffer of 1, so a burst of `Push` calls may result in a single notification to subscribers. Combine `Subscribe` with `Peek` to always read the newest value.

## Complete Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/container/circular"
)

func main() {
	// Create a circular stack that keeps the 3 most recent items
	cs := circular.NewCircularStack[int](3)

	// Subscribe to change notifications (buffered channel of size 4)
	ch, unsubscribe, err := cs.Subscribe(4)
	if err != nil {
		panic(err)
	}
	defer unsubscribe()

	// Background consumer: wake up on every change and read the newest item
	go func() {
		for range ch {
			if top, err := cs.Peek(); err == nil {
				fmt.Println("latest:", top)
			}
		}
	}()

	// Producer: push 5 items into a size-3 stack — 1 and 2 get overwritten
	for i := 1; i <= 5; i++ {
		if err := cs.Push(i); err != nil {
			panic(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Pop the newest element
	top, err := cs.Pop()
	if err != nil {
		panic(err)
	}
	fmt.Println("popped:", top) // 5

	// Pause the stack: notifications stop, Push/Pop return errors
	cs.Pause()
	if err := cs.Push(6); err != nil {
		fmt.Println("push while paused:", err) // CircularStack is paused
	}

	// Resume restarts the notification goroutine
	if err := cs.Resume(); err != nil {
		panic(err)
	}
}
```

## Features

- **Fixed Capacity, Bounded Memory**: once full, `Push` overwrites the oldest element — ideal for high-frequency producers with slow consumers.
- **LIFO Semantics**: `Peek` / `Pop` always operate on the most recently pushed item.
- **Subscription Notifications**: `Subscribe` returns a channel plus an unsubscribe function; every `Push` notifies subscribers (best-effort, non-blocking).
- **Pause / Resume**: built on `RenewableContext` — pausing stops notifications and the background goroutine; resuming revives them.
- **Concurrency Safe**: all operations are guarded by a read-write lock.
- **Generics**: works with any element type.
