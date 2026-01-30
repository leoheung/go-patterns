# 網絡

`net` 套件為 HTTP 處理、檔案下載及常用網絡任務提供網絡工具。

## 模組

### Chi 路由工具

用於 Chi 路由器的工具。

```go
import "github.com/leoheung/go-patterns/net"

// 以中間件包裝處理器
handler := net.WrapHandler(myHandler)

// 常用中間件設定
r := chi.NewRouter()
net.SetupCommonMiddleware(r)
```

### 下載

帶進度追踪的檔案下載工具。

```go
// 帶進度下載檔案
err := net.DownloadFile("https://example.com/file.zip", 
    "/path/to/save", 
    func(progress float64) {
        fmt.Printf("下載進度: %.2f%%\n", progress*100)
    })

// 簡單下載
err := net.DownloadFileSimple("https://example.com/file.zip", "/path/to/save")
```

### 常用網絡工具

常用網絡輔助函數。

```go
// 檢查 URL 是否可達
reachable := net.IsReachable("https://example.com")

// 取得可用埠
port, err := net.GetFreePort()

// 解析 URL
parsed, err := net.ParseURL("https://example.com/path?query=value")
```

## 完整範例

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/net"
)

func main() {
    // 檢查網站是否可達
    if net.IsReachable("https://google.com") {
        fmt.Println("Google 可達")
    }
    
    // 取得可用埠
    port, err := net.GetFreePort()
    if err != nil {
        panic(err)
    }
    fmt.Printf("可用埠: %d\n", port)
    
    // 帶進度下載檔案
    err = net.DownloadFile(
        "https://example.com/file.zip",
        "/tmp/file.zip",
        func(progress float64) {
            fmt.Printf("進度: %.0f%%\n", progress*100)
        },
    )
    if err != nil {
        fmt.Printf("下載失敗: %v\n", err)
    }
}
```

## 特性

- **HTTP 工具**: 常用 HTTP 輔助函數
- **檔案下載**: 帶進度追踪的下載
- **Chi 整合**: 用於 Chi 路由器的工具
- **埠管理**: 尋找可用埠
