# 工具

`utils` 套件為常用操作提供工具函數。

## 模組

### 日誌記錄

檢查是否處於開發環境並相應地記錄消息。

```go
import "github.com/leoheung/go-patterns/utils"

// 檢查是否處於開發環境
isDev := utils.IsDev() // 若 env=dev 則返回 true

// 記錄消息 (開發環境使用 fmt.Println，生產環境使用 log.Println)
utils.LogMessage("Hello, world!")
```

### 重試

以錯誤/恐慌處理重試函數。

```go
// 以錯誤/恐慌處理重試函數
// work: 要執行的函數
// retryTimes: 最大重試次數 (不包括首次嘗試)
utils.RetryWork(
    func() error {
        // 可能失敗的操作
        return nil // 或錯誤
    },
    3, // 若失敗則重試 3 次
)
```

### 超時

以超時執行函數。

```go
// 以超時執行函數
err := utils.WithTimeout(5*time.Second, func() error {
    // 長時間運行的操作
    return nil
})
```

### 美化輸出

為除錯美化輸出對象。

```go
// 美化輸出對象
utils.PrettyPrint(obj)

// 帶標籤的美化輸出
utils.PrettyPrintWithLabel("用戶", user)
```

### 彩色輸出

以顏色輸出消息到控制台。

```go
// 以不同顏色輸出
utils.Red("錯誤消息")
utils.Green("成功消息")
utils.Yellow("警告消息")
utils.Blue("資訊消息")
```

### 數字工具

數字操作的輔助函數。

```go
// Min/Max 函數
min := utils.Min(1, 2, 3) // 1
max := utils.Max(1, 2, 3) // 3

// Clamp 函數
clamped := utils.Clamp(10, 0, 5) // 5
```

## 完整範例

```go
package main

import (
    "github.com/leoheung/go-patterns/utils"
    "time"
)

func main() {
    // 設定環境為開發環境
    // os.Setenv("env", "dev")
    
    // 重試可能失敗的操作
    utils.RetryWork(func() error {
        utils.LogMessage("嘗試操作中...")
        // 模擬失敗
        if time.Now().Nanosecond()%2 == 0 {
            panic("模擬恐慌")
        }
        return nil
    }, 3)
}
```

## 特性

- **環境感知**: 開發/生產環境的不同行為
- **錯誤處理**: 帶恐慌恢復的重試機制
- **除錯工具**: 美化輸出及彩色輸出
- **常用工具**: Min、Max、Clamp 函數
