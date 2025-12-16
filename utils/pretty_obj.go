package utils

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/davecgh/go-spew/spew"
)

// PPrettyPrint 使用 spew 打印任意对象（更接近 Python pprint）
func PPrettyPrint(v any) {
	cfg := spew.ConfigState{
		Indent:                  "  ",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
		SpewKeys:                true,
	}
	cfg.Dump(v)
}

// PrettyObjStr 返回对象的漂亮字符串表示（不打印，供日志/前端返回）
func PrettyObjStr(v any) string {
	cfg := spew.ConfigState{
		Indent:                  "  ",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
		SpewKeys:                true,
	}
	return cfg.Sdump(v)
}

// JSONalizeStr 尝试将任意对象编码为漂亮的 JSON 字符串；若 JSON 编码失败，则回退到 PrettyObjStr。
func JSONalizeStr(v any) string {
	if b, err := json.MarshalIndent(v, "", "  "); err == nil {
		return string(b)
	}
	return PrettyObjStr(v)
}

// DeJSONalizeStr 将 JSON 字符串解码到 v 指向的值中。
// v 必须是非 nil 的指针，否则返回 error。
func DeJSONalizeStr(s string, v any) error {
	if v == nil {
		return fmt.Errorf("target is nil")
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	return json.Unmarshal([]byte(s), v)
}
