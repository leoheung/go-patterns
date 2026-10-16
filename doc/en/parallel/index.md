# Parallel

The `parallel` package provides concurrency patterns and primitives for Go, helping developers write efficient concurrent programs.

## Modules

### [Barrier](./barrier.md)
Synchronization primitive that allows multiple goroutines to wait for each other.

Features:
- Cyclic barrier implementation
- Support for condition variables
- Thread-safe operations

### [Limiter](./limiter.md)
Rate limiter for controlling the rate of operations.

Features:
- Static rate limiting
- Token bucket algorithm
- Configurable intervals

### [Mutex](./mutex.md)
Simple mutual exclusion lock implementation.

Features:
- Basic Lock/Unlock operations
- Channel-based implementation
- Simple and efficient

### [Pipeline](./pipeline.md)
Pipeline patterns for data processing.

Features:
- Fan-in and Fan-out patterns
- Broadcast pattern
- Take operation
- Composable pipeline stages

### [Pools](./pool.md)
Worker pool, context-aware async task pool and generic object pool.

Features:
- WorkerPool - fixed-size worker pool based on a semaphore
- AsyncPoolV2 - context-aware async pool with bounded task queue and stats
- ObjPool - generic object pool to avoid frequent allocation/GC
- [AsyncPool](./asyncpool.md) - Async task pool with buffer support

### [PubSub](./pubsub.md)
Publish-Subscribe pattern implementation.

Features:
- Multiple subscribers
- Topic-based messaging
- Async message delivery

### [Read-Write Lock](./rwlock.md)
Read-write lock supporting multiple readers or a single writer.

Features:
- Multiple concurrent readers
- Exclusive writer access
- Starvation-free implementation

### [Semaphore](./semaphore.md)
Semaphore for limiting concurrent access to resources.

Features:
- Configurable permits
- Acquire/Release operations
- Channel and condition variable implementations

### [Token](./token.md)
Token-based synchronization primitives.

Features:
- BoolToken - Thread-safe boolean value
- StaticTokens - Fixed token count for rate limiting

### [Stream](./stream.md)
Composable streaming pipeline for data processing.

Features:
- Chainable stages (Transform, Filter, FlatMap, Parallel, Buffer, ...)
- Fan-in / Fan-out / Broadcast
- Terminal operations (Collect, ForEach, Take, Skip)
- Done channel based lifecycle

### [Communication](./communication.md)
Interfaces for message passing and time abstractions.

Features:
- Message / MessageHeader structures
- MQ interfaces (Publisher, Subscriber, Broadcaster, HttpClient)
- Clock interface
