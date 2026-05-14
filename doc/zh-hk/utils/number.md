# 數字工具

## 概述

`utils` 包提供了一個 `Number` 類型，模擬 TypeScript/JavaScript 的 `number` 類型，使用 `float64` 內部存儲來統一處理整數和浮點數。

## Number 類型

```go
type Number struct {
    value float64
}
```

## API 參考

### `ParseNumber(raw string) (*Number, error)`

將字符串解析為 `*Number`。處理整數（"100"）和浮點數（"100.5"）格式。

```go
num, err := utils.ParseNumber("42")
if err != nil {
    log.Fatal(err)
}
fmt.Println(num.Float()) // 42
```

### `(n *Number) Float() float64`

返回底層的 `float64` 值。

```go
num, _ := utils.ParseNumber("123.456")
fmt.Println(num.Float()) // 123.456
```

### `(n *Number) Int() int`

返回數值的整數部分（直接截斷小數部分）。

```go
num, _ := utils.ParseNumber("99.9")
fmt.Println(num.Int()) // 99
```

### `(n *Number) Int64() int64`

以 `int64` 形式返回整數部分。

```go
num, _ := utils.ParseNumber("9999999999.5")
fmt.Println(num.Int64()) // 9999999999
```

### `(n *Number) IsInteger() bool`

判斷該數值是否為整數（即沒有小數部分）。

```go
num1, _ := utils.ParseNumber("10.0")
num2, _ := utils.ParseNumber("10.5")

fmt.Println(num1.IsInteger()) // true
fmt.Println(num2.IsInteger()) // false
```

## 示例

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// 解析各種數字格式
	examples := []string{"42", "3.14", "100.0", "99.99", "-5"}

	for _, s := range examples {
		num, err := utils.ParseNumber(s)
		if err != nil {
			fmt.Printf("解析 %s 出錯: %v\n", s, err)
			continue
		}

		fmt.Printf("輸入: %s\n", s)
		fmt.Printf("  Float: %f\n", num.Float())
		fmt.Printf("  Int: %d\n", num.Int())
		fmt.Printf("  Int64: %d\n", num.Int64())
		fmt.Printf("  IsInteger: %v\n", num.IsInteger())
		fmt.Println()
	}
}
```

## 注意事項

- 內部使用 `float64` 以匹配 JavaScript/TypeScript 的行為
- 整數轉換會截斷小數部分（不四捨五入）
- `IsInteger()` 使用 `math.Trunc()` 進行精度比較
- 字符串解析使用 64 位精度的 `strconv.ParseFloat`