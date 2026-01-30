# 工具 (Utils)

`utils` 套件提供了一系列常用操作的工具函數，涵蓋日誌記錄、重試機制、超時控制、對象美化及數值處理等。

## 模組

### 1. 日誌記錄與環境感知

自動識別開發或生產環境並採取不同的記錄方式。

```go
import "github.com/leoheung/go-patterns/utils"

// 檢查是否處於開發環境 (判斷環境變量 env == "dev")
isDev := utils.IsDev()

// 記錄消息 (開發環境使用 fmt.Println，生產環境使用標準庫 log)
utils.LogMessage("Hello, world!")

// 開發環境專用彩色日誌 (僅在 IsDev() 為 true 時輸出)
utils.DevLogError("這是一條錯誤消息")
utils.DevLogInfo("這是一條資訊消息")
utils.DevLogSuccess("這是一條成功消息")
```

### 2. 重試與超時控制

為不穩定的操作提供保護機制。

```go
// 重試工作函數
// work: 執行函數，返回 (any, error)
// retryTimes: 失敗後重試的次數
data, err := utils.RetryWork(func() (any, error) {
    // 業務邏輯
    return "result", nil
}, 3)

// 帶超時的執行
// 若超時則返回 fmt.Errorf("timeout")
res, err := utils.TimeoutWork(func() (any, error) {
    time.Sleep(2 * time.Second)
    return "done", nil
}, 1 * time.Second)
```

### 3. 對象美化與 JSON 處理

基於 `go-spew` 實現的高級美化工具。

```go
// 格式化打印任意對象 (常用於除錯)
utils.PPrint(myStruct)
utils.PPrettyPrint(myStruct)

// 取得對象的漂亮字符串表示 (不直接打印)
str := utils.PrettyObjStr(myStruct)

// 將對象序列化為漂亮的 JSON 字符串 (失敗則回退到 PrettyObjStr)
jsonStr := utils.JSONalizeStr(myStruct)

// 將 JSON 字符串反序列化到對象 (必須傳入指針)
err := utils.DeJSONalizeStr(jsonStr, &myTarget)
```

### 4. 彩色終端輸出

```go
// 使用 ANSI 轉義序列輸出彩色文字
utils.PrintlnColor(utils.Red, "紅色文字")
utils.PrintlnColor(utils.Green, "綠色文字")
utils.PrintlnColor(utils.BrightBlue, "亮藍色文字")

// 可用顏色常量：
// utils.Red, utils.Green, utils.BrightBlue, utils.Magenta, utils.Cyan
```

### 5. 數值工具 (Number)

模擬動態類型語言中的數字處理。

```go
// 解析字符串為 Number 對象
n, err := utils.ParseNumber("100.5")

// 取得不同類型的數值
f := n.Float()   // 100.5
i := n.Int()     // 100
i64 := n.Int64() // 100

// 檢查是否為整數 (無小數部分)
isUint := n.IsInteger() // false
```

### 6. 通用輔助函數

```go
// 檢查任何值是否為 nil (支援 Interface, Slice, Map, Ptr 等)
isNull := utils.IsNil(someVar)

// 檢查字符串是否全由數字組成
allDigits := utils.IsDigits("12345")

// 延遲執行函數 (阻塞)
utils.DelayDo(500 * time.Millisecond, func() {
    fmt.Println("延遲執行")
})

// 無限期阻塞當前 Goroutine
utils.Hold()
```

## 完整範例

```go
package main

import (
    "fmt"
    "github.com/leoheung/go-patterns/utils"
)

func main() {
    // 1. 格式化輸出
    user := struct {
        Name string
        Age  int
    }{"Leon", 25}
    
    fmt.Println("用戶資料：")
    utils.PPrint(user)

    // 2. 重試邏輯
    count := 0
    utils.RetryWork(func() (any, error) {
        count++
        if count < 2 {
            return nil, fmt.Errorf("暫時性錯誤")
        }
        return "成功", nil
    }, 3)
}
```

## 特性

- **健壯性**: 重試與超時函數內部均包含 `recover()`，可防止業務邏輯 Panic 導致程序崩潰。
- **易用性**: 簡化了 Go 語言中繁瑣的類型指針轉換與反射檢查。
- **除錯友好**: 提供多種層次的对象序列化與打印工具。
