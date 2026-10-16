# RenewableContext

A revivable `context.Context` with a generic data payload.

Standard Go contexts are one-shot: once cancelled or timed out, the `Done()` channel is closed forever and the context can never be reused. `RenewableContext[T]` solves this by allowing a dead context to be **reactivated** — a brand-new internal context is created and every subscriber is notified — while also carrying a type-safe generic data payload (`T`) protected by a read-write lock.

It fully implements `context.Context`, so it can be passed to any function expecting a standard context.

## Installation

```go
import "github.com/leoheung/go-patterns/container/context"
```

> Note: the package name is `context`, so it is usually imported with an alias to avoid clashing with the standard library:
>
> ```go
> import lctx "github.com/leoheung/go-patterns/container/context"
> ```

## Concepts

- **Revivable**: `Cancel()` (or a timeout) kills the context, but `Reactivate()` brings it back to life with a fresh internal context. `Done()` then returns the new channel.
- **Generic data payload**: `GetData()` / `SetData()` store a value of type `T` alongside the context, guarded by a `sync.RWMutex` — safe for concurrent use.
- **Reactivation subscription**: `SubscribeReactivation()` returns a channel that receives a notification every time the context is reactivated. Background workers can wait on it to restart after a pause.
- **Merge**: `MergeContext(otherCtx)` binds the lifetime of another context into this one — whichever dies first kills the merged context.

## API Reference

### Create

```go
// Create a RenewableContext carrying data.
// If timeout is nil, the context has no deadline; otherwise it times out after *timeout.
func NewRenewableContext[T any](timeout *time.Duration, data T) *RenewableContext[T]
```

### Lifecycle

```go
// Cancel the context (kills it; can later be revived with Reactivate).
func (gt *RenewableContext[T]) Cancel()

// Done returns the done channel of the current internal context.
// After Reactivate, a new channel is returned.
func (gt *RenewableContext[T]) Done() <-chan struct{}

// IsAlive reports whether the current internal context is still alive.
func (gt *RenewableContext[T]) IsAlive() bool

// Reactivate revives a dead context with a fresh internal context.
// All reactivation subscribers are notified.
// Returns an error if the current context is still alive.
func (gt *RenewableContext[T]) Reactivate(timeout *time.Duration) error

// SubscribeReactivation returns a channel that receives a notification
// whenever Reactivate succeeds. The channel is never closed.
//
// Note: the notification is a best-effort non-blocking send — make sure a
// receiver is already waiting on the channel before calling Reactivate,
// otherwise the notification is dropped.
func (gt *RenewableContext[T]) SubscribeReactivation() <-chan struct{}
```

### Generic Data Payload

```go
// GetData returns the current payload (concurrency-safe).
func (gt *RenewableContext[T]) GetData() T

// SetData replaces the payload (concurrency-safe).
func (gt *RenewableContext[T]) SetData(newData T)
```

### Context Merging

```go
// MergeContext merges otherCtx into this context: the resulting context is
// cancelled as soon as either the current context or otherCtx is cancelled.
func (gt *RenewableContext[T]) MergeContext(otherCtx context.Context)
```

### context.Context Interface

`RenewableContext[T]` implements `context.Context` (verified at compile time in the package: `var _ context.Context = new(RenewableContext[struct{}])`):

```go
func (gt *RenewableContext[T]) Deadline() (deadline time.Time, ok bool)
func (gt *RenewableContext[T]) Done() <-chan struct{}
func (gt *RenewableContext[T]) Err() error
func (gt *RenewableContext[T]) Value(key any) any
```

All four methods delegate to the current internal context under a read lock.

## Complete Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	lctx "github.com/leoheung/go-patterns/container/context"
)

func main() {
	// Create a context with a 50ms timeout and a string payload
	timeout := 50 * time.Millisecond
	ctx := lctx.NewRenewableContext(&timeout, "hello")

	// It implements context.Context — use it anywhere a std context is expected
	var _ context.Context = ctx

	// Subscribe to reactivation notifications (e.g. a background worker).
	// The notification is a non-blocking send, so the receiver must already
	// be waiting when Reactivate is called.
	resumeCh := ctx.SubscribeReactivation()
	notified := make(chan struct{})
	go func() {
		<-resumeCh
		close(notified)
	}()

	// Wait for the timeout to kill the context
	time.Sleep(100 * time.Millisecond)
	fmt.Println("alive after timeout:", ctx.IsAlive()) // false

	// Revive the context
	if err := ctx.Reactivate(nil); err != nil {
		panic(err)
	}
	fmt.Println("alive after reactivate:", ctx.IsAlive()) // true

	// The subscriber is notified
	<-notified
	fmt.Println("reactivation notified")

	// Generic data payload
	fmt.Println("data:", ctx.GetData()) // hello
	ctx.SetData("world")
	fmt.Println("data:", ctx.GetData()) // world

	// Cancel and revive again
	ctx.Cancel()
	_ = ctx.Reactivate(nil)

	// Merge with another context: the merged context dies when either side dies
	other, cancelOther := context.WithCancel(context.Background())
	ctx.MergeContext(other)
	cancelOther()
	time.Sleep(10 * time.Millisecond)
	fmt.Println("alive after merge-cancel:", ctx.IsAlive()) // false
}
```

## Features

- **Revivable Context**: unlike `context.Context`, a cancelled/timed-out context can be brought back to life with `Reactivate`.
- **Implements context.Context**: drop-in replacement in any API accepting a standard context.
- **Generic Data Payload**: carry type-safe data (`T`) next to the context, with built-in `RWMutex` protection.
- **Reactivation Subscription**: background goroutines can sleep on `SubscribeReactivation()` and restart automatically when the context revives.
- **Context Merging**: bind the lifetime of an external context with `MergeContext` — the merged context is cancelled when either side ends.
- **Concurrency Safe**: all methods are guarded by a read-write lock.
