package msgpack

import (
	"fmt"
	"time"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

const (
	sortMapKeysFlag uint32 = 1 << iota
	useCompactIntsFlag
	useCompactFloatsFlag
)

// DefaultEncoder is the default Encoder and is used by Append*.
var DefaultEncoder = Encoder{}

type Encoder struct {
	flags uint32
}

// SetSortMapKeys causes the Encoder to encode map keys in increasing order.
// Supported map types are:
//   - map[string]string
//   - map[string]interface{}
func (e *Encoder) SetSortMapKeys(on bool) {
	if on {
		e.flags |= sortMapKeysFlag
	} else {
		e.flags &= ^sortMapKeysFlag
	}
}

// UseCompactEncoding causes the Encoder to chose the most compact encoding. For
// example, it allows to encode small Go int64 as msgpack int8 saving 7 bytes.
func (e *Encoder) UseCompactInts(on bool) {
	if on {
		e.flags |= useCompactIntsFlag
	} else {
		e.flags &= ^useCompactIntsFlag
	}
}

// UseCompactFloats causes the Encoder to chose a compact integer encoding for
// floats that can be represented as integers.
func (e *Encoder) UseCompactFloats(on bool) {
	if on {
		e.flags |= useCompactFloatsFlag
	} else {
		e.flags &= ^useCompactFloatsFlag
	}
}

func (e Encoder) Append(dst []byte, v interface{}) []byte {
	switch v := v.(type) {
	case nil:
		return e.AppendNil(dst)
	case string:
		return e.AppendString(dst, v)
	case []byte:
		return e.AppendBytes(dst, v)
	case int:
		return e.AppendInt(dst, int64(v))
	case int8:
		return e.appendInt8Cond(dst, v)
	case int16:
		return e.appendInt16Cond(dst, v)
	case int32:
		return e.appendInt32Cond(dst, v)
	case int64:
		return e.appendInt64Cond(dst, v)
	case uint:
		return e.AppendUint(dst, uint64(v))
	case uint8:
		return e.appendUint8Cond(dst, v)
	case uint16:
		return e.appendUint16Cond(dst, v)
	case uint32:
		return e.appendUint32Cond(dst, v)
	case uint64:
		return e.appendUint64Cond(dst, v)
	case bool:
		return e.AppendBool(dst, v)
	case float32:
		return e.AppendFloat32(dst, v)
	case float64:
		return e.AppendFloat64(dst, v)
	case time.Duration:
		return e.appendInt64Cond(dst, int64(v))
	case time.Time:
		return e.AppendTime(dst, v)
	case []string:
		return e.AppendStringSlice(dst, v)
	case map[string]string:
		if e.flags&sortMapKeysFlag != 0 {
			return e.appendSortedMapStringString(dst, v)
		} else {
			return e.appendMapStringString(dst, v)
		}
	case map[string]interface{}:
		if e.flags&sortMapKeysFlag != 0 {
			return e.AppendMapSorted(dst, v)
		} else {
			return e.AppendMap(dst, v)
		}
	default:
		panic(fmt.Errorf("unsupported type: %T", v))
	}
}

func Append(dst []byte, v interface{}) []byte {
	return DefaultEncoder.Append(dst, v)
}

func (e Encoder) AppendMulti(dst []byte, v ...interface{}) []byte {
	for _, vv := range v {
		dst = e.Append(dst, vv)
	}
	return nil
}

func AppendMulti(dst []byte, v ...interface{}) []byte {
	return DefaultEncoder.AppendMulti(dst, v)
}

func (e Encoder) AppendNil(dst []byte) []byte {
	return e.appendCode(dst, msgpcode.Nil)
}

func AppendNil(dst []byte) []byte {
	return DefaultEncoder.AppendNil(dst)
}

func (e Encoder) AppendBool(dst []byte, value bool) []byte {
	if value {
		return e.appendCode(dst, msgpcode.True)
	}
	return e.appendCode(dst, msgpcode.False)
}

func AppendBool(dst []byte, value bool) []byte {
	return DefaultEncoder.AppendBool(dst, value)
}

func (e Encoder) AppendDuration(dst []byte, d time.Duration) []byte {
	return e.AppendInt(dst, int64(d))
}

func AppendDuration(dst []byte, d time.Duration) []byte {
	return DefaultEncoder.AppendDuration(dst, d)
}

func (e Encoder) appendCode(dst []byte, c byte) []byte {
	return append(dst, c)
}
