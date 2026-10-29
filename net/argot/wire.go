package argot

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// WireEnc 槽位在线上使用的编码变体（手册中的稳定标识符）
type WireEnc string

const (
	EncVarint       WireEnc = "varint"        // 无号 LEB128
	EncVarintZigzag WireEnc = "varint-zigzag" // zigzag 后无号 LEB128
	EncFixed64      WireEnc = "fixed64"       // 8 字节大端
	EncFixed8       WireEnc = "fixed8"        // 1 字节布尔
	EncLen16        WireEnc = "len16-prefixed"
	EncLen32        WireEnc = "len32-prefixed"
)

func (e WireEnc) valid() bool {
	switch e {
	case EncVarint, EncVarintZigzag, EncFixed64, EncFixed8, EncLen16, EncLen32:
		return true
	}
	return false
}

func writeOpcode(w *bytes.Buffer, op uint16) {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], op)
	w.Write(b[:])
}

func readOpcode(r *bytes.Reader) (uint16, error) {
	var b [2]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, fmt.Errorf("%w: opcode", ErrTruncated)
	}
	return binary.BigEndian.Uint16(b[:]), nil
}

// writeSlot 按槽位布局写入单个值
func writeSlot(buf *bytes.Buffer, sl SlotLayout, v any) error {
	switch sl.Encoding {
	case EncVarint:
		n, err := asUint64(v)
		if err != nil {
			return err
		}
		buf.Write(binary.AppendUvarint(nil, n))

	case EncVarintZigzag:
		n, err := asInt64(v)
		if err != nil {
			return err
		}
		buf.Write(binary.AppendUvarint(nil, uint64(n<<1)^uint64(n>>63)))

	case EncFixed64:
		var b [8]byte
		switch sl.Type {
		case ValInt:
			n, err := asInt64(v)
			if err != nil {
				return err
			}
			binary.BigEndian.PutUint64(b[:], uint64(n))
		case ValUint:
			n, err := asUint64(v)
			if err != nil {
				return err
			}
			binary.BigEndian.PutUint64(b[:], n)
		default:
			f, err := asFloat64(v)
			if err != nil {
				return err
			}
			binary.BigEndian.PutUint64(b[:], math.Float64bits(f))
		}
		buf.Write(b[:])

	case EncFixed8:
		b, err := asBool(v)
		if err != nil {
			return err
		}
		if b {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}

	case EncLen16, EncLen32:
		p, err := asPayload(sl.Type, v)
		if err != nil {
			return err
		}
		width := 4
		if sl.Encoding == EncLen16 {
			width = 2
		}
		if width == 2 && len(p) > math.MaxUint16 {
			return fmt.Errorf("argot: payload too long for len16 (%d bytes)", len(p))
		}
		if width == 4 && uint64(len(p)) > math.MaxUint32 {
			return fmt.Errorf("argot: payload too long for len32 (%d bytes)", len(p))
		}
		var b [4]byte
		if width == 2 {
			binary.BigEndian.PutUint16(b[:2], uint16(len(p)))
			buf.Write(b[:2])
		} else {
			binary.BigEndian.PutUint32(b[:], uint32(len(p)))
			buf.Write(b[:])
		}
		buf.Write(p)

	default:
		return fmt.Errorf("argot: unknown wire encoding %q", string(sl.Encoding))
	}
	return nil
}

// readSlot 按槽位布局读取单个值
func readSlot(r *bytes.Reader, sl SlotLayout) (any, error) {
	switch sl.Encoding {
	case EncVarint:
		n, err := readUvarint(r)
		if err != nil {
			return nil, err
		}
		if sl.Type == ValInt {
			return int64(n), nil
		}
		return n, nil

	case EncVarintZigzag:
		u, err := readUvarint(r)
		if err != nil {
			return nil, err
		}
		return int64(u>>1) ^ -int64(u&1), nil

	case EncFixed64:
		var b [8]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return nil, fmt.Errorf("%w: fixed64 slot", ErrTruncated)
		}
		bits := binary.BigEndian.Uint64(b[:])
		switch sl.Type {
		case ValInt:
			return int64(bits), nil
		case ValUint:
			return bits, nil
		default:
			return math.Float64frombits(bits), nil
		}

	case EncFixed8:
		b, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("%w: fixed8 slot", ErrTruncated)
		}
		if b > 1 {
			return nil, fmt.Errorf("argot: invalid bool byte 0x%02x", b)
		}
		return b == 1, nil

	case EncLen16, EncLen32:
		n, err := readLength(r, sl.Encoding == EncLen16)
		if err != nil {
			return nil, err
		}
		p := make([]byte, n)
		if _, err := io.ReadFull(r, p); err != nil {
			return nil, fmt.Errorf("%w: payload of %d bytes", ErrTruncated, n)
		}
		return decodePayload(sl.Type, p)

	default:
		return nil, fmt.Errorf("argot: unknown wire encoding %q", string(sl.Encoding))
	}
}

func readUvarint(r *bytes.Reader) (uint64, error) {
	n, err := binary.ReadUvarint(r)
	if err != nil {
		return 0, fmt.Errorf("%w: varint", ErrTruncated)
	}
	return n, nil
}

func readLength(r *bytes.Reader, len16 bool) (int, error) {
	if len16 {
		var b [2]byte
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, fmt.Errorf("%w: len16 prefix", ErrTruncated)
		}
		return int(binary.BigEndian.Uint16(b[:])), nil
	}
	var b [4]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, fmt.Errorf("%w: len32 prefix", ErrTruncated)
	}
	return int(binary.BigEndian.Uint32(b[:])), nil
}

// asPayload 把值转换为长度前缀型负载
func asPayload(t ValueType, v any) ([]byte, error) {
	switch t {
	case ValJSON:
		return json.Marshal(v)
	case ValString:
		s, err := asString(v)
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	case ValBytes:
		return asBytes(v)
	default:
		return nil, fmt.Errorf("argot: payload encoding is not applicable to %s", t)
	}
}

// decodePayload 把负载还原为对应类型的值
func decodePayload(t ValueType, p []byte) (any, error) {
	switch t {
	case ValJSON:
		var out any
		if err := json.Unmarshal(p, &out); err != nil {
			return nil, fmt.Errorf("argot: invalid json payload: %w", err)
		}
		return out, nil
	case ValString:
		return string(p), nil
	case ValBytes:
		return p, nil
	default:
		return nil, fmt.Errorf("argot: payload encoding is not applicable to %s", t)
	}
}

// ---- 入参类型归一 ----

func asInt64(v any) (int64, error) {
	switch n := v.(type) {
	case int:
		return int64(n), nil
	case int8:
		return int64(n), nil
	case int16:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case int64:
		return n, nil
	case uint:
		if uint64(n) > math.MaxInt64 {
			return 0, fmt.Errorf("argot: %d overflows int64", n)
		}
		return int64(n), nil
	case uint8:
		return int64(n), nil
	case uint16:
		return int64(n), nil
	case uint32:
		return int64(n), nil
	case uint64:
		if n > math.MaxInt64 {
			return 0, fmt.Errorf("argot: %d overflows int64", n)
		}
		return int64(n), nil
	case float32:
		return int64(n), nil
	case float64:
		return int64(n), nil
	case json.Number:
		return n.Int64()
	default:
		return 0, fmt.Errorf("argot: cannot use %T as int", v)
	}
}

func asUint64(v any) (uint64, error) {
	switch n := v.(type) {
	case uint:
		return uint64(n), nil
	case uint8:
		return uint64(n), nil
	case uint16:
		return uint64(n), nil
	case uint32:
		return uint64(n), nil
	case uint64:
		return n, nil
	case int:
		if n < 0 {
			return 0, fmt.Errorf("argot: %d is negative", n)
		}
		return uint64(n), nil
	case int8:
		if n < 0 {
			return 0, fmt.Errorf("argot: %d is negative", n)
		}
		return uint64(n), nil
	case int16:
		if n < 0 {
			return 0, fmt.Errorf("argot: %d is negative", n)
		}
		return uint64(n), nil
	case int32:
		if n < 0 {
			return 0, fmt.Errorf("argot: %d is negative", n)
		}
		return uint64(n), nil
	case int64:
		if n < 0 {
			return 0, fmt.Errorf("argot: %d is negative", n)
		}
		return uint64(n), nil
	case json.Number:
		n64, err := n.Int64()
		if err != nil {
			return 0, err
		}
		if n64 < 0 {
			return 0, fmt.Errorf("argot: %d is negative", n64)
		}
		return uint64(n64), nil
	default:
		return 0, fmt.Errorf("argot: cannot use %T as uint", v)
	}
}

func asFloat64(v any) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int8:
		return float64(n), nil
	case int16:
		return float64(n), nil
	case int32:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case uint:
		return float64(n), nil
	case uint8:
		return float64(n), nil
	case uint16:
		return float64(n), nil
	case uint32:
		return float64(n), nil
	case uint64:
		return float64(n), nil
	case json.Number:
		return n.Float64()
	default:
		return 0, fmt.Errorf("argot: cannot use %T as float", v)
	}
}

func asString(v any) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", fmt.Errorf("argot: cannot use %T as string", v)
}

func asBytes(v any) ([]byte, error) {
	if b, ok := v.([]byte); ok {
		return b, nil
	}
	return nil, fmt.Errorf("argot: cannot use %T as []byte", v)
}

func asBool(v any) (bool, error) {
	if b, ok := v.(bool); ok {
		return b, nil
	}
	return false, fmt.Errorf("argot: cannot use %T as bool", v)
}
