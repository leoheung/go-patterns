# RenewableContext

可復活嘅 `context.Context`，附帶泛型數據載體。

標準 Go 嘅 context 係一次性：一旦取消或者超時，`Done()` channel 就永久關閉，唔可以再用。`RenewableContext[T]` 解決咗呢個問題——已死嘅 context 可以被**復活（Reactivate）**：內部會建立一個全新嘅 context，並且通知所有訂閱者；同時佢仲附帶一個類型安全嘅泛型數據載體（`T`），由讀寫鎖保護。

佢完整實作咗 `context.Context` 介面，可以傳入任何接收標準 context 嘅函數。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/context"
```

> 注意：呢個套件嘅名稱係 `context`，通常會用別名匯入，避免同標準庫撞名：
>
> ```go
> import lctx "github.com/leoheung/go-patterns/container/context"
> ```

## 概念

- **可復活**：`Cancel()`（或者超時）會殺死 context，但 `Reactivate()` 可以用一個全新嘅內部 context 令佢復活，`Done()` 之後會返回新嘅 channel。
- **泛型數據載體**：`GetData()` / `SetData()` 喺 context 旁邊存放一個 `T` 類型嘅值，由 `sync.RWMutex` 保護，並發使用安全。
- **復活訂閱**：`SubscribeReactivation()` 返回一個 channel，每次 context 復活時都會收到通知。背景 worker 可以等呢個通知嚟重啟。
- **合併**：`MergeContext(otherCtx)` 將另一個 context 嘅生命週期綁入嚟——邊個先死，合併後嘅 context 就會被取消。

## API 參考

### 建立

```go
// 建立一個帶數據載體嘅 RenewableContext。
// 如果 timeout 為 nil，context 冇截止時間；否則喺 *timeout 之後超時。
func NewRenewableContext[T any](timeout *time.Duration, data T) *RenewableContext[T]
```

### 生命週期

```go
// 取消 context（殺死佢；之後可以用 Reactivate 復活）。
func (gt *RenewableContext[T]) Cancel()

// Done 返回當前內部 context 嘅 done channel。
// Reactivate 之後會返回新嘅 channel。
func (gt *RenewableContext[T]) Done() <-chan struct{}

// IsAlive 檢查當前內部 context 係咪仍然存活。
func (gt *RenewableContext[T]) IsAlive() bool

// Reactivate 用一個全新嘅內部 context 復活已死嘅 context，
// 並通知所有復活訂閱者。
// 如果當前 context 仍然存活，會返回 error。
func (gt *RenewableContext[T]) Reactivate(timeout *time.Duration) error

// SubscribeReactivation 返回一個 channel，每次 Reactivate 成功時
// 都會收到通知。呢個 channel 永遠唔會被關閉。
//
// 注意：通知係非阻塞嘅 best-effort 發送——要確保有接收者已經
// 緊盯住個 channel 先好呼叫 Reactivate，否則通知會被丟棄。
func (gt *RenewableContext[T]) SubscribeReactivation() <-chan struct{}
```

### 泛型數據載體

```go
// GetData 返回當前數據（並發安全）。
func (gt *RenewableContext[T]) GetData() T

// SetData 替換數據（並發安全）。
func (gt *RenewableContext[T]) SetData(newData T)
```

### Context 合併

```go
// MergeContext 將 otherCtx 合併入呢個 context：只要當前 context
// 或者 otherCtx 其中之一被取消，合併後嘅 context 就會被取消。
func (gt *RenewableContext[T]) MergeContext(otherCtx context.Context)
```

### context.Context 介面

`RenewableContext[T]` 實作咗 `context.Context`（套件內有編譯期驗證：`var _ context.Context = new(RenewableContext[struct{}])`）：

```go
func (gt *RenewableContext[T]) Deadline() (deadline time.Time, ok bool)
func (gt *RenewableContext[T]) Done() <-chan struct{}
func (gt *RenewableContext[T]) Err() error
func (gt *RenewableContext[T]) Value(key any) any
```

呢四個方法都喺讀鎖下轉發畀當前嘅內部 context。

## 完整範例

```go
package main

import (
	"context"
	"fmt"
	"time"

	lctx "github.com/leoheung/go-patterns/container/context"
)

func main() {
	// 建立一個 50ms 超時、帶字符串數據嘅 context
	timeout := 50 * time.Millisecond
	ctx := lctx.NewRenewableContext(&timeout, "hello")

	// 佢實作咗 context.Context——可以當標準 context 咁用
	var _ context.Context = ctx

	// 訂閱復活通知（例如一個背景 worker）
	// 通知係非阻塞發送，呼叫 Reactivate 之前要有接收者緊盯住
	resumeCh := ctx.SubscribeReactivation()
	notified := make(chan struct{})
	go func() {
		<-resumeCh
		close(notified)
	}()

	// 等超時殺死 context
	time.Sleep(100 * time.Millisecond)
	fmt.Println("超時後存活:", ctx.IsAlive()) // false

	// 復活 context
	if err := ctx.Reactivate(nil); err != nil {
		panic(err)
	}
	fmt.Println("復活後存活:", ctx.IsAlive()) // true

	// 訂閱者收到通知
	<-notified
	fmt.Println("已收到復活通知")

	// 泛型數據載體
	fmt.Println("數據:", ctx.GetData()) // hello
	ctx.SetData("world")
	fmt.Println("數據:", ctx.GetData()) // world

	// 取消之後再復活
	ctx.Cancel()
	_ = ctx.Reactivate(nil)

	// 合併另一個 context：任何一邊死，合併後嘅 context 都會死
	other, cancelOther := context.WithCancel(context.Background())
	ctx.MergeContext(other)
	cancelOther()
	time.Sleep(10 * time.Millisecond)
	fmt.Println("合併取消後存活:", ctx.IsAlive()) // false
}
```

## 特性

- **可復活 Context**：唔似 `context.Context`，已取消／已超時嘅 context 可以用 `Reactivate` 復活
- **實作 context.Context**:可以無縫替換任何接收標準 context 嘅 API
- **泛型數據載體**:喺 context 旁邊帶住類型安全嘅數據(`T`)，內置 `RWMutex` 保護
- **復活訂閱**:背景 goroutine 可以等 `SubscribeReactivation()`，context 復活時自動重啟
- **Context 合併**:用 `MergeContext` 綁定外部 context 嘅生命週期——任何一邊結束，合併後嘅 context 就取消
- **並發安全**:所有方法都由讀寫鎖保護
