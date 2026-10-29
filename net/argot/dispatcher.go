package argot

import "fmt"

// Handler 服务端函数实现
type Handler func(args map[string]any) (map[string]any, error)

// Dispatcher 服务端分发胶水：解码 → 分派 → 编码。
// 它本身不含任何混淆逻辑，仅免去调用方手写这层样板代码。
type Dispatcher struct {
	a *Argot
	h map[string]Handler
}

// NewDispatcher 基于已载入的手册创建分发器
func NewDispatcher(a *Argot) *Dispatcher {
	return &Dispatcher{a: a, h: make(map[string]Handler)}
}

// Register 注册函数实现（不存在于手册中的名字也可注册，但永远不会被调用）
func (d *Dispatcher) Register(name string, h Handler) {
	d.h[name] = h
}

// Handle 处理一帧调用，返回应答帧。
// 协议层错误（无法解码）在本地返回；handler 的应用错误写入应答帧的 Result.Err。
func (d *Dispatcher) Handle(payload []byte) ([]byte, error) {
	if d == nil || d.a == nil {
		return nil, fmt.Errorf("argot: dispatcher not initialized")
	}
	name, args, err := d.a.DecodeCall(payload)
	if err != nil {
		return nil, err
	}

	h, ok := d.h[name]
	if !ok {
		return nil, fmt.Errorf("%w: no handler for %q", ErrUnknownFunc, name)
	}

	values, appErr := h(args)
	res := Result{Values: values}
	if appErr != nil {
		res.Err = appErr.Error()
	}
	return d.a.EncodeResult(name, res)
}
