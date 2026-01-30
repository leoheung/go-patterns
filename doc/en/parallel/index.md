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

### [Worker Pool](./pool.md)
Worker pool pattern for managing concurrent tasks.

Features:
- Fixed-size worker pool
- Task queue
- Graceful shutdown

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
