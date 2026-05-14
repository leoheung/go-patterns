# Dynamic AES Encryption

## Overview

`DynamicAES` is a dynamic AES key management and encryption tool that combines ECDH key agreement with AES-GCM encryption. It provides secure key transmission for scenarios like client-server communication.

## Features

- **Dynamic AES Key Generation**: Generate random AES keys of any length
- **ECDH Key Agreement**: Use X25519 (ECDH) for secure key exchange
- **AES-GCM Encryption**: Authenticated encryption that prevents data tampering
- **Object Serialization**: Encrypt and decrypt arbitrary Go objects directly
- **Base64 Encoding**: Keys and ciphertext are Base64 encoded for easy transmission

## Security

- Uses X25519 (ECDH) for secure key agreement
- Uses AES-GCM for authenticated encryption
- Keys and data are Base64 encoded for easy integration

## API Reference

### `NewDynamicAES(length int) *DynamicAES`

Creates a new DynamicAES instance and generates a random AES key of the specified length.

```go
client := cryptography.NewDynamicAES(32) // AES-256
```

### `(d *DynamicAES) GetKey(pk string) (string, error)`

Encrypts the current AES key using the provided public key (for client-side). Returns an encrypted package in Base64 format.

```go
encryptedPackage, err := client.GetKey(serverPublicKey)
if err != nil {
    log.Fatal(err)
}
// Send encryptedPackage to server
```

### `(d *DynamicAES) SetKey(sk string, encryptedPackage string) error`

Decrypts the encrypted package and sets the AES key using the provided private key (for server-side).

```go
server := cryptography.NewDynamicAES(32)
err := server.SetKey(serverPrivateKey, encryptedPackage)
if err != nil {
    log.Fatal(err)
}
```

### `(d *DynamicAES) Encrypt(obj any) (string, error)`

Encrypts an arbitrary object and returns the ciphertext as a Base64 string.

```go
type GameData struct {
    Gold int
    Pos  string
}

cipherText, err := client.Encrypt(GameData{Gold: 99, Pos: "10,20"})
if err != nil {
    log.Fatal(err)
}
```

### `(d *DynamicAES) Decrypt(cipherTextStr string, obj any) error`

Decrypts the Base64 ciphertext and unmarshals it into the provided object.

```go
var gameData GameData
err := server.Decrypt(cipherText, &gameData)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Gold: %d, Position: %s\n", gameData.Gold, gameData.Pos)
```

## Example: Client-Server Communication

### Client Side

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/cryptography"
)

type GameData struct {
	Gold int
	Pos  string
}

func main() {
	// 1. Initialize DynamicAES with AES-256
	client := cryptography.NewDynamicAES(32)

	// 2. Get encrypted key package using server's public key
	encryptedPackage, _ := client.GetKey(serverPublicKey)
	// Send encryptedPackage to server...

	// 3. Encrypt game data
	cipherText, _ := client.Encrypt(GameData{Gold: 99, Pos: "10,20"})
	fmt.Println("Data to send:", cipherText)
}
```

### Server Side

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/cryptography"
)

type GameData struct {
	Gold int
	Pos  string
}

func main() {
	// 1. Initialize DynamicAES
	server := cryptography.NewDynamicAES(32)

	// 2. Set AES key by decrypting the package from client
	err := server.SetKey(serverPrivateKey, encryptedPackage)
	if err != nil {
		panic(err)
	}

	// 3. Decrypt client's data
	var gameData GameData
	server.Decrypt(cipherText, &gameData)
	fmt.Printf("Server received: Gold=%d, Pos=%s\n", gameData.Gold, gameData.Pos)
}
```

## Usage Flow

1. **Client**: Generate random AES key with `NewDynamicAES(32)`
2. **Client**: Get encrypted key package with `GetKey(serverPublicKey)` and send to server
3. **Server**: Set AES key with `SetKey(serverPrivateKey, encryptedPackage)`
4. **Client/Server**: Use `Encrypt()` and `Decrypt()` for secure data transmission

## Notes

- Key length of 32 bytes uses AES-256
- ECDH uses X25519 curve for key agreement
- The encrypted key package contains: ephemeral public key (32 bytes) + encrypted AES key
- JSON serialization is used for object encryption/decryption
- Both encryption and decryption are thread-safe when using separate instances