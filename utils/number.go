package utils

import (
    "math"
    "strconv"
)

// Number 模拟 TypeScript 的 number 类型，既可以表示整数也可以表示浮点数。
// 底层使用 float64 存储，这与 JS/TS 的实现机制一致。
type Number struct {
    value float64
}

// ParseNumber 将字符串解析为 *Number。
// 它使用 strconv.ParseFloat，因此可以处理 "100" 和 "100.5" 两种情况。
func ParseNumber(raw string) (*Number, error) {
    f, err := strconv.ParseFloat(raw, 64)
    if err != nil {
        return nil, err
    }
    return &Number{value: f}, nil
}

// Float 返回底层的 float64 值
func (n *Number) Float() float64 {
    return n.value
}

// Int 返回数值的整数部分 (直接截断小数)
func (n *Number) Int() int {
    return int(n.value)
}

// Int64 返回 int64 类型的整数部分
func (n *Number) Int64() int64 {
    return int64(n.value)
}

// IsInteger 判断该数值是否为一个整数（即没有小数部分）
// 例如 10.0 会返回 true, 10.5 会返回 false
func (n *Number) IsInteger() bool {
    return n.value == math.Trunc(n.value)
}