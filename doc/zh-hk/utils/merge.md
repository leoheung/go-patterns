# Context 合併

## 概述

`utils` 套件提供 `MergeContexts`，用於將兩個 `context.Context` 合併成一個。合併後的 context 會在其中任何一個原始 context 被取消時，或者調用返回的 `CancelFunc` 時被取消。適用於一個操作需要同時尊重多個取消來源的場景（例如請求 context 同伺服器關閉 context）。

## 安裝

```go
import "github.com/leoheung/go-patterns/utils"
```

## API 參考

### `MergeContexts(ctx1, ctx2 context.Context) (context.Context, context.CancelFunc)`

將 `ctx1` 和 `ctx2` 合併為單一 context。返回的 context 衍生自 `ctx1`，並會在 `ctx1` 或 `ctx2` 任一完成時被取消。返回合併後的 context 及一個 `CancelFunc`，該函數會停止內部掛鈎並取消合併後的 context。

如果其中一個輸入 context 已經被取消，會立即返回一個已取消的 context（快速路徑）。

```go
ctx1, cancel1 := context.WithCancel(context.Background())
defer cancel1()

ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
defer cancel2()

merged, cancel := utils.MergeContexts(ctx1, ctx2)
defer cancel()

// 當 ctx1 被取消、ctx2 超時，或者調用 cancel() 時，merged 即完成
<-merged.Done()
```

## 完整範例

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// 請求 context（例如客戶端斷線時被取消）
	reqCtx, cancelReq := context.WithCancel(context.Background())
	defer cancelReq()

	// 帶期限的伺服器 context（例如關閉超時）
	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelTimeout()

	// 合併後的 context 會在任一 context 完成時立即完成
	merged, cancel := utils.MergeContexts(reqCtx, timeoutCtx)
	defer cancel()

	select {
	case <-merged.Done():
		fmt.Println("合併後的 context 已完成:", merged.Err())
	case <-time.After(time.Second):
		fmt.Println("操作已完成")
	}
	// 輸出（約 100 毫秒後）：
	// 合併後的 context 已完成: context deadline exceeded
}
```

## 特性

- **雙向取消**：任一輸入 context 被取消時，合併後的 context 亦會被取消
- **快速路徑**：如果任一輸入已被取消，立即返回已取消的 context
- **掛鈎式傳播**：內部使用 `context.AfterFunc` 傳播 `ctx2` 的取消信號，無需額外 goroutine
- **乾淨關閉**：返回的 `CancelFunc` 會停止內部掛鈎並取消合併後的 context，避免資源洩漏

## 注意事項

- 務必調用返回的 `CancelFunc`（通常用 `defer`）以釋放資源
- 鍵值（Value）只從 `ctx1` 繼承（合併後的 context 衍生自 `ctx1`）
- 需要 Go 1.21 或以上版本（`context.AfterFunc`）
