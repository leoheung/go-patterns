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
A lightweight topic-based message queue (`MQLite`) that delivers messages to subscribers over HTTP callbacks. The implementation lives in `parallel/communication/mq`.

Features:
- Topic-based pub/sub over HTTP callback URLs
- Non-blocking publish with automatic retry
- Root context lifecycle management

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

### [Renewable Context](./context.md)
A revivable `context.Context` implementation with a generic data carrier.

Features:
- Cancel and reactivate (revive) the context
- Reactivation subscriptions
- Generic get/set data attached to the context

### [Circular Stack](./circular.md)
Fixed-capacity circular stack with subscription notifications.

Features:
- Fixed capacity, overwrites oldest entries
- Subscription notifications on push
- Generic type support

### [Subscribe](./subscribe.md)
Observable value provider with subscriber notification.

Features:
- Observable value with change notification
- Subscribe/unsubscribe watchers
- Generic type support

### [Safe Map](./safemap.md)
Sharded concurrent map.

Features:
- Sharded locking for high concurrency
- Get/Set/Compute operations
- Generic key-value support

### [Safe Slice](./safeslice.md)
Concurrency-safe slice.

Features:
- Thread-safe append/read/remove
- Conditional removal (RemoveIf)
- Generic type support

### [Skip List](./skiplist.md)
Probabilistic skip list for ordered key-value storage.

Features:
- O(log n) expected insert/lookup/delete
- Ordered traversal
- Concurrent-friendly design

### [Tree](./tree.md)
Tree-based structures: treap (bst), dbtree and binary heap.

Features:
- bst — treap (randomized BST)
- dbtree — distance/bearing tree
- heap — generic binary heap
