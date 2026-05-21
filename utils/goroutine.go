package utils

import (
	"runtime"
	"strconv"
	"strings"
)

// GetGoid 获取当前 Goroutine ID
// 注意：Go 不鼓励使用 goroutine ID，仅用于调试
func GetGoid() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)

	// 格式: "goroutine 123 [running]:"
	s := string(buf[:n])

	// 更健壮的解析
	if !strings.HasPrefix(s, "goroutine ") {
		return 0
	}

	fields := strings.Fields(s)
	if len(fields) < 2 {
		return 0
	}

	id, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}

	return id
}
