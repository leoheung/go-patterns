# 類型工具

## 概述

`utils` 套件提供 `IsInstanceOf`，一個用於檢查某個值是否為（或實現了）指定類型的泛型輔助函數。它簡化了 Go 中冗長的類型斷言語法，尤其適合處理 `any` 類型的值。

## 安裝

```go
import "github.com/leoheung/go-patterns/utils"
```

## API 參考

### `IsInstanceOf[T any](obj any) bool`

如果 `obj` 是類型 `T`（或當 `T` 為接口時實現了 `T`），返回 `true`，否則返回 `false`。要檢查的類型透過泛型參數指定。

```go
var obj any = "hello"

ok := utils.IsInstanceOf[string](obj) // true
ok = utils.IsInstanceOf[int](obj)     // false
```

## 完整範例

```go
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/leoheung/go-patterns/utils"
)

func describe(obj any) {
	switch {
	case utils.IsInstanceOf[string](obj):
		fmt.Println("這是字符串:", obj)
	case utils.IsInstanceOf[int](obj):
		fmt.Println("這是整數:", obj)
	case utils.IsInstanceOf[io.Reader](obj):
		fmt.Println("這是 io.Reader")
	default:
		fmt.Println("未知類型")
	}
}

func main() {
	describe("hello")        // 這是字符串: hello
	describe(42)             // 這是整數: 42
	describe(os.Stdin)       // 這是 io.Reader
}
```

## 特性

- **泛型類型參數**：直接用 `[T]` 傳入目標類型，無需反射樣板代碼
- **支持接口類型**：可用於接口類型（例如 `IsInstanceOf[io.Reader](obj)`）
- **不會 panic**：內部使用逗號-ok 形式的類型斷言，類型不匹配時絕不 panic

## 注意事項

- 等同於 `_, ok := obj.(T)`，只是類型透過泛型參數傳入
- 指針類型需明確指定（例如 `IsInstanceOf[*os.File](obj)`）
- `any` 值為 `nil` 時一律返回 `false`
