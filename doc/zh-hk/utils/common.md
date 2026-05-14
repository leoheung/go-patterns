# 通用輔助函數

## 概述

`utils` 包提供常見的輔助函數，包括類型檢查、字符串處理和 goroutine 控制。

## API 參考

### `IsNil[T any](v T) bool`

檢查值是否為 nil。支持接口、指針、切片、映射、通道、函數和 chan 類型。

```go
var s *string = nil
if utils.IsNil(s) {
    fmt.Println("值為 nil")
}

// 使用泛型
var arr []int = nil
utils.IsNil(arr) // true
```

### `IsDigits(s string) bool`

檢查字符串是否完全由數字（0-9）組成。

```go
utils.IsDigits("12345") // true
utils.IsDigits("12a45") // false
utils.IsDigits("")      // true (空字符串)
```

### `DelayDo(d time.Duration, fn func())`

延遲後執行函數。這是一個阻塞操作。

```go
utils.DelayDo(500*time.Millisecond, func() {
    fmt.Println("延遲執行")
})
```

### `PPrint(obj interface{})`

以格式化方式打印對象。是 `PPrettyPrint` 的別名。

```go
utils.PPrint(myStruct)
```

### `Hold()`

無限期阻塞當前 goroutine。用於調試或保持程序運行。

```go
utils.Hold() // 永久阻塞
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
	// 檢查 nil 值
	var ptr *int = nil
	fmt.Printf("ptr 為 nil: %v\n", utils.IsNil(ptr))

	var slice []string = nil
	fmt.Printf("slice 為 nil: %v\n", utils.IsNil(slice))

	// 檢查字符串是否全為數字
	fmt.Printf("\"123\" 全為數字: %v\n", utils.IsDigits("123"))
	fmt.Printf("\"12a\" 全為數字: %v\n", utils.IsDigits("12a"))

	// 延遲執行
	fmt.Println("開始延遲任務...")
	utils.DelayDo(1*time.Second, func() {
		fmt.Println("延遲任務已執行！")
	})
	fmt.Println("延遲之後")
}
```

## 注意事項

- `IsNil` 使用反射並正確處理各種類型
- `IsDigits` 對空字符串返回 `true`
- `DelayDo` 阻塞直到延遲完成
- `Hold()` 故意創建死鎖，僅用於調試