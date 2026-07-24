package utils

// 通用版本
func ModEuclid(num, base int) int {
	r := num % base
	if r < 0 {
		r += base
	}
	return r
}
