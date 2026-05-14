# 重試機制

## 概述

`utils` 包提供了重試機制，用於執行可能失敗的函數，具有自動 panic 恢復和可配置的重試次數。

## API 參考

### `RetryWork(work func() (any, error), retryTimes int) (any, error)`

執行工作函數，捕獲 panic 或錯誤，最多重試 `retryTimes` 次。

**參數：**
- `work`：要執行的函數
- `retryTimes`：最大重試次數（不包括首次執行）

**返回：**
- 工作函數返回的數據
- 遇到的任何錯誤（在耗盡所有重試次數後返回）

```go
func RetryWork(work func() (any, error), retryTimes int) (any, error)
```

## 示例

```go
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// 模擬可能失敗的函數
	attempt := 0
	work := func() (any, error) {
		attempt++
		// 模擬隨機失敗（70% 失敗概率）
		if rand.Float32() < 0.7 {
			return nil, fmt.Errorf("操作在第 %d 次嘗試時失敗", attempt)
		}
		return fmt.Sprintf("第 %d 次嘗試成功", attempt), nil
	}

	// 最多重試 3 次（總共 4 次嘗試）
	result, err := utils.RetryWork(work, 3)

	if err != nil {
		fmt.Printf("最終錯誤: %v\n", err)
	} else {
		fmt.Printf("結果: %v\n", result)
	}
}
```

## 重試行為

1. 執行工作函數
2. 如果成功，立即返回
3. 如果失敗（錯誤或 panic）：
   - 記錄失敗
   - 如果還有重試次數，睡眠 500ms 後重試
   - 如果沒有重試次數，返回錯誤
4. 自動從 panic 中恢復並將其視為錯誤

## 注意事項

- Panic 會自動恢復並轉換為錯誤
- 重試之間插入 500ms 延遲
- 所有嘗試（包括失敗）都會被記錄
- 與適當的同步一起使用時，該函數是線程安全的