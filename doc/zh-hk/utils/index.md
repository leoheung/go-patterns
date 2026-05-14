# 工具 (Utils)

`utils` 套件提供了一系列常用操作的工具函數，涵蓋日誌記錄、重試機制、超時控制、對象美化及數值處理等。

## 模組

### [日誌記錄](./log.md)
自動識別開發或生產環境並採取不同的記錄方式。

### [重試](./retry.md)
為不穩定的操作提供重試機制。

### [超時控制](./timeout.md)
為操作添加超時限制。

### [對象美化](./pretty.md)
基於 `go-spew` 實現的高級美化工具。

### [彩色終端](./color.md)
使用 ANSI 轉義序列輸出彩色文字。

### [數值工具](./number.md)
模擬動態類型語言中的數字處理。

### [通用輔助](./common.md)
常見的輔助函數。

### [Channel 操作](./channel.md)
非阻塞和帶超時的 Channel 操作。

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
