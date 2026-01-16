package cryptography

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandString 返回长度为 n 的随机字符串（字符集 a-zA-Z0-9）。
// 当 n <= 0 时返回空字符串。若底层随机源出错会返回 error。
func RandString(n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, n)
	base := big.NewInt(int64(len(letters)))
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, base)
		if err != nil {
			return ""
		}
		out[i] = letters[num.Int64()]
	}
	return string(out)
}

// RandUUID 生成一个符合 RFC 4122 标准的 Version 4 UUID。
// 返回的格式为 "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"。
// 如果生成随机数失败，返回空字符串。
func RandUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return ""
	}

	// 设置版本号 (Version 4)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// 设置变体 (RFC 4122 Variant)
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
