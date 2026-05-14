# 隨機數生成

## 概述

`cryptography` 包提供了密碼學安全的隨機數生成工具，包括隨機字符串和 UUID 生成。

## RandString

使用密碼學安全的隨機數生成指定長度的隨機字符串。

### API 參考

#### `RandString(n int) string`

返回長度為 `n` 的隨機字符串，使用字符 `a-zA-Z0-9`。如果 `n <= 0` 返回空字符串。

```go
randomStr := cryptography.RandString(16)
fmt.Println(randomStr) // 例如: "aB3xY9kLmN2pQrT4"
```

### 示例

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/cryptography"
)

func main() {
	// 生成不同長度的隨機字符串
	for i := 4; i <= 16; i += 4 {
		fmt.Printf("長度 %d: %s\n", i, cryptography.RandString(i))
	}

	// 生成隨機密碼
	password := cryptography.RandString(32)
	fmt.Printf("隨機密碼: %s\n", password)
}
```

## RandUUID

使用密碼學安全的隨機數生成符合 RFC 4122 標準的版本 4 UUID。

### API 參考

#### `RandUUID() string`

返回格式為 `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx` 的 UUID。如果隨機數生成失敗，返回空字符串。

```go
uuid := cryptography.RandUUID()
fmt.Println(uuid) // 例如: "550e8400-e29b-41d4-a716-446655440000"
```

### 示例

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/cryptography"
)

func main() {
	// 生成多個 UUID
	for i := 0; i < 5; i++ {
		uuid := cryptography.RandUUID()
		fmt.Printf("UUID %d: %s\n", i+1, uuid)
	}
}
```

## 注意事項

- 兩個函數都使用 `crypto/rand` 進行密碼學安全的隨機數生成
- `RandString` 使用小寫字母、大寫字母和數字（總共 62 個字符）
- `RandUUID` 生成符合 RFC 4122 標準的版本 4 UUID
- 兩個函數都安全可用於安全敏感的應用場景，如會話 ID、令牌和加密密鑰