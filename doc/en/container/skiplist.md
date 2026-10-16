# SkipList

A generic skip list with optional concurrency safety.

A skip list is a probabilistic sorted data structure providing O(log n) search, insert and delete on average — a simpler alternative to balanced trees. This implementation is generic (`T any`) with a user-supplied comparison function, and offers both a plain single-goroutine mode and a read-write-locked concurrent mode.

## Installation

```go
import "github.com/leoheung/go-patterns/container/skiplist"
```

## Concepts

- **Ordering function** `better(a, b T) bool`: returns `true` when `a` comes before `b` (i.e. `a` is smaller / higher priority). All ordering behaviour (search position, `GetMin`, `RangeQuery`, …) is derived from it.
- **Levels & probability**: each node has a random number of levels. The default config is `DefaultMaxLevel = 16` with `DefaultProbability = 0.5` (each level promotes with probability 1/2).
- **Dedup on insert**: `Insert` does nothing if an equal element already exists (no duplicates).
- **Optional locking**: `withRWLock` decides whether a `sync.RWMutex` guards all operations. Pass `true` for shared concurrent use, `false` for single-goroutine use (zero lock overhead).

## API Reference

### Create

```go
const (
	DefaultMaxLevel    = 16  // maximum levels
	DefaultProbability = 0.5 // promotion probability 1/2
)

// Create a skip list with default configuration.
// better: comparison function — true means a comes before b.
// withRWLock: true enables internal read-write locking (concurrency-safe).
func New[T any](better func(a, b T) bool, withRWLock bool) *SkipList[T]

// Create a skip list with a custom maxLevel and promotion probability.
func NewWithConfig[T any](better func(a, b T) bool, maxLevel int, probability float64, withRWLock bool) *SkipList[T]
```

### Core Operations

```go
// Insert adds an element. Duplicates are ignored (dedup).
func (sl *SkipList[T]) Insert(value T)

// Search looks up an element and returns its node, or nil if not found.
// Note: node fields are unexported; the return value is mainly useful as a
// found/not-found check — prefer Contains for existence tests.
func (sl *SkipList[T]) Search(value T) *Node[T]

// Delete removes an element. Returns true if removed, false if not found.
func (sl *SkipList[T]) Delete(value T) bool
```

### Query Operations

```go
// RangeQuery returns all elements within the closed range [min, max],
// in ascending order.
func (sl *SkipList[T]) RangeQuery(min, max T) []T

// GetMin returns the smallest element (zero value if empty).
func (sl *SkipList[T]) GetMin() T

// GetMax returns the largest element (zero value if empty).
func (sl *SkipList[T]) GetMax() T

// Contains reports whether an element exists.
func (sl *SkipList[T]) Contains(value T) bool

// GetAll returns all elements in ascending order.
func (sl *SkipList[T]) GetAll() []T

// Len returns the number of elements.
func (sl *SkipList[T]) Len() int
```

### Utility

```go
// Clear removes all elements and resets the list.
func (sl *SkipList[T]) Clear()
```

## Complete Example

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/container/skiplist"
)

func main() {
	// Create a concurrency-safe skip list of ints in ascending order
	sl := skiplist.New(func(a, b int) bool { return a < b }, true)

	// Insert elements (duplicates are ignored)
	for _, v := range []int{5, 3, 9, 1, 7, 3} {
		sl.Insert(v)
	}
	fmt.Println("len:", sl.Len()) // 5

	// Search / Contains
	fmt.Println("contains 7:", sl.Contains(7))   // true
	fmt.Println("contains 4:", sl.Contains(4))   // false

	// Min / Max
	fmt.Println("min:", sl.GetMin()) // 1
	fmt.Println("max:", sl.GetMax()) // 9

	// Range query [3, 7]
	fmt.Println("range [3,7]:", sl.RangeQuery(3, 7)) // [3 5 7]

	// All elements in order
	fmt.Println("all:", sl.GetAll()) // [1 3 5 7 9]

	// Delete
	fmt.Println("delete 5:", sl.Delete(5)) // true
	fmt.Println("delete 5:", sl.Delete(5)) // false (already gone)

	// Clear
	sl.Clear()
	fmt.Println("len after clear:", sl.Len()) // 0
}
```

## Features

- **O(log n) Average Complexity**: search, insert and delete are probabilistically balanced by level promotion.
- **Generics + Custom Ordering**: any type `T` with any total order via `better`.
- **Optional Concurrency Safety**: `withRWLock` toggles internal `RWMutex` locking — pay only for what you need.
- **Dedup Insert**: equal elements are never inserted twice.
- **Range Queries**: closed-interval `RangeQuery(min, max)` in ascending order.
- **Configurable Shape**: custom `maxLevel` and promotion `probability` via `NewWithConfig`.
