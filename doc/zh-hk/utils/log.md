# 日誌記錄

## 概述

`utils` 包提供具有環境感知功能的日誌工具，自動在開發和生產日誌模式之間切換。

## API 參考

### `IsDev() bool`

檢查應用程序是否在開發環境中運行（當 `env` 環境變量等於 `"dev"` 時）。

```go
if utils.IsDev() {
    fmt.Println("運行在開發模式")
}
```

### `LogMessage(message string)`

記錄消息。在開發模式下使用 `fmt.Println`，在生產模式下使用 `log.Println`。

```go
utils.LogMessage("應用程序已啟動")
```

### `DevLogError(errMsg string)`

以紅色顯示帶時間戳的錯誤消息。

```go
utils.DevLogError("連接數據庫失敗")
// 輸出: [Dev Logs] - 2024-01-01 12:00:00: 連接數據庫失敗
```

### `DevLogInfo(infoMsg string)`

以亮藍色顯示帶時間戳的信息消息。

```go
utils.DevLogInfo("正在處理請求")
```

### `DevLogSuccess(successMsg string)`

以綠色顯示帶時間戳的成功消息。

```go
utils.DevLogSuccess("用戶登錄成功")
```

## 示例

```go
package main

import (
	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// 檢查環境
	if utils.IsDev() {
		utils.LogMessage("運行在開發模式")
	}

	// 根據環境記錄消息
	utils.LogMessage("應用程序已啟動")

	// 僅在開發模式下顯示彩色日誌
	utils.DevLogError("這是一個錯誤")
	utils.DevLogInfo("這是一條信息")
	utils.DevLogSuccess("這是一條成功消息")
}
```

## 注意事項

- 設置 `env=dev` 環境變量以啟用開發模式
- 開發日誌為彩色並包含時間戳
- 生產日誌使用標準庫 `log` 包