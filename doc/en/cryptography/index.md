# Cryptography

The `cryptography` package provides a set of tools combining modern encryption standards, including dynamic AES key management and secure random number generation.

## Modules

### 1. Dynamic AES (DynamicAES)

`DynamicAES` is an advanced encryption tool that combines **ECDH (X25519)** key agreement with **AES-GCM** authenticated encryption. It is specifically designed for scenarios where a session key needs to be securely transferred between a client and a server.

#### Initialization

```go
import "github.com/leoheung/go-patterns/cryptography"

// Create an encryptor with a specified key length (e.g., 32 bytes for AES-256)
// If length > 0, a random initial key is automatically generated.
aesTool := cryptography.NewDynamicAES(32)
```

#### Secure Key Transfer (Handshake)

```go
// --- Client (App) ---
// Encrypt the current AES key using the server's Base64 public key
// Returns an encrypted package (Base64) containing a temporary public key and the encrypted AES key
encryptedPackage, err := client.GetKey(serverPublicKeyStr)

// --- Server (Server) ---
// Use its own private key to decrypt the package and set the internal AES key as the current session key
err := server.SetKey(serverPrivateKeyStr, encryptedPackage)
```

#### Encryption and Decryption

```go
// Encrypt any object (internally serialized to JSON)
// Returns ciphertext in Base64 format (includes Nonce)
cipherText, err := aesTool.Encrypt(myData)

// Decrypt Base64 ciphertext into a target object
err := aesTool.Decrypt(cipherText, &myResult)
```

### 2. Random Generation

Provides cryptographically secure random value generation functions.

```go
// Generate a random string of specified length (Charset: a-zA-Z0-9)
str := cryptography.RandString(16)

// Generate a Version 4 UUID compliant with RFC 4122
uuid := cryptography.RandUUID() // Format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

## Complete Example (Simulated Handshake and Communication)

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/cryptography"
)

func main() {
    // 1. Initialization
    client := cryptography.NewDynamicAES(32)
    server := cryptography.NewDynamicAES(0) // Server starts without a key

    // 2. Simulated Key Transfer (Assuming ECC keys exist)
    // encryptedPkg, _ := client.GetKey(pubKey)
    // server.SetKey(privKey, encryptedPkg)

    // 3. Data Encryption Communication
    type Message struct {
        Text string
        ID   int
    }
    
    msg := Message{"Secret Message", 101}
    cipherText, _ := client.Encrypt(msg)
    
    // Server Decryption
    var received Message
    server.Decrypt(cipherText, &received)
    
    fmt.Printf("Decrypted result: %+v\n", received)
}
```

## Features

- **Security**: 
    - Uses **X25519** for ECDH key agreement to ensure secure transmission.
    - Uses **AES-GCM** (with 256-bit hash-derived keys) for authenticated encryption, preventing data tampering.
- **Flexibility**: Supports direct encryption of any `any` type in Go (Struct, Map, Slice, etc.).
- **Convenience**: All inputs and outputs are Base64 encoded for easy transmission over HTTP or JSON.
- **Compliance**: Random number generation is based on `crypto/rand`, meeting cryptographic security requirements.
