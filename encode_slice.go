package msgpack

import (
	"math"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

func (e Encoder) AppendBytesLen(dst []byte, l int) []byte {
	if l < 256 {
		return e.append1(dst, msgpcode.Bin8, uint8(l))
	}
	if l <= math.MaxUint16 {
		return e.append2(dst, msgpcode.Bin16, uint16(l))
	}
	return e.append4(dst, msgpcode.Bin32, uint32(l))
}

func (e Encoder) appendStringLen(dst []byte, l int) []byte {
	if l < 32 {
		return e.appendCode(dst, msgpcode.FixedStrLow|byte(l))
	}
	if l < 256 {
		return e.append1(dst, msgpcode.Str8, uint8(l))
	}
	if l <= math.MaxUint16 {
		return e.append2(dst, msgpcode.Str16, uint16(l))
	}
	return e.append4(dst, msgpcode.Str32, uint32(l))
}

func (e Encoder) AppendString(dst []byte, v string) []byte {
	return e.appendNormalString(dst, v)
}

func (e Encoder) appendNormalString(dst []byte, v string) []byte {
	dst = e.appendStringLen(dst, len(v))
	return append(dst, v...)
}

func (e Encoder) AppendBytes(dst []byte, v []byte) []byte {
	if v == nil {
		return e.AppendNil(dst)
	}
	dst = e.AppendBytesLen(dst, len(v))
	return append(dst, v...)
}

func (e Encoder) AppendArrayLen(dst []byte, l int) []byte {
	if l < 16 {
		return e.appendCode(dst, msgpcode.FixedArrayLow|byte(l))
	}
	if l <= math.MaxUint16 {
		return e.append2(dst, msgpcode.Array16, uint16(l))
	}
	return e.append4(dst, msgpcode.Array32, uint32(l))
}
