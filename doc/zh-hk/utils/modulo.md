# 取模工具

## 概述

`utils` 套件提供 `ModEuclid`，一個歐幾里得取模（Euclidean modulo）運算。Go 內建的 `%` 運算符是截斷餘數（結果跟隨被除數的正負號），而 `ModEuclid` 在 `base` 為正數時一律返回非負結果。適用於環形索引、哈希分桶、負數循環包裝等場景。

## 安裝

```go
import "github.com/leoheung/go-patterns/utils"
```

## API 參考

### `ModEuclid(num, base int) int`

返回 `num` 對 `base` 的歐幾里得取模。如果內建餘數為負數，會加上 `base` 使結果非負（當 `base` 為正數時）。

```go
r := utils.ModEuclid(-1, 5) // 4  （內建的 -1 % 5 會是 -1）
r = utils.ModEuclid(7, 5)   // 2
```

## 完整範例

```go
package main

import (
	"fmt"

	"github.com/leoheung/go-patterns/utils"
)

func main() {
	// 將可能為負的索引包裝成有效範圍
	items := []string{"a", "b", "c", "d", "e"}
	for _, i := range []int{-1, -6, 3, 8} {
		idx := utils.ModEuclid(i, len(items))
		fmt.Printf("items[ModEuclid(%d, 5)] = %s\n", i, items[idx])
	}
	// 輸出：
	// items[ModEuclid(-1, 5)] = e
	// items[ModEuclid(-6, 5)] = e
	// items[ModEuclid(3, 5)] = d
	// items[ModEuclid(8, 5)] = d

	// 同內建運算符比較
	fmt.Println(-1 % 5)                 // -1
	fmt.Println(utils.ModEuclid(-1, 5)) // 4
}
```

## 特性

- **結果恆為非負**：當 `base` 為正數時，結果必定落在 `[0, base)` 範圍內
- **僅多一個分支**：只在內建餘數為負時加上 `base`，額外開銷極小
- **環形安全**：非常適合跨零點包裝數組索引或循環計數器

## 注意事項

- `base` 不可為 `0`（與內建 `%` 運算符的限制相同）
- 當 `num` 和 `base` 均為正數時，`ModEuclid` 的行為與 `num % base` 完全一致
- 結果遵循歐幾里得定義：當 `base > 0` 時 `0 <= 結果 < base`
