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

### 流式下載

將任意 `io.Reader` 流式傳輸到 HTTP 回應，自動設置正確的標頭。支援固定大小和分塊傳輸模式。

```go
// 已知檔案大小的流式下載
fileData := bytes.NewReader(fileBytes)
size := int64(len(fileBytes))
net.StreamDownloadHandler(w, fileData, "report.pdf", "application/pdf", &size)

// 未知檔案大小的流式下載（使用分塊傳輸）
s3Reader := getS3ObjectReader(key)
net.StreamDownloadHandler(w, s3Reader, "backup.zip", "application/zip", nil)
```

**參數說明：**

- `w`: HTTP ResponseWriter
- `reader`: 任意 io.Reader（檔案、記憶體緩衝、S3 物件等）
- `filename`: 顯示給用戶的下載檔案名稱
- `contentType`: MIME 類型（例如 "application/pdf", "application/octet-stream"）
- `size`: 檔案大小指標（可選，傳入 nil 則使用分塊傳輸）

### WebSocket 管理器

高級 WebSocket 連接管理框架。詳見 [WebSocket 文檔](./wrapsocket)。

```go
import "github.com/leoheung/go-patterns/net/wrapsocket"

// 創建 WebSocket Handler
handler := wrapsocket.NewDefaultHandler(nil)

// 設置回調
handler.SetOnConnect(func(conn *wrapsocket.Conn) {
    fmt.Printf("客戶端已連接: %s\n", conn.ID)
})

handler.SetOnMessage(func(conn *wrapsocket.Conn, msg *wrapsocket.Message) {
    // 回傳消息
    conn.Write(ctx, msg.Type, msg.Data)
})

http.ListenAndServe(":8080", handler)
```

**特性：**

- 連接生命週期管理
- 心跳檢測
- 分組廣播
- 元數據存儲

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

### 共享 HTTP 客戶端

線程安全的可重用 HTTP 客戶端，具有連接池功能。詳見 [客戶端文檔](./clients)。

```go
// 初始化共享 HTTP 客戶端
clients.InitDefaultSharedHTTPClient()

// 發送請求
resp, headers, code, err := clients.Request(req)
```

**特性：**
- 可配置的連接池限制
- TLS 和代理配置
- 內置 Panic 恢復
- 通用 JSON 響應解析

### 指針輔助函數

用於快速建立基本類型的指針（在處理資料庫模型的可選字段時非常有用）。

```go
s := net.PtrString("hello")
i := net.PtrInt(100)
b := net.PtrBool(true)
t := net.PtrTime(time.Now())
```

### 安全讀取 Body

安全地讀取 HTTP 請求/回應 Body，限制大小以防止記憶體問題。

```go
// 讀取 Body，限制大小（以 MB 為單位）
data, err := net.SafelyReadBody(r.Body, net.PtrInt(10))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("讀取了 %d 位元組\n", len(data))
```

### 深拷貝 HTTP 請求

創建 HTTP 請求的完整深拷貝，包括 Body 內容。

```go
// 深拷貝請求，Body 限制為 10MB
reqCopy, err := net.DeepCopyRequest(r, 10)
if err != nil {
    log.Fatal(err)
}

// 現在您可以處理 reqCopy 而不影響原始的 r.Body
process(reqCopy)
next.ServeHTTP(w, r) // 原始請求的 Body 仍然完好
```

### 深拷貝 HTTP 回應

創建 HTTP 回應的完整深拷貝，包括 Body 內容。

```go
// 深拷貝回應，Body 限制為 10MB
respCopy, err := net.DeepCopyResponse(resp, 10)
if err != nil {
    log.Fatal(err)
}

// 現在您可以處理 respCopy 而不影響原始回應
cacheResponse(respCopy)
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
- **安全 Body 讀取**: 通過限制 Body 大小來防止記憶體問題。
- **深拷貝請求/回應**: 完整的 HTTP 消息深拷貝，用於中間件處理。
