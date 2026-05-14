# 彩色輸出

## 概述

`utils` 包提供了使用 ANSI 轉義序列的終端彩色輸出功能。

## 顏色類型

```go
const (
    Red        // 紅色
    Green      // 綠色
    BrightBlue // 亮藍色
    Magenta    // 紫色
    Cyan       // 青色
)
```

## API 參考

### `PrintlnColor(c color, text string)`

以指定顏色打印文本並添加換行符。如果找不到顏色，則回退到普通文本。

```go
import "github.com/leoheung/go-patterns/utils"

utils.PrintlnColor(utils.Red, "這是紅色文本")
utils.PrintlnColor(utils.Green, "這是綠色文本")
utils.PrintlnColor(utils.BrightBlue, "這是藍色文本")
utils.PrintlnColor(utils.Magenta, "這是紫色文本")
utils.PrintlnColor(utils.Cyan, "這是青色文本")
```

### `GetNextColor() color`

返回輪換中的下一個顏色（Red → Green → BrightBlue → Magenta → Cyan → Red...）。

```go
c1 := utils.GetNextColor() // Red
c2 := utils.GetNextColor() // Green
c3 := utils.GetNextColor() // BrightBlue
```

## 示例

```go
package main

import (
	"fmt"
	"github.com/leoheung/go-patterns/utils"
)

func main() {
	colors := []utils.color{
		utils.Red,
		utils.Green,
		utils.BrightBlue,
		utils.Magenta,
		utils.Cyan,
	}

	messages := []string{
		"錯誤消息",
		"成功消息",
		"信息消息",
		"警告消息",
		"調試消息",
	}

	for i, msg := range messages {
		utils.PrintlnColor(colors[i], msg)
	}

	fmt.Println("\n使用 GetNextColor():")
	for i := 0; i < 3; i++ {
		c := utils.GetNextColor()
		utils.PrintlnColor(c, fmt.Sprintf("消息 %d", i+1))
	}
}
```

## 注意事項

- 顏色使用 ANSI 轉義序列呈現
- 並非所有終端都支持所有顏色
- 每行後自動重置顏色
- `GetNextColor()` 循環遍歷顏色，便於交替模式