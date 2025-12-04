package utils

import (
	"fmt"
)

// 定义颜色枚举
type color int

const (
	Red color = iota
	Green
	BrightBlue
	Magenta // 紫色
	Cyan    // 青色
)

var idx color = Cyan

// 颜色映射表（ANSI 转义序列）
var colorCodes = map[color]string{
	Red:        "\033[31m",
	Green:      "\033[32m",
	Magenta:    "\033[35m",
	Cyan:       "\033[36m",
	BrightBlue: "\033[94m",
}

// PrintlnColor 输出彩色文字并换行
func PrintlnColor(c color, text string) {
	if code, ok := colorCodes[c]; ok {
		fmt.Println(code + text + "\033[0m")
	} else {
		fmt.Println(text) // 如果找不到颜色，直接普通输出
	}
}

func GetNextColor() color {
	if idx != Cyan {
		idx++
	} else {
		idx = Red
	}

	return idx
}
