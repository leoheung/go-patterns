package net

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-chi/chi/v5"
)

func PrintCHIRoutes(r *chi.Mux) {
	err := chi.Walk(r, func(method string, route string, h http.Handler, mws ...func(http.Handler) http.Handler) error {
		hName := handlerName(h)
		var mwNames []string
		for _, mw := range mws {
			mwNames = append(mwNames, funcName(mw))
		}
		if len(mwNames) > 0 {
			fmt.Printf("%-6s %-60s -> %s | mws: %s\n", method, route, hName, strings.Join(mwNames, ", "))
		} else {
			fmt.Printf("%-6s %-60s -> %s\n", method, route, hName)
		}
		return nil
	})
	if err != nil {
		fmt.Println("route walk error:", err)
	}
}

func handlerName(h http.Handler) string {
	// 最终 handler 可能是 chi 的链式包装，这里打印可解析到的函数/类型名
	v := reflect.ValueOf(h)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// 优先尝试 http.HandlerFunc
	if hf, ok := h.(http.HandlerFunc); ok {
		return trimPkg(runtime.FuncForPC(reflect.ValueOf(hf).Pointer()).Name())
	}
	// 一般情况回退到类型或可执行入口名
	if v.IsValid() {
		if v.Kind() == reflect.Struct {
			return v.Type().String()
		}
	}
	// 直接用函数名（可能是闭包）
	if fn := runtime.FuncForPC(reflect.ValueOf(h).Pointer()); fn != nil {
		return trimPkg(fn.Name())
	}
	return "unknownHandler"
}

func funcName(fn interface{}) string {
	val := reflect.ValueOf(fn)
	if val.Kind() != reflect.Func {
		return "unknownMiddleware"
	}
	if f := runtime.FuncForPC(val.Pointer()); f != nil {
		return trimPkg(f.Name())
	}
	return "unknownMiddleware"
}

func trimPkg(name string) string {
	// 去掉长包路径，只保留最后段，便于阅读
	if i := strings.LastIndex(name, "/"); i >= 0 && i < len(name)-1 {
		name = name[i+1:]
	}
	// 将包名.函数 拆成最后一段
	if i := strings.LastIndex(name, "."); i >= 0 && i < len(name)-1 {
		return name[i+1:]
	}
	return name
}
