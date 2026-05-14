# Token

## Overview

The token package provides two types of token-based synchronization primitives for managing concurrent access:

1. **BoolToken** - A thread-safe boolean value wrapper
2. **StaticTokens** - A fixed number of tokens for rate limiting

## BoolToken

A simple thread-safe boolean value that can be get and set atomically.

### API Reference

#### `NewBoolToken(value bool) *BoolToken`

Creates a new BoolToken with the given initial value.

#### `(bt *BoolToken) Get() bool`

Returns the current boolean value in a thread-safe manner.

#### `(bt *BoolToken) Set(value bool)`

Sets the boolean value in a thread-safe manner.

### Example

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/parallel/token"
)

func main() {
	bt := token.NewBoolToken(true)

	// Read the value
	fmt.Println("Initial value:", bt.Get()) // true

	// Update the value
	bt.Set(false)
	fmt.Println("After set:", bt.Get()) // false
}
```

## StaticTokens

A fixed-size token bucket for controlling concurrent access or rate limiting.

### API Reference

#### `NewStaticTokens(numTokens int) (*StaticTokens, error)`

Creates a new StaticTokens with the specified number of tokens. Returns an error if `numTokens <= 0`.

#### `(st *StaticTokens) GrantNextToken() bool`

Attempts to grant the next token. Returns `true` if a token was successfully granted, `false` if no tokens are available. This operation is atomic and thread-safe.

### Example

```go
package main

import (
	"fmt"
	"sync"
	"github.com/leoheung/go-patterns/parallel/token"
)

func main() {
	st, err := token.NewStaticTokens(3)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if st.GrantNextToken() {
				fmt.Printf("Worker %d: acquired token\n", id)
				// Do work...
				// Token is automatically returned when done
			} else {
				fmt.Printf("Worker %d: no token available\n", id)
			}
		}(i)
	}
	wg.Wait()
}
```

## Notes

- `BoolToken` is useful for simple flag-based synchronization
- `StaticTokens` is useful for implementing semaphore-like behavior or simple rate limiting
- Both types are safe for concurrent use without additional synchronization