# Container

The `container` package provides generic data structures for Go, designed to be efficient and easy to use.

## Modules

### [List](./list.md)
A generic dynamic array implementation supporting Python list and JavaScript Array operations.

Features:
- Generic type support
- Dynamic resizing
- Support for negative indices
- Rich API (Append, Push, Pop, Shift, Unshift, etc.)
- Functional operations (Map, Filter, Reduce)

### [Message Queue](./msgqueue.md)
A channel-based message queue implementation with basic queue operations.

Features:
- Channel-based implementation
- Context support for cancellation
- Queue lifecycle management
- Thread-safe operations

### [Priority Queue](./pq.md)
A generic priority queue implementation with customizable priority comparison.

Features:
- Generic type support
- Custom comparison function
- Binary heap implementation
- Efficient enqueue/dequeue operations

### [Cache](./cache.md)
A cache implementation for storing and retrieving data.

Features:
- Key-value storage
- TTL support
- Thread-safe operations
