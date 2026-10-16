# Goroutine 工具

## 概述

`utils` 套件提供 `GetGoid`，用於獲取當前 goroutine 的 ID。由於 Go 運行時並無官方方式暴露 goroutine ID，此函數會從堆疊追蹤（stack trace）輸出中解析取得。只建議用於除錯及記錄日誌用途。

## 安裝

```go
import "github.com/leoheung/go-patterns/utils"
```

## API 參考

### `GetGoid() uint64`

透過解析 `runtime.Stack` 的輸出，返回當前 goroutine 的 ID。如果無法解析堆疊輸出，返回 `0`。

```go
id := utils.GetGoid()
fmt.Printf("運行於 goroutine %d\n", id)
```

## 完整範例

```go
package main

import (
	"fmt"
	"sync"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	fmt.Println("main goroutine:", utils.GetGoid())

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			// 每個 goroutine 打印自己的 ID
			fmt.Printf("工作協程 %d 運行於 goroutine %d\n", n, utils.GetGoid())
		}(i)
	}
	wg.Wait()
}
```

## 特性

- **零分配解析**：重用固定的 64 字節緩衝區讀取堆疊輸出
- **穩健解析**：解析前會驗證 `"goroutine "` 前綴及欄位格式
- **安全回退**：當輸出無法解析時返回 `0`，不會令程序 panic

## 注意事項

- Go 官方不鼓勵依賴 goroutine ID，此函數只應用於除錯及日誌記錄
- goroutine 退出後其 ID 可能被重用
- 堆疊緩衝區上限為 64 字節，足以容納 `"goroutine N [狀態]:"` 這行標頭
