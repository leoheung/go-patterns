# ValueProvider

An observable value with subscriber notification (generic).

`ValueProvider[T]` wraps a single value of type `T` and notifies all subscribers whenever the value changes. Every successful `Set` broadcasts the new value to each subscriber's channel. It supports pausing/resuming (built on `RenewableContext`) and safe unsubscription.

## Installation

```go
import "github.com/leoheung/go-patterns/container/subscribe"
```

## API Reference

### Create

```go
// Create a ValueProvider with an initial value.
// buffer_size is the buffer of the internal broadcast channel (clamped to >= 0).
func NewValueProvider[T any](val T, buffer_size int) *ValueProvider[T]
```

### Get / Set

```go
// Get returns the current value (concurrency-safe).
func (v *ValueProvider[T]) Get() T

// Set stores the new value and broadcasts it to all subscribers.
// Returns an error if the provider is paused.
func (v *ValueProvider[T]) Set(val T) error
```

### Subscribe

```go
// Subscribe registers a subscriber channel.
//
//   - buffer: subscriber channel buffer size (clamped to >= 0)
//   - returns the value channel, an unsubscribe function, and an error
//     (error if the provider is paused)
//
// Each successful Set delivers the new value to every subscriber channel
// via a best-effort (non-blocking) send — a full subscriber channel drops
// the value, so size the buffer according to your consumer's speed.
//
// Calling the unsubscribe function removes the subscription and closes the
// channel (safe to call multiple times — only the first call closes it).
func (v *ValueProvider[T]) Subscribe(buffer int) (<-chan T, func(), error)
```

### Pause / Resume

```go
// Pause stops the provider: Set and Subscribe return errors while paused.
func (v *ValueProvider[T]) Pause()

// Resume revives the provider. Returns an error if it is already alive.
func (v *ValueProvider[T]) Resume() error
```

## Complete Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/container/subscribe"
)

func main() {
	// Create a provider holding an int, with an internal broadcast buffer of 8
	vp := subscribe.NewValueProvider(0, 8)

	// Subscribe with a buffered channel
	ch, unsubscribe, err := vp.Subscribe(4)
	if err != nil {
		panic(err)
	}
	defer unsubscribe()

	// Background observer: print every change
	go func() {
		for val := range ch {
			fmt.Println("changed:", val)
		}
	}()

	// Update the value — subscribers are notified
	for i := 1; i <= 3; i++ {
		if err := vp.Set(i * 10); err != nil {
			panic(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("current:", vp.Get()) // 30

	// Pause: Set and Subscribe return errors
	vp.Pause()
	if err := vp.Set(99); err != nil {
		fmt.Println("set while paused:", err) // ValueProvider is paused
	}

	// Resume revives the provider
	if err := vp.Resume(); err != nil {
		panic(err)
	}
	_ = vp.Set(40) // notified again
	time.Sleep(10 * time.Millisecond)
}
```

## Features

- **Observable Value**: every `Set` broadcasts the new value to all subscribers.
- **Type-Safe Generics**: wraps any value type `T`.
- **Safe Unsubscription**: `Subscribe` returns an unsubscribe closure; it removes the subscription and closes the channel exactly once.
- **Pause / Resume**: built on `RenewableContext` — pausing rejects `Set`/`Subscribe`; resuming revives the provider.
- **Concurrency Safe**: `Get`/`Set` are guarded by a read-write lock.
- **Decoupled Broadcast**: values flow through an internal buffered channel (`buffer_size`), then a background worker fans them out to subscribers with best-effort sends — slow subscribers drop values instead of blocking each other. Note `Set` writes to the internal channel, so it can block briefly if that buffer is full; size `buffer_size` for your producer's burst rate.
