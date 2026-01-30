# 加密 (Cryptography)

`cryptography` 套件提供了一套結合了現代加密標準的工具，包括動態 AES 密鑰管理與安全隨機數生成。

## 模組

### 1. 動態 AES (DynamicAES)

`DynamicAES` 是一個高級加密工具，結合了 **ECDH (X25519)** 密鑰協商與 **AES-GCM** 認證加密。它特別適用於需要在客戶端與服務端之間安全傳遞會話密鑰的場景。

#### 初始化

```go
import "github.com/leoheung/go-patterns/cryptography"

// 建立一個指定密鑰長度的加密器 (例如 32 字節對應 AES-256)
// 若長度大於 0，會自動生成隨機初始密鑰
aesTool := cryptography.NewDynamicAES(32)
```

#### 金鑰安全傳遞 (握手)

```go
// --- 客戶端 (App) ---
// 使用服務端的 Base64 公钥加密當前的 AES 密鑰
// 返回一個包含臨時公鑰和加密後的 AES 密鑰的封裝包 (Base64)
encryptedPackage, err := client.GetKey(serverPublicKeyStr)

// --- 服務端 (Server) ---
// 使用自己的私鑰解開數據包並將內部的 AES 密鑰設置為當前會話密鑰
err := server.SetKey(serverPrivateKeyStr, encryptedPackage)
```

#### 加密與解密

```go
// 加密任意對象 (內部自動進行 JSON 序列化)
// 返回 Base64 格式的密文 (包含 Nonce)
cipherText, err := aesTool.Encrypt(myData)

// 解密 Base64 密文到指定對象
err := aesTool.Decrypt(cipherText, &myResult)
```

### 2. 隨機數生成

提供加密安全的隨機值生成函數。

```go
// 生成指定長度的隨機字符串 (字符集：a-zA-Z0-9)
str := cryptography.RandString(16)

// 生成符合 RFC 4122 標準的 Version 4 UUID
uuid := cryptography.RandUUID() // 格式: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
```

## 完整範例 (模擬握手與通信)

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/cryptography"
)

func main() {
    // 1. 初始化
    client := cryptography.NewDynamicAES(32)
    server := cryptography.NewDynamicAES(0) // 服務端初始不帶密鑰

    // 2. 模擬金鑰傳遞 (假設已有 ECC 密鑰對)
    // encryptedPkg, _ := client.GetKey(pubKey)
    // server.SetKey(privKey, encryptedPkg)

    // 3. 數據加密通信
    type Message struct {
        Text string
        ID   int
    }
    
    msg := Message{"Secret Message", 101}
    cipherText, _ := client.Encrypt(msg)
    
    // 服務端解密
    var received Message
    server.Decrypt(cipherText, &received)
    
    fmt.Printf("解密結果: %+v\n", received)
}
```

## 特性

- **安全性**: 
    - 使用 **X25519** 進行 ECDH 密鑰協商，確保傳輸過程安全。
    - 使用 **AES-GCM** (具有 256 位哈希衍生的密鑰) 提供認證加密，防止數據被篡改。
- **靈活性**: 支持對 Go 中的任意 `any` 類型（Struct, Map, Slice 等）直接進行加密。
- **便捷性**: 所有輸入輸出均經過 Base64 編碼，方便通過 HTTP 或 JSON 進行傳輸。
- **合規性**: 隨機數生成均基於 `crypto/rand`，滿足加密安全要求。
