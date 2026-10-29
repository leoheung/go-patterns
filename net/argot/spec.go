/*
Package argot: 协议混淆（黑话）模式

简介:
argot 把一份明文协议定义（函数调用 + 参数 schema）编译成一份「手册（manifest）」。
手册中，所有自我描述的要素（函数名、字段名、字段顺序、类型标记、稳定 opcode）
都被替换成由共享密钥派生出的值。持有同一份手册的 client 与 server 可以照常互相
调用函数；未持有手册的观察者只能看到一串无法归因的字节。

核心功能:
1. Compiler.Compile: 确定性混淆编译（同 key + 同 spec ⇒ 产物完全一致）。
2. Manifest: 语言无关的交换格式（JSON），跨语言实现的唯一契约。
3. Argot: 运行时编解码器，按手册完成 EncodeCall / DecodeCall / EncodeResult / DecodeResult。
4. Dispatcher: 可选的服务端胶水（解码 → 分派 → 编码）。

使用场景:
  - 三方结构：云端 Argot Service 持有明文 spec 与混淆密钥，编译并签名分发手册；
    私密 client 与私密 server 只持有手册，在 Noise 等加密传输之上通信。
  - 需要在加密之外进一步消除应用层协议指纹（反 DPI / 反被动指纹）的场合。

安全性:
  - argot 不是加密方案。机密性、认证与完整性由 Noise 等传输层负责；
    argot 只负责移除线上的协议语义（无函数名、无字段名、无类型标记、无 magic）。
  - 手册本身就是解码字典：必须经加密通道分发并验签，泄露即混淆归零。
  - 编译器只在云端运行；双端永不接触混淆密钥，只持有派生出的手册。

Usage Example:
// 1. 云端：混淆协议并签名分发
man, _ := argot.NewCompiler(key).Compile(proto)
payload, _ := argot.MarshalManifest(man)
sig := ed25519.Sign(signingKey, payload)

// 2. 双端：验签后载入手册
a, _ := argot.LoadArgot(payload)

// 3. client 编码 / server 解码
frame, _ := a.EncodeCall("fs.read", map[string]any{"path": "/etc/hosts", "offset": 0})
name, args, _ := a.DecodeCall(frame)
*/
package argot

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ValueType 参数值类型
type ValueType int

const (
	// ValString UTF-8 字符串
	ValString ValueType = iota
	// ValBytes 原始字节
	ValBytes
	// ValInt int64
	ValInt
	// ValUint uint64
	ValUint
	// ValFloat float64
	ValFloat
	// ValBool bool
	ValBool
	// ValJSON 任意 JSON 值（以 any 透传）
	ValJSON
)

// String 返回类型在手册中的稳定名称
func (t ValueType) String() string {
	switch t {
	case ValString:
		return "string"
	case ValBytes:
		return "bytes"
	case ValInt:
		return "int"
	case ValUint:
		return "uint"
	case ValFloat:
		return "float"
	case ValBool:
		return "bool"
	case ValJSON:
		return "json"
	default:
		return fmt.Sprintf("unknown(%d)", int(t))
	}
}

// MarshalJSON 序列化为手册中的稳定名称
func (t ValueType) MarshalJSON() ([]byte, error) {
	if !t.valid() {
		return nil, fmt.Errorf("argot: invalid value type %d", int(t))
	}
	return json.Marshal(t.String())
}

// UnmarshalJSON 从手册中的稳定名称反序列化
func (t *ValueType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := valueTypeFromString(s)
	if err != nil {
		return err
	}
	*t = v
	return nil
}

func valueTypeFromString(s string) (ValueType, error) {
	switch s {
	case "string":
		return ValString, nil
	case "bytes":
		return ValBytes, nil
	case "int":
		return ValInt, nil
	case "uint":
		return ValUint, nil
	case "float":
		return ValFloat, nil
	case "bool":
		return ValBool, nil
	case "json":
		return ValJSON, nil
	default:
		return 0, fmt.Errorf("argot: unknown value type %q", s)
	}
}

func (t ValueType) valid() bool { return t >= ValString && t <= ValJSON }

// 协议层错误哨兵。协议层错误只在本地返回，永不上线。
var (
	ErrUnknownFunc   = errors.New("argot: unknown function")
	ErrUnknownOpcode = errors.New("argot: unknown opcode")
	ErrTruncated     = errors.New("argot: truncated frame")
	ErrArgMismatch   = errors.New("argot: argument mismatch")
)

// Protocol 明文协议定义（协议作者撰写，本身可公开）
type Protocol struct {
	Name    string           // 协议名（仅元信息，永不上线）
	Version uint32           // 双端必须一致
	Funcs   map[string]*Func // 函数名 → 定义；key 即身份，天然防重名
}

// Func 单个函数的明文定义
type Func struct {
	Params  []Param // 入参 schema；声明顺序仅用于阅读，线上顺序由编译决定
	Results []Param // 返回值 schema
}

// Param 单个参数的明文定义
type Param struct {
	Name string // 仅编译期定位用，永不上线
	Type ValueType
}

// maxFuncs opcode 空间上限（uint16）
const maxFuncs = 1 << 16

func (p *Protocol) validate() error {
	if p == nil {
		return errors.New("argot: nil protocol")
	}
	if len(p.Funcs) == 0 {
		return errors.New("argot: protocol has no functions")
	}
	if len(p.Funcs) > maxFuncs {
		return fmt.Errorf("argot: too many functions (%d > %d)", len(p.Funcs), maxFuncs)
	}
	for name, fn := range p.Funcs {
		if name == "" {
			return errors.New("argot: empty function name")
		}
		if fn == nil {
			return fmt.Errorf("argot: nil definition for function %q", name)
		}
		if err := validateParams(name, "param", fn.Params); err != nil {
			return err
		}
		if err := validateParams(name, "result", fn.Results); err != nil {
			return err
		}
	}
	return nil
}

func validateParams(fn, kind string, params []Param) error {
	seen := make(map[string]struct{}, len(params))
	for _, p := range params {
		if p.Name == "" {
			return fmt.Errorf("argot: %s: empty %s name", fn, kind)
		}
		if !p.Type.valid() {
			return fmt.Errorf("argot: %s: %s %q has invalid type %d", fn, kind, p.Name, int(p.Type))
		}
		if _, dup := seen[p.Name]; dup {
			return fmt.Errorf("argot: %s: duplicate %s name %q", fn, kind, p.Name)
		}
		seen[p.Name] = struct{}{}
	}
	return nil
}
