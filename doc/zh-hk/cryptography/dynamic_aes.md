# 動態 AES 加密

## 概述

`DynamicAES` 是一個動態 AES 密鑰管理與加密工具，結合了 ECDH 密鑰協商和 AES-GCM 加密。它適用於需要安全傳遞對稱密鑰的場景，例如客戶端與服務器之間的通信。

## 功能特點

- **動態 AES 密鑰生成**：生成任意長度的隨機 AES 密鑰
- **ECDH 密鑰協商**：使用 X25519 (ECDH) 安全交換密鑰
- **AES-GCM 加密**：提供認證加密，防止數據篡改
- **對象序列化**：直接加密和解密任意 Go 對象
- **Base64 編碼**：密鑰和密文以 Base64 格式傳輸，便於集成

## 安全性

- 使用 X25519 (ECDH) 進行安全的密鑰協商
- 使用 AES-GCM 提供認證加密
- 密鑰和數據以 Base64 編碼，便於集成

## API 參考

### `NewDynamicAES(length int) *DynamicAES`

創建一個新的 DynamicAES 實例並生成指定長度的隨機 AES 密鑰。

```go
client := cryptography.NewDynamicAES(32) // AES-256
```

### `(d *DynamicAES) GetKey(pk string) (string, error)`

使用提供的公鑰加密當前的 AES 密鑰（用於客戶端）。返回 Base64 格式的加密包。

```go
encryptedPackage, err := client.GetKey(serverPublicKey)
if err != nil {
    log.Fatal(err)
}
// 發送 encryptedPackage 到服務器
```

### `(d *DynamicAES) SetKey(sk string, encryptedPackage string) error`

使用提供的私鑰解密加密包並設置 AES 密鑰（用於服務器端）。

```go
server := cryptography.NewDynamicAES(32)
err := server.SetKey(serverPrivateKey, encryptedPackage)
if err != nil {
    log.Fatal(err)
}
```

### `(d *DynamicAES) Encrypt(obj any) (string, error)`

加密任意對象並返回 Base64 字符串格式的密文。

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

解密 Base64 密文並將其解組到提供的對象中。

```go
var gameData GameData
err := server.Decrypt(cipherText, &gameData)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("金幣: %d, 位置: %s\n", gameData.Gold, gameData.Pos)
```

## 示例：客戶端-服務器通信

### 客戶端

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
	// 1. 初始化 DynamicAES，使用 AES-256
	client := cryptography.NewDynamicAES(32)

	// 2. 使用服務器的公鑰獲取加密密鑰包
	encryptedPackage, _ := client.GetKey(serverPublicKey)
	// 發送 encryptedPackage 到服務器...

	// 3. 加密遊戲數據
	cipherText, _ := client.Encrypt(GameData{Gold: 99, Pos: "10,20"})
	fmt.Println("發送的數據:", cipherText)
}
```

### 服務器端

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
	// 1. 初始化 DynamicAES
	server := cryptography.NewDynamicAES(32)

	// 2. 使用私鑰解密客戶端的密鑰包並設置 AES 密鑰
	err := server.SetKey(serverPrivateKey, encryptedPackage)
	if err != nil {
		panic(err)
	}

	// 3. 解密客戶端的數據
	var gameData GameData
	server.Decrypt(cipherText, &gameData)
	fmt.Printf("服務器收到: 金幣=%d, 位置=%s\n", gameData.Gold, gameData.Pos)
}
```

## 使用流程

1. **客戶端**：使用 `NewDynamicAES(32)` 生成隨機 AES 密鑰
2. **客戶端**：使用 `GetKey(serverPublicKey)` 獲取加密密鑰包並發送到服務器
3. **服務器**：使用 `SetKey(serverPrivateKey, encryptedPackage)` 設置 AES 密鑰
4. **客戶端/服務器**：使用 `Encrypt()` 和 `Decrypt()` 進行安全數據傳輸

## 注意事項

- 密鑰長度 32 字節使用 AES-256
- ECDH 使用 X25519 曲線進行密鑰協商
- 加密密鑰包包含：臨時公鑰（32 字節）+ 加密的 AES 密鑰
- 對象加密/解密使用 JSON 序列化
- 使用不同實例時，加密和解密都是線程安全的