# 快取

用於儲存及檢索數據的快取實現，支援 TTL。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/cache"
```

## API 參考

### 建立快取

```go
// 以預設設定建立新快取
c := cache.NewCache()

// 以自定義 TTL 建立快取
c := cache.NewCacheWithTTL(5 * time.Minute)
```

### 設定

```go
// 以預設 TTL 設定數值
c.Set("key", "value")

// 以自定義 TTL 設定數值
c.SetWithTTL("key", "value", 10*time.Minute)
```

### 取得

```go
// 取得數值
if value, ok := c.Get("key"); ok {
    // 使用數值
}
```

### 刪除

```go
// 刪除鍵
c.Delete("key")

// 清空所有項目
c.Clear()
```

### 快取操作

```go
// 檢查鍵是否存在
exists := c.Has("key")

// 取得快取大小
size := c.Len()
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
    // 建立 5 分鐘 TTL 的快取
    c := cache.NewCacheWithTTL(5 * time.Minute)

    // 設定數值
    c.Set("user:1", "Alice")
    c.Set("user:2", "Bob")

    // 取得數值
    if user, ok := c.Get("user:1"); ok {
        fmt.Printf("用戶 1: %s\n", user)
    }

    // 以自定義 TTL 設定
    c.SetWithTTL("session:abc", "active", 30*time.Minute)

    // 檢查存在性
    if c.Has("user:2") {
        fmt.Println("用戶 2 存在")
    }

    // 取得快取大小
    fmt.Printf("快取大小: %d\n", c.Len())

    // 刪除鍵
    c.Delete("user:1")

    // 清空所有
    c.Clear()
}
```

## 特性

- **支援 TTL**: 項目自動過期
- **線程安全**: 可安全地並發使用
- **簡單 API**: 易於使用的 Get/Set 介面
- **記憶體高效**: 自動清理過期項目
