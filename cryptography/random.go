package cryptography

import (
	"crypto/rand"
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
