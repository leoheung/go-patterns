# ValueProvider

可觀察值 + 訂閱者通知（泛型）。

`ValueProvider[T]` 包住一個 `T` 類型嘅值，每次值變更都會通知所有訂閱者。每個成功嘅 `Set` 都會將新值廣播畀每個訂閱者嘅 channel。佢支援暫停／恢復（底層用 `RenewableContext`），亦支援安全取消訂閱。

## 安裝

```go
import "github.com/leoheung/go-patterns/container/subscribe"
```

## API 參考

### 建立

```go
// 建立一個帶初始值嘅 ValueProvider。
// buffer_size 係內部廣播 channel 嘅緩衝大小（小於 0 會被設為 0）。
func NewValueProvider[T any](val T, buffer_size int) *ValueProvider[T]
```

### Get / Set

```go
// Get 返回當前值（並發安全）。
func (v *ValueProvider[T]) Get() T

// Set 儲存新值並廣播畀所有訂閱者。
// 如果 provider 處於暫停狀態，返回 error。
func (v *ValueProvider[T]) Set(val T) error
```

### 訂閱

```go
// Subscribe 註冊一個訂閱者 channel。
//
//   - buffer：訂閱者 channel 嘅緩衝大小（小於 0 會被設為 0）
//   - 返回值 channel、取消訂閱函數同 error
//     （provider 暫停時返回 error）
//
// 每次成功嘅 Set 都會透過 best-effort（非阻塞）發送
// 將新值傳畀每個訂閱者 channel——滿咗嘅訂閱者 channel
// 會丟棄該值，所以要按消費者嘅速度設定緩衝大小。
//
// 呼叫取消訂閱函數會移除訂閱並關閉 channel
// （可以安全多次呼叫——只有第一次會關閉）。
func (v *ValueProvider[T]) Subscribe(buffer int) (<-chan T, func(), error)
```

### 暫停 / 恢復

```go
// Pause 停止 provider：暫停期間 Set 同 Subscribe 都會返回 error。
func (v *ValueProvider[T]) Pause()

// Resume 復活 provider。如果佢仍然存活，返回 error。
func (v *ValueProvider[T]) Resume() error
```

## 完整範例

```go
package main

import (
	"fmt"
	"time"

	"github.com/leoheung/go-patterns/container/subscribe"
)

func main() {
	// 建立一個裝住 int 嘅 provider，內部廣播緩衝為 8
	vp := subscribe.NewValueProvider(0, 8)

	// 訂閱，使用帶緩衝嘅 channel
	ch, unsubscribe, err := vp.Subscribe(4)
	if err != nil {
		panic(err)
	}
	defer unsubscribe()

	// 背景觀察者：每次變更都印出嚟
	go func() {
		for val := range ch {
			fmt.Println("變更:", val)
		}
	}()

	// 更新值——訂閱者會收到通知
	for i := 1; i <= 3; i++ {
		if err := vp.Set(i * 10); err != nil {
			panic(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	fmt.Println("當前值:", vp.Get()) // 30

	// 暫停：Set 同 Subscribe 都會返回 error
	vp.Pause()
	if err := vp.Set(99); err != nil {
		fmt.Println("暫停時 set:", err) // ValueProvider is paused
	}

	// Resume 復活 provider
	if err := vp.Resume(); err != nil {
		panic(err)
	}
	_ = vp.Set(40) // 再次通知
	time.Sleep(10 * time.Millisecond)
}
```

## 特性

- **可觀察值**：每次 `Set` 都會將新值廣播畀所有訂閱者
- **類型安全泛型**：包住任何值類型 `T`
- **安全取消訂閱**：`Subscribe` 返回取消訂閱閉包；佢會移除訂閱並保證只關閉 channel 一次
- **暫停 / 恢復**：底層用 `RenewableContext`——暫停會拒絕 `Set`/`Subscribe`，恢復就會復活 provider
- **並發安全**：`Get`/`Set` 由讀寫鎖保護
- **解耦廣播**：值經過內部帶緩衝嘅 channel（`buffer_size`），再由背景 worker 用 best-effort 發送扇出畀訂閱者——慢訂閱者會丟值而唔會互相阻塞。注意 `Set` 會寫入內部 channel，緩衝滿咗時可能會短暫阻塞；`buffer_size` 要按生產者嘅突發速率設定
