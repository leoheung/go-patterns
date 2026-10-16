# ShardedMap

A sharded concurrent map (generics) with per-shard locking.

`ShardedMap[K, V]` splits its keyspace into a fixed number of shards, each with its own `sync.RWMutex`. Concurrent operations on different keys usually hit different shards and never block each other — much lower lock contention than a single mutex around a plain `map[K]V`, especially under heavy parallel access.

Shard selection uses a type-optimised hash: `string` keys use FNV-1a 64-bit, integer keys hash to their value directly, other types fall back to a memory-based hash.

## Installation

```go
import "github.com/leoheung/go-patterns/container/safemap"
```

## API Reference

### Create

```go
// Create a sharded map with shardCount shards.
// If shardCount <= 0, it defaults to 32.
func NewShardedMap[K comparable, V any](shardCount int) *ShardedMap[K, V]
```

### Basic Operations

```go
// Get returns the value for a key and whether it exists (read lock).
func (sm *ShardedMap[K, V]) Get(key K) (V, bool)

// Set stores a value under a key (write lock).
func (sm *ShardedMap[K, V]) Set(key K, value V)

// Delete removes a key (write lock).
func (sm *ShardedMap[K, V]) Delete(key K)

// GetOrStore returns the existing value if present; otherwise stores value
// and returns it. loaded reports whether the value was already present
// (same semantics as sync.Map).
func (sm *ShardedMap[K, V]) GetOrStore(key K, value V) (actual V, loaded bool)
```

### Compute

```go
// Compute atomically updates a key's value under the shard's write lock:
//
//   - if the key exists and computePresent != nil:
//     computePresent(oldVal) is stored;
//   - otherwise (key absent, or computePresent == nil):
//     computeAbsent() is called and its result is stored (when the key exists
//     but computePresent is nil, this overwrites the existing value);
//   - returns an error ("invalid parameters") if computeAbsent is nil in the
//     second case.
func (sm *ShardedMap[K, V]) Compute(key K, computeAbsent func() V, computePresent func(V) V) error
```

## Complete Example

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/container/safemap"
)

func main() {
	// Create a map with 16 shards
	sm := safemap.NewShardedMap[string, int](16)

	// Set / Get
	sm.Set("alice", 1)
	sm.Set("bob", 2)

	if v, ok := sm.Get("alice"); ok {
		fmt.Println("alice:", v) // 1
	}

	// GetOrStore: keep the existing value if present
	actual, loaded := sm.GetOrStore("alice", 100)
	fmt.Println(actual, loaded) // 1 true
	actual, loaded = sm.GetOrStore("carol", 3)
	fmt.Println(actual, loaded) // 3 false

	// Compute: atomically increment a counter
	_ = sm.Compute("counter", func() int {
		return 1 // absent: start at 1
	}, func(old int) int {
		return old + 1 // present: increment
	})
	v, _ := sm.Get("counter")
	fmt.Println("counter:", v) // 1

	_ = sm.Compute("counter", func() int { return 1 }, func(old int) int { return old + 1 })
	v, _ = sm.Get("counter")
	fmt.Println("counter:", v) // 2

	// Delete
	sm.Delete("bob")
	_, ok := sm.Get("bob")
	fmt.Println("bob exists:", ok) // false
}
```

## Features

- **Sharded Locking**: fixed number of shards, each with its own `RWMutex` — low lock contention under parallel access.
- **Generics**: `K comparable, V any`, works with any key/value types.
- **Type-Optimised Hashing**: FNV-1a for `string`, identity for integers, memory-based fallback for other types.
- **sync.Map-Like API**: `Get` / `Set` / `Delete` / `GetOrStore` mirror familiar semantics.
- **Atomic Compute**: read-modify-write under a single write lock, without manual locking.
