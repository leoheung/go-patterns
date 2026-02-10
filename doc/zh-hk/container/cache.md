# 快取

用於儲存及檢索數據的快取實現，支援 TTL 及永久緩存。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/cache"
import "github.com/leoheung/go-patterns/net" // 用於 Ptr 輔助函數
```

## API 參考

### 建立快取

```go
// 建立新快取
c, err := cache.NewCache()
if err != nil {
    // 處理錯誤
}
```

### 新增

```go
// 以指定 TTL 新增數值 (需要傳入 *time.Duration)
duration := 5 * time.Minute
err := c.Add("key", "value", &duration)

// 新增永久緩存 (傳入 nil)
err := c.Add("permanent_key", "value", nil)
```

### 取得

```go
// 取得數值（若不存在或已過期則返回 nil）
// 對於非永久緩存，每次 Get 會自動重置過期時間
value := c.Get("key")
if value != nil {
    // 使用數值
}
```

### 刪除

```go
// 刪除指定的鍵
c.Delete("key")
```

### 快取狀態

```go
// 取得快取狀態字符串（包含項目數量及調度器狀態）
status := c.String()
```

## 完整範例

```go
package main

import (
    "fmt"
    "time"
    "github.com/leoheung/go-patterns/container/cache"
)

func main() {
    // 建立快取
    c, err := cache.NewCache()
    if err != nil {
        fmt.Printf("建立快取錯誤: %v\n", err)
        return
    }

    // 1. 新增帶過期時間的數值
    duration := 5 * time.Minute
    err = c.Add("user:1", "Alice", &duration)
    if err != nil {
        fmt.Printf("新增用戶 1 錯誤: %v\n", err)
    }

    // 2. 新增永久緩存 (nil duration)
    err = c.Add("config:version", "v1.0.0", nil)
    if err != nil {
        fmt.Printf("新增配置錯誤: %v\n", err)
    }

    // 取得數值
    value := c.Get("user:1")
    if value != nil {
        fmt.Printf("用戶 1: %v\n", value)
    }

    // 刪除鍵
    c.Delete("user:1")

    // 打印快取狀態
    fmt.Println("快取狀態:")
    fmt.Println(c.String())
}
```

## 特性

- **靈活的過期控制**: 支援自動過期 (TTL) 與永久緩存 (`nil` duration)。
- **自動續期**: 對於有 TTL 的項目，每次 `Get` 訪問會自動重置過期時間。
- **線程安全**: 內部使用 RWMutex，支援高並發讀取。
- **基於優先級調度**: 使用 `PriorityScheduledTaskManager` 精確管理過期任務。
