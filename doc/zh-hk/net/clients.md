# 共享 HTTP 客戶端

## 概述

`clients` 包提供了一個線程安全、可重用的 HTTP 客戶端，具有可配置的連接池和連接管理。它採用單例模式確保應用程序中只有一個 HTTP 客戶端實例。

## 功能特點

- **連接池**：可配置的最大空閒連接數、每個主機的最大連接數
- **TLS 配置**：自定義 TLS 設置支持
- **代理支持**：可配置的 HTTP 代理
- **Panic 恢復**：內置 panic 處理，防止崩潰
- **響應解析**：通用 JSON 響應解析

## HTTPClientConfig

共享 HTTP 客戶端的配置結構。

### 字段

| 字段 | 類型 | 默認值 | 描述 |
|------|------|--------|------|
| MaxIdleConns | `int` | `100` | 連接池中的最大空閒連接數 |
| MaxIdleConnsPerHost | `int` | `10` | 每個主機的最大空閒連接數 |
| MaxConnsPerHost | `int` | `100` | 每個主機的最大連接數 |
| IdleConnTimeout | `time.Duration` | `90s` | 空閒連接超時時間 |
| TLSConfig | `*tls.Config` | `nil` | TLS 配置 |
| CheckRedirect | `func(req *http.Request, via []*http.Request) error` | `nil` | 重定向策略 |
| Proxy | `func(*http.Request) (*url.URL, error)` | `nil` | 代理配置 |
| DisableKeepAlives | `bool` | `false` | 禁用 HTTP keep-alives |
| DisableCompression | `bool` | `false` | 禁用壓縮 |
| ForceAttemptHTTP2 | `bool` | `true` | 強制使用 HTTP/2 |

## API 參考

### `InitDefaultSharedHTTPClient()`

使用默認配置初始化共享 HTTP 客戶端。

```go
clients.InitDefaultSharedHTTPClient()
```

### `InitSharedHTTPClientWithConfig(config *HTTPClientConfig)`

使用自定義配置初始化共享 HTTP 客戶端。由於採用單例模式，此函數只能調用一次。

```go
config := &clients.HTTPClientConfig{
    MaxIdleConns:        200,
    MaxIdleConnsPerHost: 20,
    MaxConnsPerHost:     200,
    IdleConnTimeout:     60 * time.Second,
}
clients.InitSharedHTTPClientWithConfig(config)
```

### `Request(req *http.Request) (data []byte, headers http.Header, httpCode int, err error)`

使用共享客戶端執行 HTTP 請求。返回響應體、響應頭、狀態碼和任何錯誤。

```go
req, _ := http.NewRequest("GET", "https://api.example.com/data", nil)
data, headers, code, err := clients.Request(req)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Status: %d, Body: %s\n", code, string(data))
```

### `ParseResponse[T any](data []byte, dest *T) error`

將 JSON 響應數據解析為類型化結構體。

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

var user User
err := clients.ParseResponse(data, &user)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("User: %s <%s>\n", user.Name, user.Email)
```

## 示例

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/leoheung/go-patterns/net/clients"
)

func main() {
	// 1. 初始化共享 HTTP 客戶端
	clients.InitDefaultSharedHTTPClient()

	// 2. 創建帶有超時的請求（通過 context）
	req, _ := http.NewRequest("GET", "https://api.example.com/users/1", nil)
	req = req.WithContext(req.Context())

	// 3. 執行請求
	data, headers, statusCode, err := clients.Request(req)
	if err != nil {
		fmt.Printf("錯誤: %v\n", err)
		return
	}

	// 4. 處理響應
	fmt.Printf("狀態碼: %d\n", statusCode)
	fmt.Printf("Content-Type: %s\n", headers.Get("Content-Type"))
	fmt.Printf("響應體: %s\n", string(data))
}
```

## 注意事項

- 共享 HTTP 客戶端使用 `sync.Once` 確保只初始化一次
- 超時應該通過請求級別的 context 控制，而不是客戶端級別的超時
- 響應體在讀取後會自動關閉
- 內置 panic 恢復機制，防止請求失敗導致應用程序崩潰
- 連接池可顯著提高對同一主機的多次請求的性能