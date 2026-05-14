# Channel 工具

## 概述

`utils` 包提供了非阻塞和基於超時的通道操作，用於更安全的並發編程。

## 功能特點

- **非阻塞 Enqueue/Dequeue**：發送和接收而不阻塞
- **基於超時的操作**：為通道操作添加時間限制
- **泛型支持**：適用於任何數據類型

## API 參考

### `TryEnqueue[T any](c chan<- T, data T) bool`

嘗試非阻塞地發送數據到通道。如果成功返回 `true`，如果通道已滿或阻塞返回 `false`。

```go
ch := make(chan int, 5)

ok := utils.TryEnqueue(ch, 42)
if ok {
    fmt.Println("入隊成功")
}
```

### `TryDequeue[T any](c <-chan T) (*T, bool)`

嘗試非阻塞地從通道接收數據。如果成功返回數據和 `true`，如果通道為空返回 `nil` 和 `false`。

```go
ch := make(chan int, 5)
ch <- 10

val, ok := utils.TryDequeue(ch)
if ok {
    fmt.Printf("出隊: %d\n", *val)
}
```

### `EnqueueWithTimeout[T any](c chan<- T, data T, timeout time.Duration) bool`

嘗試在超時時間內發送數據到通道。如果成功返回 `true`，如果超時則返回 `false`。

```go
ch := make(chan int, 2)

ok := utils.EnqueueWithTimeout(ch, 42, 5*time.Second)
if ok {
    fmt.Println("在超時前入隊成功")
} else {
    fmt.Println("入隊超時")
}
```

### `DequeueWithTimeout[T any](c <-chan T, timeout time.Duration) (*T, bool)`

嘗試在超時時間內從通道接收數據。如果成功返回數據和 `true`，如果超時則返回 `nil` 和 `false`。

```go
ch := make(chan int, 5)

val, ok := utils.DequeueWithTimeout(ch, 5*time.Second)
if ok {
    fmt.Printf("在超時前出隊: %d\n", *val)
} else {
    fmt.Println("出隊超時")
}
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
	ch := make(chan string, 3)

	// 非阻塞入隊
	for _, msg := range []string{"a", "b", "c"} {
		if ok := utils.TryEnqueue(ch, msg); ok {
			fmt.Printf("入隊: %s\n", msg)
		} else {
			fmt.Printf("入隊失敗: %s\n", msg)
		}
	}

	// 非阻塞出隊
	for i := 0; i < 4; i++ {
		if val, ok := utils.TryDequeue(ch); ok {
			fmt.Printf("出隊: %s\n", *val)
		} else {
			fmt.Println("通道為空")
		}
	}

	// 基於超時的操作
	largeCh := make(chan int, 1)
	largeCh <- 1

	// 由於通道已滿，這個會超時
	ok := utils.EnqueueWithTimeout(largeCh, 2, 100*time.Millisecond)
	fmt.Printf("帶超時入隊: %v\n", ok)
}
```

## 注意事項

- `TryEnqueue` 和 `TryDequeue` 是非阻塞的，會立即返回
- 超時函數使用 `time.Timer` 進行高效的計時
- 所有函數都是線程安全的
- 泛型類型參數 `[T any]` 適用於任何數據類型