# 超時控制

## 概述

`utils` 包提供了超時機制，用於執行可能掛起或花費太長時間完成的函數。

## API 參考

### `TimeoutWork(work func() (any, error), timeout time.Duration) (any, error)`

使用超時限制執行工作函數。如果函數在指定時間內未完成，則返回超時錯誤。

**參數：**
- `work`：要執行的函數
- `timeout`：等待完成的最大持續時間

**返回：**
- 工作函數返回的數據
- 如果函數未及時完成，則返回超時錯誤
- 工作函數返回的任何錯誤

```go
func TimeoutWork(work func() (any, error), timeout time.Duration) (any, error)
```

## 示例

```go
package main

import (
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// 示例 1：及時完成的函數
	fastWork := func() (any, error) {
		time.Sleep(100 * time.Millisecond)
		return "任務完成", nil
	}

	result, err := utils.TimeoutWork(fastWork, 1*time.Second)
	if err != nil {
		fmt.Printf("錯誤: %v\n", err)
	} else {
		fmt.Printf("結果: %v\n", result)
	}

	// 示例 2：超時的函數
	slowWork := func() (any, error) {
		time.Sleep(5 * time.Second) // 模擬慢操作
		return "任務完成", nil
	}

	result, err = utils.TimeoutWork(slowWork, 1*time.Second)
	if err != nil {
		fmt.Printf("超時錯誤: %v\n", err)
	} else {
		fmt.Printf("結果: %v\n", result)
	}
}
```

## 行為

1. 為指定的超時持續時間啟動計時器
2. 在單獨的 goroutine 中執行工作函數
3. 等待以下任一情況：
   - 工作函數完成 → 返回結果
   - 超時到期 → 返回超時錯誤
4. 自動從工作函數中的 panic 恢復

## 注意事項

- 工作函數即使超時到期也會在其 goroutine 中一直運行到完成
- 內置 panic 恢復以防止崩潰
- 超時是精確的，不會阻塞調用的 goroutine
- 適用於網絡請求、數據庫查詢和其他可能很慢的操作