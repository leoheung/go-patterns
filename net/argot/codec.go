package argot

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
)

// 结果帧状态字节。它是整个协议中唯一的自我描述要素。
const (
	statusOK  byte = 0x00
	statusErr byte = 0x01
)

type funcEntry struct {
	name   string
	layout *FuncLayout
}

// Argot 运行时编解码器（Go 参考实现）。非 Go 实现只需按手册建出等价的双索引即可。
type Argot struct {
	man    *Manifest
	byName map[string]*FuncLayout
	byOp   map[uint16]funcEntry
	raw    []byte // 规范化后的手册字节，用于指纹
}

// LoadArgot 载入手册并建立双索引（名字 → 布局、opcode → 名字）。全程不涉及密钥。
func LoadArgot(manifestJSON []byte) (*Argot, error) {
	man, err := LoadManifest(manifestJSON)
	if err != nil {
		return nil, err
	}
	a := &Argot{
		man:    man,
		byName: make(map[string]*FuncLayout, len(man.Funcs)),
		byOp:   make(map[uint16]funcEntry, len(man.Funcs)),
	}
	for name, fl := range man.Funcs {
		if fl == nil {
			return nil, fmt.Errorf("argot: nil layout for function %q", name)
		}
		if _, dup := a.byOp[fl.Opcode]; dup {
			return nil, fmt.Errorf("argot: duplicate opcode %d in manifest", fl.Opcode)
		}
		a.byName[name] = fl
		a.byOp[fl.Opcode] = funcEntry{name: name, layout: fl}
	}
	raw, err := MarshalManifest(man)
	if err != nil {
		return nil, err
	}
	a.raw = raw
	return a, nil
}

// Manifest 返回底层手册（转发 / 持久化用）
func (a *Argot) Manifest() *Manifest { return a.man }

// Fingerprint 手册摘要，双端在握手之后互验
func (a *Argot) Fingerprint() string {
	sum := sha256.Sum256(a.raw)
	return hex.EncodeToString(sum[:])
}

// EncodeCall client 端：一次调用 → 线上字节。
// 线上格式：[opcode 2B][params 按布局顺序][decoys][padding]
func (a *Argot) EncodeCall(name string, args map[string]any) ([]byte, error) {
	fl, ok := a.byName[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownFunc, name)
	}
	if err := checkArgs(fl.Params, args); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writeOpcode(&buf, fl.Opcode)
	for _, sl := range fl.Params {
		if err := writeSlot(&buf, sl, args[sl.Ref]); err != nil {
			return nil, fmt.Errorf("argot: encoding %s.%s: %w", name, sl.Ref, err)
		}
	}
	for _, sl := range fl.Decoys {
		if err := writeSlot(&buf, sl, randomDecoyValue(sl.Type)); err != nil {
			return nil, fmt.Errorf("argot: encoding decoy: %w", err)
		}
	}
	padTo(&buf, a.man.PadPolicy.Bucket)
	return buf.Bytes(), nil
}

// DecodeCall server 端：线上字节 → 调用。
// 未知 opcode / 截断 / 帧内编码错误都是协议层错误，只在本地返回。
func (a *Argot) DecodeCall(data []byte) (string, map[string]any, error) {
	r := bytes.NewReader(data)
	op, err := readOpcode(r)
	if err != nil {
		return "", nil, err
	}
	entry, ok := a.byOp[op]
	if !ok {
		return "", nil, fmt.Errorf("%w: 0x%04x", ErrUnknownOpcode, op)
	}

	args := make(map[string]any, len(entry.layout.Params))
	for _, sl := range entry.layout.Params {
		v, err := readSlot(r, sl)
		if err != nil {
			return "", nil, fmt.Errorf("argot: decoding %s.%s: %w", entry.name, sl.Ref, err)
		}
		args[sl.Ref] = v
	}
	// 诱饵必须被解码才能推进偏移，随后直接丢弃。
	for _, sl := range entry.layout.Decoys {
		if _, err := readSlot(r, sl); err != nil {
			return "", nil, fmt.Errorf("argot: decoding decoy of %s: %w", entry.name, err)
		}
	}
	// 剩余字节是 padding，忽略。
	return entry.name, args, nil
}

// Result 调用应答：应用层错误与协议层错误分离
type Result struct {
	Err    string         // 应用错误文字；空串代表成功
	Values map[string]any // 依 Results schema 命名
}

// EncodeResult server 端：应答 → 线上字节。
// 成功：[opcode][status=0][results 槽位][padding]
// 失败：[opcode][status=1][len16 错误串][padding]
func (a *Argot) EncodeResult(name string, res Result) ([]byte, error) {
	fl, ok := a.byName[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrUnknownFunc, name)
	}

	var buf bytes.Buffer
	writeOpcode(&buf, fl.Opcode)

	if res.Err != "" {
		buf.WriteByte(statusErr)
		if err := writeSlot(&buf, SlotLayout{Type: ValString, Encoding: EncLen16}, res.Err); err != nil {
			return nil, fmt.Errorf("argot: encoding error text: %w", err)
		}
	} else {
		buf.WriteByte(statusOK)
		if err := checkArgs(fl.Results, res.Values); err != nil {
			return nil, err
		}
		for _, sl := range fl.Results {
			if err := writeSlot(&buf, sl, res.Values[sl.Ref]); err != nil {
				return nil, fmt.Errorf("argot: encoding result %s.%s: %w", name, sl.Ref, err)
			}
		}
	}
	padTo(&buf, a.man.PadPolicy.Bucket)
	return buf.Bytes(), nil
}

// DecodeResult client 端：应答 → Result。name 用于选定结果布局。
func (a *Argot) DecodeResult(data []byte, name string) (Result, error) {
	fl, ok := a.byName[name]
	if !ok {
		return Result{}, fmt.Errorf("%w: %q", ErrUnknownFunc, name)
	}
	r := bytes.NewReader(data)
	op, err := readOpcode(r)
	if err != nil {
		return Result{}, err
	}
	if op != fl.Opcode {
		return Result{}, fmt.Errorf("argot: result frame opcode 0x%04x does not match function %q", op, name)
	}
	status, err := r.ReadByte()
	if err != nil {
		return Result{}, fmt.Errorf("%w: status byte", ErrTruncated)
	}

	switch status {
	case statusOK:
		values := make(map[string]any, len(fl.Results))
		for _, sl := range fl.Results {
			v, err := readSlot(r, sl)
			if err != nil {
				return Result{}, fmt.Errorf("argot: decoding result %s.%s: %w", name, sl.Ref, err)
			}
			values[sl.Ref] = v
		}
		return Result{Values: values}, nil

	case statusErr:
		v, err := readSlot(r, SlotLayout{Type: ValString, Encoding: EncLen16})
		if err != nil {
			return Result{}, fmt.Errorf("argot: decoding error text: %w", err)
		}
		return Result{Err: v.(string)}, nil

	default:
		return Result{}, fmt.Errorf("argot: invalid status byte 0x%02x", status)
	}
}

// checkArgs 校验调用方提供的参数名集合与布局完全一致（缺参 / 多参都是协议层错误）
func checkArgs(slots []SlotLayout, args map[string]any) error {
	want := make(map[string]struct{}, len(slots))
	for _, sl := range slots {
		if _, ok := args[sl.Ref]; !ok {
			return fmt.Errorf("%w: missing %q", ErrArgMismatch, sl.Ref)
		}
		want[sl.Ref] = struct{}{}
	}
	for k := range args {
		if _, ok := want[k]; !ok {
			return fmt.Errorf("%w: unexpected %q", ErrArgMismatch, k)
		}
	}
	return nil
}

// padTo 在尾部补零到 bucket 的倍数；解码端忽略尾部剩余字节
func padTo(buf *bytes.Buffer, bucket int) {
	if bucket <= 1 {
		return
	}
	if rem := buf.Len() % bucket; rem != 0 {
		buf.Write(make([]byte, bucket-rem))
	}
}

// randomDecoyValue 生成诱饵槽位的随机内容（对端会丢弃）
func randomDecoyValue(t ValueType) any {
	switch t {
	case ValBytes:
		return randomBytes()
	case ValString:
		return hex.EncodeToString(randomBytes())
	case ValInt:
		return int64(rand.IntN(200001)) - 100000
	case ValUint:
		return uint64(rand.IntN(100000))
	case ValFloat:
		return rand.Float64() * 100
	case ValBool:
		return rand.IntN(2) == 1
	case ValJSON:
		return map[string]any{"v": rand.IntN(1000)}
	default:
		return nil
	}
}

func randomBytes() []byte {
	n := 4 + rand.IntN(17)
	b := make([]byte, n)
	for i := 0; i < n; i += 4 {
		v := rand.Uint32()
		for j := 0; j < 4 && i+j < n; j++ {
			b[i+j] = byte(v >> (8 * j))
		}
	}
	return b
}
