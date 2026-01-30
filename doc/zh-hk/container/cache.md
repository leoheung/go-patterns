# 快取

用於儲存及檢索數據的快取實現，支援 TTL。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/cache"
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
// 以指定 TTL 新增數值
err := c.Add("key", "value", 5*time.Minute)
if err != nil {
    // 處理錯誤
}
```

### 取得

```go
// 取得數值（若不存在或已過期則返回 nil）
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

    // 新增數值
    err = c.Add("user:1", "Alice", 5*time.Minute)
    if err != nil {
        fmt.Printf("新增用戶 1 錯誤: %v\n", err)
    }

    err = c.Add("user:2", "Bob", 10*time.Minute)
    if err != nil {
        fmt.Printf("新增用戶 2 錯誤: %v\n", err)
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

- **支援 TTL**: 項目會根據設定的時間自動過期並清理
- **線程安全**: 內部使用 RWMutex，支援高並發讀取
- **簡單 API**: 易於使用的 Add/Get 介面
- **基於優先級調度**: 使用 `PriorityScheduledTaskManager` 精確管理過期任務
