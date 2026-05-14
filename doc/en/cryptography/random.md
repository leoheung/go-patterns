# Random Number Generation

## Overview

The `cryptography` package provides cryptographically secure random number generation utilities, including random strings and UUID generation.

## RandString

Generates a random string of specified length using cryptographically secure random numbers.

### API Reference

#### `RandString(n int) string`

Returns a random string of length `n` using characters `a-zA-Z0-9`. Returns an empty string if `n <= 0`.

```go
randomStr := cryptography.RandString(16)
fmt.Println(randomStr) // e.g., "aB3xY9kLmN2pQrT4"
```

### Example

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/cryptography"
)

func main() {
	// Generate random strings of different lengths
	for i := 4; i <= 16; i += 4 {
		fmt.Printf("Length %d: %s\n", i, cryptography.RandString(i))
	}

	// Generate random password
	password := cryptography.RandString(32)
	fmt.Printf("Random password: %s\n", password)
}
```

## RandUUID

Generates a RFC 4122 Version 4 UUID using cryptographically secure random numbers.

### API Reference

#### `RandUUID() string`

Returns a UUID in the format `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`. Returns an empty string if random number generation fails.

```go
uuid := cryptography.RandUUID()
fmt.Println(uuid) // e.g., "550e8400-e29b-41d4-a716-446655440000"
```

### Example

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/cryptography"
)

func main() {
	// Generate multiple UUIDs
	for i := 0; i < 5; i++ {
		uuid := cryptography.RandUUID()
		fmt.Printf("UUID %d: %s\n", i+1, uuid)
	}
}
```

## Notes

- Both functions use `crypto/rand` for cryptographically secure random number generation
- `RandString` uses lowercase letters, uppercase letters, and digits (62 characters total)
- `RandUUID` generates RFC 4122 compliant Version 4 UUIDs
- Both functions are safe for use in security-sensitive applications like session IDs, tokens, and encryption keys