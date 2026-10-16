# SafeSlice

A concurrency-safe generic slice.

`SafeSlice[T]` wraps a plain `[]T` with a `sync.RWMutex`, providing safe concurrent access for producer/consumer and shared-state scenarios. Read operations use a read lock (non-blocking for other readers); mutations use a write lock.

## Installation

```go
import "github.com/leoheung/go-patterns/container/safeslice"
```

## API Reference

### Create

```go
// Create a SafeSlice.
//   - capacity: slice capacity (clamped to >= 0, and >= length)
//   - length: initial length (clamped to >= 0)
func NewSafeSlice[T any](capacity, length int) *SafeSlice[T]
```

### Element Access

```go
// Peek returns the element at index without removing it.
// Out-of-range indices return (zero, false).
func (l *SafeSlice[T]) Peek(index int) (T, bool)

// PeekFirst returns the first element (index 0).
func (l *SafeSlice[T]) PeekFirst() (T, bool)

// PeekLast returns the last element.
func (l *SafeSlice[T]) PeekLast() (T, bool)
```

### Iteration

```go
// Range iterates all elements under a read lock (does not block other readers).
// Stop early by returning false from f.
func (l *SafeSlice[T]) Range(f func(index int, item T) bool)
```

### Mutation

```go
// Append appends an element to the end.
func (l *SafeSlice[T]) Append(item T)

// RemoveIf removes all elements matching the predicate and returns the
// number removed.
func (l *SafeSlice[T]) RemoveIf(predicate func(T) bool) int
```

### Query

```go
// Len returns the current length.
func (l *SafeSlice[T]) Len() int
```

## Complete Example

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/container/safeslice"
)

func main() {
	// Create an empty SafeSlice of strings
	l := safeslice.NewSafeSlice[string](0, 0)

	// Append elements
	l.Append("apple")
	l.Append("banana")
	l.Append("cherry")
	fmt.Println("len:", l.Len()) // 3

	// Peek elements
	if v, ok := l.Peek(1); ok {
		fmt.Println("index 1:", v) // banana
	}
	if v, ok := l.PeekFirst(); ok {
		fmt.Println("first:", v) // apple
	}
	if v, ok := l.PeekLast(); ok {
		fmt.Println("last:", v) // cherry
	}

	// Range with early stop
	l.Range(func(index int, item string) bool {
		fmt.Println(index, item)
		return index < 1 // stop after the second element
	})

	// RemoveIf: remove elements longer than 5 characters
	removed := l.RemoveIf(func(s string) bool { return len(s) > 5 })
	fmt.Println("removed:", removed) // 2 (banana, cherry)
	fmt.Println("len:", l.Len())     // 1
}
```

## Features

- **Concurrency Safe**: every operation is guarded by a `sync.RWMutex`.
- **Read-Lock Iteration**: `Range` holds only a read lock — concurrent readers do not block each other.
- **Subscriber-List Friendly**: `RemoveIf` makes it easy to maintain subscriber lists (used internally by `circular` and `subscribe`).
- **Generics**: works with any element type `T`.
- **Zero-Value Safe Access**: out-of-range `Peek` returns `(zero, false)` instead of panicking.
