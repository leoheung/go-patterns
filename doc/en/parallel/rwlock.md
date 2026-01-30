# Read-Write Lock

A read-write lock implementation supporting multiple readers or a single writer.

## Installation

```go
import "github.com/leoheung/go-patterns/parallel/rwlock"
```

## API Reference

### Create a Read-Write Lock

```go
// Create a new read-write lock
rw := rwlock.NewRWLock()
```

### Read Lock

```go
// Acquire read lock
rw.RLock()
defer rw.RUnlock()
```

### Write Lock

```go
// Acquire write lock
rw.WLock()
defer rw.WUnlock()
```

## Complete Example

```go
package main

import (
    "fmt"
    "sync"
    "time"
    "github.com/leoheung/go-patterns/parallel/rwlock"
)

func main() {
    rw := rwlock.NewRWLock()
    data := make(map[string]string)
    var wg sync.WaitGroup

    // Writers
    for i := 0; i < 2; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            rw.WLock()
            defer rw.WUnlock()

            key := fmt.Sprintf("key-%d", id)
            data[key] = fmt.Sprintf("value-%d", id)
            fmt.Printf("Writer %d wrote %s\n", id, key)
            time.Sleep(100 * time.Millisecond)
        }(i)
    }

    // Readers
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()

            rw.RLock()
            defer rw.RUnlock()

            fmt.Printf("Reader %d reading, data count: %d\n", id, len(data))
            time.Sleep(50 * time.Millisecond)
        }(i)
    }

    wg.Wait()
}
```

## Output

```
Writer 0 wrote key-0
Reader 0 reading, data count: 1
Reader 1 reading, data count: 1
Writer 1 wrote key-1
Reader 2 reading, data count: 2
...
```

## Features

- **Multiple readers**: Concurrent read access
- **Exclusive writer**: Only one writer at a time
- **No reader starvation**: Fair scheduling
