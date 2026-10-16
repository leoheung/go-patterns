# CircularStack

固定容量、後進先出（LIFO/FILO）嘅環形緩衝區，附帶變更通知訂閱。

專為生產者／消費者場景設計：生產者高頻率推送數據，消費者就慢啲讀取最新值。如果用普通 stack，快生產者 + 慢消費者會令記憶體無限增長。`CircularStack` 用固定大小限制記憶體：滿咗之後，每次 `Push` 都會覆蓋最舊嘅元素，所以緩衝區最多只保留最近推入嘅 `size` 個項目。

消費者可以訂閱變更通知，再用 `Peek` 或 `Pop` 攞最新值。Stack 可以暫停同恢復：暫停會停止通知並且結束背景通知 goroutine，恢復就會重新啟動佢（底層由 `RenewableContext` 驅動）。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/circular"
```

## API 參考

### 建立

```go
// 建立一個固定大小嘅環形棧。如果 size <= 0，會被強制設為 1。
func NewCircularStack[T any](size int) *CircularStack[T]
```

### Push / Pop / Peek

```go
// Push 將元素推入棧頂。
// Stack 滿咗（count == size）之後，最舊嘅元素會被覆蓋，
// 棧頂指針繼續前進。
// 如果 stack 處於暫停狀態，返回 error。
func (cs *CircularStack[T]) Push(data T) error

// Peek 返回最新元素但唔移除。
// Stack 為空時返回 error。（暫停狀態下都可以用。）
func (cs *CircularStack[T]) Peek() (T, error)

// Pop 移除並返回最新元素。
// Stack 暫停或者為空時返回 error。
func (cs *CircularStack[T]) Pop() (T, error)
```

### 暫停 / 恢復

```go
// Pause 停止通知，並結束背景通知 goroutine。
// 暫停期間 Push / Pop / Subscribe 都會返回 error。
func (cs *CircularStack[T]) Pause()

// Resume 復活 stack 並重新啟動背景通知 goroutine。
// 如果 stack 仍然存活，返回 error。
func (cs *CircularStack[T]) Resume() error
```

### 訂閱

```go
// Subscribe 註冊一個變更通知 channel。
// 每次成功嘅 Push 都會向每個訂閱者 channel 發送
// 一次 best-effort（非阻塞）通知。
//
//   - buffer：訂閱者 channel 嘅緩衝大小（小於 0 會被設為 0）
//   - 返回通知 channel、取消訂閱函數同 error
//     （stack 暫停時返回 error）
//
// 呼叫取消訂閱函數會移除訂閱並關閉 channel
// （可以安全多次呼叫——只有第一次會關閉）。
func (cs *CircularStack[T]) Subscribe(buffer int) (<-chan struct{}, func(), error)
```

> 注意：通知係合併（coalesced）嘅——內部通知 channel 嘅緩衝只有 1，所以一連串 `Push` 可能只會觸發一次通知。將 `Subscribe` 同 `Peek` 一齊用，先可以永遠讀到最新值。

## 完整範例

```go
package main

import (
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/container/circular"
)

func main() {
	// 建立一個保留最近 3 個項目嘅環形棧
	cs := circular.NewCircularStack[int](3)

	// 訂閱變更通知（緩衝大小為 4 嘅 channel）
	ch, unsubscribe, err := cs.Subscribe(4)
	if err != nil {
		panic(err)
	}
	defer unsubscribe()

	// 背景消費者：一有變更就醒嚟讀取最新元素
	go func() {
		for range ch {
			if top, err := cs.Peek(); err == nil {
				fmt.Println("最新:", top)
			}
		}
	}()

	// 生產者：推入 5 個元素到大小為 3 嘅 stack——1 同 2 會被覆蓋
	for i := 1; i <= 5; i++ {
		if err := cs.Push(i); err != nil {
			panic(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Pop 最新元素
	top, err := cs.Pop()
	if err != nil {
		panic(err)
	}
	fmt.Println("彈出:", top) // 5

	// 暫停 stack：通知停止，Push/Pop 返回 error
	cs.Pause()
	if err := cs.Push(6); err != nil {
		fmt.Println("暫停時 push:", err) // CircularStack is paused
	}

	// Resume 重新啟動通知 goroutine
	if err := cs.Resume(); err != nil {
		panic(err)
	}
}
```

## 特性

- **固定容量、記憶體有界**：滿咗之後 `Push` 會覆蓋最舊嘅元素——最適合高頻生產者 + 慢消費者嘅場景
- **LIFO 語義**：`Peek` / `Pop` 永遠操作最近推入嘅項目
- **訂閱通知**：`Subscribe` 返回 channel 加取消訂閱函數；每次 `Push` 都會通知訂閱者（best-effort、非阻塞）
- **暫停 / 恢復**：底層用 `RenewableContext`——暫停會停止通知同背景 goroutine，恢復就會復活佢哋
- **並發安全**：所有操作都由讀寫鎖保護
- **支援泛型**：任何元素類型都用得
