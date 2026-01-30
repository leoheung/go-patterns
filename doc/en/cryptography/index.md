# Cryptography

The `cryptography` package provides cryptographic utilities.

## Modules

### Dynamic AES

Dynamic AES encryption with key rotation support.

```go
import "github.com/leoheung/go-patterns/cryptography"

// Create AES encrypter with key
encrypter := cryptography.NewDynamicAES([]byte("your-32-byte-key-here!!!!!!!!!!!!"))

// Encrypt data
encrypted, err := encrypter.Encrypt([]byte("sensitive data"))
if err != nil {
    // Handle error
}

// Decrypt data
decrypted, err := encrypter.Decrypt(encrypted)
if err != nil {
    // Handle error
}
```

### Random

Generate cryptographically secure random values.

```go
// Generate random bytes
randomBytes, err := cryptography.GenerateRandomBytes(32)

// Generate random string
randomString, err := cryptography.GenerateRandomString(16)

// Generate random int
randomInt := cryptography.GenerateRandomInt(100)
```

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/cryptography"
)

func main() {
    // Generate a random key
    key, _ := cryptography.GenerateRandomBytes(32)

    // Create encrypter
    aes := cryptography.NewDynamicAES(key)

    // Encrypt
    plaintext := []byte("Hello, World!")
    ciphertext, err := aes.Encrypt(plaintext)
    if err != nil {
        panic(err)
    }

    // Decrypt
    decrypted, err := aes.Decrypt(ciphertext)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Original: %s\n", plaintext)
    fmt.Printf("Decrypted: %s\n", decrypted)
}
```

## Features

- **AES encryption**: Industry-standard encryption
- **Key rotation**: Support for dynamic key changes
- **Secure random**: Cryptographically secure random generation
- **Simple API**: Easy to use encrypt/decrypt interface
