package argot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strconv"
)

// ArgotVersion 手册与线上格式的版本
const ArgotVersion uint32 = 1

// DefaultBucket 默认填充分桶大小（字节）
const DefaultBucket = 64

// PadPolicy 填充与诱饵策略。它记录编译期的选择参数，
// 而每个函数实际注入的诱饵槽位数量已固定在 manifest 的 decoys 中。
type PadPolicy struct {
	DecoyMin int `json:"decoy_min"` // 每函数诱饵数量的下界
	DecoyMax int `json:"decoy_max"` // 每函数诱饵数量的上界
	Bucket   int `json:"bucket"`    // 消息长度向上取整的倍数
}

// DefaultPadPolicy 默认策略
var DefaultPadPolicy = PadPolicy{DecoyMin: 0, DecoyMax: 2, Bucket: DefaultBucket}

// Manifest 黑话手册：唯一的分发物，跨语言契约
type Manifest struct {
	ArgotVersion uint32                 `json:"argot_version"`
	ProtoName    string                 `json:"proto_name"`
	ProtoVersion uint32                 `json:"proto_version"`
	PadPolicy    PadPolicy              `json:"pad_policy"`
	Funcs        map[string]*FuncLayout `json:"funcs"`
}

// FuncLayout 单个函数的混淆布局
type FuncLayout struct {
	Opcode  uint16       `json:"opcode"`  // 线上 opcode（keyed shuffle 产物）
	Params  []SlotLayout `json:"params"`  // 已按线上顺序排列
	Results []SlotLayout `json:"results"` // 已按线上顺序排列
	Decoys  []SlotLayout `json:"decoys"`  // 编译期注入的假参数，解码时丢弃
}

// SlotLayout 单个槽位的线上编码规则（非 Go 实现照此复刻）
type SlotLayout struct {
	Ref      string    `json:"ref"`      // 对应 spec 参数名（诱饵为生成名）
	Type     ValueType `json:"type"`     // 值类型
	Encoding WireEnc   `json:"encoding"` // 编码变体
}

// Compiler 确定性混淆编译器。只在云端运行，密钥永不外泄。
type Compiler struct {
	key []byte
}

// NewCompiler 以混淆主种子创建编译器
func NewCompiler(key []byte) *Compiler {
	return &Compiler{key: key}
}

// Compile 把明文协议编译为手册。
// 确定性：同一 (key, spec) 每次产出完全一致的 manifest。
func (c *Compiler) Compile(p *Protocol) (*Manifest, error) {
	if c == nil || len(c.key) == 0 {
		return nil, errors.New("argot: empty obfuscation key")
	}
	if err := p.validate(); err != nil {
		return nil, err
	}
	if DefaultPadPolicy.DecoyMax < DefaultPadPolicy.DecoyMin {
		return nil, fmt.Errorf("argot: invalid pad policy %+v", DefaultPadPolicy)
	}

	rng := newKeyedRand(c.key)

	names := make([]string, 0, len(p.Funcs))
	for name := range p.Funcs {
		names = append(names, name)
	}
	sort.Strings(names) // 遍历顺序稳定，迭代顺序永不渗入产物

	// opcode：对 0..65535 池做 keyed shuffle，按名字顺序取前 N 个。
	pool := make([]uint16, maxFuncs)
	for i := range pool {
		pool[i] = uint16(i)
	}
	rng.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})

	man := &Manifest{
		ArgotVersion: ArgotVersion,
		ProtoName:    p.Name,
		ProtoVersion: p.Version,
		PadPolicy:    DefaultPadPolicy,
		Funcs:        make(map[string]*FuncLayout, len(names)),
	}
	for i, name := range names {
		fn := p.Funcs[name]
		man.Funcs[name] = &FuncLayout{
			Opcode:  pool[i],
			Params:  layoutSlots(rng, fn.Params),
			Results: layoutSlots(rng, fn.Results),
			Decoys:  decoySlots(rng, man.PadPolicy),
		}
	}
	return man, nil
}

// layoutSlots 打乱参数顺序，并为每个槽位挑选编码变体
func layoutSlots(rng *rand.Rand, params []Param) []SlotLayout {
	out := make([]SlotLayout, 0, len(params))
	for _, idx := range rng.Perm(len(params)) {
		p := params[idx]
		out = append(out, SlotLayout{
			Ref:      p.Name,
			Type:     p.Type,
			Encoding: pickEncoding(rng, p.Type),
		})
	}
	return out
}

// pickEncoding 按类型在适用变体中挑选一个
func pickEncoding(rng *rand.Rand, t ValueType) WireEnc {
	var candidates []WireEnc
	switch t {
	case ValString:
		candidates = []WireEnc{EncLen16}
	case ValBytes:
		candidates = []WireEnc{EncLen16, EncLen32}
	case ValInt:
		candidates = []WireEnc{EncVarintZigzag, EncFixed64}
	case ValUint:
		candidates = []WireEnc{EncVarint, EncFixed64}
	case ValFloat:
		candidates = []WireEnc{EncFixed64}
	case ValBool:
		candidates = []WireEnc{EncFixed8}
	case ValJSON:
		candidates = []WireEnc{EncLen32}
	}
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	return candidates[rng.IntN(len(candidates))]
}

// decoySlots 为单个函数生成固定数量的诱饵槽位。
// 数量在编译期定死并写入手册，解码端据此知道要读几个槽位，线上不携带数量字段。
func decoySlots(rng *rand.Rand, policy PadPolicy) []SlotLayout {
	span := policy.DecoyMax - policy.DecoyMin + 1
	n := policy.DecoyMin
	if span > 0 {
		n += rng.IntN(span)
	}
	out := make([]SlotLayout, 0, n)
	types := []ValueType{ValBytes, ValString, ValInt, ValUint, ValBool, ValJSON}
	for i := 0; i < n; i++ {
		t := types[rng.IntN(len(types))]
		out = append(out, SlotLayout{
			Ref:      "d" + strconv.Itoa(i),
			Type:     t,
			Encoding: pickEncoding(rng, t),
		})
	}
	return out
}

// newKeyedRand 由密钥派生确定性 PRNG（ChaCha8，跨平台一致）
func newKeyedRand(key []byte) *rand.Rand {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("argot/v1/seed"))
	var seed [32]byte
	copy(seed[:], mac.Sum(nil))
	return rand.New(rand.NewChaCha8(seed))
}

// MarshalManifest 序列化为语言无关的 JSON 手册（map key 自动排序，输出规范化）
func MarshalManifest(m *Manifest) ([]byte, error) {
	if m == nil {
		return nil, errors.New("argot: nil manifest")
	}
	return json.MarshalIndent(m, "", "  ")
}

// LoadManifest 从 JSON 手册反序列化，并规范化填充策略的非法取值
func LoadManifest(data []byte) (*Manifest, error) {
	m := new(Manifest)
	if err := json.Unmarshal(data, m); err != nil {
		return nil, fmt.Errorf("argot: invalid manifest: %w", err)
	}
	if len(m.Funcs) == 0 {
		return nil, errors.New("argot: manifest has no functions")
	}
	m.PadPolicy = normalizePadPolicy(m.PadPolicy)
	return m, nil
}

func normalizePadPolicy(p PadPolicy) PadPolicy {
	if p.DecoyMin < 0 {
		p.DecoyMin = 0
	}
	if p.DecoyMax < p.DecoyMin {
		p.DecoyMax = p.DecoyMin
	}
	if p.Bucket <= 0 {
		p.Bucket = DefaultBucket
	}
	return p
}

// Fingerprint 手册摘要（规范化序列化后的 sha256），双端握手后互验
func (m *Manifest) Fingerprint() string {
	raw, err := MarshalManifest(m)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
