# 加密

`cryptography` 套件提供加密工具。

## 模組

### 動態 AES

支援密鑰輪換的動態 AES 加密。

```go
import "github.com/leoheung/go-patterns/cryptography" 

// 以密鑰建立 AES 加密器
encrypter := cryptography.NewDynamicAES([]byte("your-32-byte-key-here!!!!!!!!!!!!"))

// 加密數據
encrypted, err := encrypter.Encrypt([]byte("sensitive data"))
if err != nil {
    // 處理錯誤
}

// 解密數據
decrypted, err := encrypter.Decrypt(encrypted)
if err != nil {
    // 處理錯誤
}
```

### 隨機數

生成加密安全的隨機值。

```go
// 生成隨機字節
randomBytes, err := cryptography.GenerateRandomBytes(32)

// 生成隨機字串
randomString, err := cryptography.GenerateRandomString(16)

// 生成隨機整數
randomInt := cryptography.GenerateRandomInt(100)
```

## 完整範例

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/cryptography"
)

func main() {
    // 生成隨機密鑰
    key, _ := cryptography.GenerateRandomBytes(32)
    
    // 建立加密器
    aes := cryptography.NewDynamicAES(key)
    
    // 加密
    plaintext := []byte("Hello, World!")
    ciphertext, err := aes.Encrypt(plaintext)
    if err != nil {
        panic(err)
    }
    
    // 解密
    decrypted, err := aes.Decrypt(ciphertext)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("原始: %s\n", plaintext)
    fmt.Printf("解密: %s\n", decrypted)
}
```

## 特性

- **AES 加密**: 業界標準加密
- **密鑰輪換**: 支援動態密鑰變更
- **安全隨機**: 加密安全的隨機生成
- **簡單 API**: 易於使用的加密/解密介面
