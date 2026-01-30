# 網絡 (Net)

`net` 套件提供 HTTP 回應處理、並發檔案下載及常用指針輔助函數。

## 安裝

```go
import "github.com/leoheung/go-patterns/net"
```

## API 參考

### 並發下載

使用多個 Goroutine 進行分片下載，並自動追蹤進度。

```go
// 使用 4 個併發數下載檔案到指定目錄
err := net.DownloadFileByConcurrent("https://example.com/file.zip", "./downloads/", 4)
```

### HTTP 回應工具

為 Web 服務提供標準化的 JSON 及 CSV 回應格式。

```go
// 返回標準化的 JSON 成功回應
net.ReturnJsonResponse(w, http.StatusOK, map[string]string{"message": "success"})

// 返回標準化的 JSON 錯誤回應
net.ReturnErrorResponse(w, http.StatusBadRequest, "輸入無效")

// 返回 CSV 檔案下載回應
headers := []string{"ID", "姓名"}
rows := [][]string{{"1", "Alice"}, {"2", "Bob"}}
net.ReturnCSVResponse(w, "users.csv", headers, rows)
```

### Chi 路由工具

```go
// 打印 Chi 路由器中所有已註冊的路由
net.PrintCHIRoutes(r)
```

### 指針輔助函數

用於快速建立基本類型的指針（在處理資料庫模型的可選字段時非常有用）。

```go
s := net.PtrString("hello")
i := net.PtrInt(100)
b := net.PtrBool(true)
t := net.PtrTime(time.Now())
```

## 完整範例

```go
package main

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/leoheung/go-patterns/net"
)

func main() {
    r := chi.NewRouter()

    r.Get("/api/data", func(w http.ResponseWriter, r *http.Request) {
        data := struct {
            ID   int    `json:"id"`
            Name string `json:"name"`
        }{ID: 1, Name: "Pattern"}
        
        net.ReturnJsonResponse(w, http.StatusOK, data)
    })

    // 打印路由結構用於調試
    net.PrintCHIRoutes(r)

    // 在背景並發下載檔案
    go net.DownloadFileByConcurrent("https://example.com/large-file.bin", "./tmp/", 8)

    http.ListenAndServe(":8080", r)
}
```

## 特性

- **並發下載**: 自動計算分片並支持斷點續傳邏輯（視服務端支持而定），自動提取原始文件名。
- **標準化回應**: 統一的 `UniversalResponse` 結構，包含 `isSuccess` 標誌。
- **路由可視化**: 方便地查看 Chi 路由器的層級結構與中間件。
- **類型指針化**: 簡化 Go 中基本類型轉指針的操作。
