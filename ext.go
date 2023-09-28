package msgpack

import (
	"math"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

func (e *Encoder) AppendExtHeader(dst []byte, extID int8, extLen int) []byte {
	dst = e.appendExtLen(dst, extLen)
	return append(dst, byte(extID))
}

func (e *Encoder) appendExtLen(dst []byte, l int) []byte {
	switch l {
	case 1:
		return e.appendCode(dst, msgpcode.FixExt1)
	case 2:
		return e.appendCode(dst, msgpcode.FixExt2)
	case 4:
		return e.appendCode(dst, msgpcode.FixExt4)
	case 8:
		return e.appendCode(dst, msgpcode.FixExt8)
	case 16:
		return e.appendCode(dst, msgpcode.FixExt16)
	}
	if l <= math.MaxUint8 {
		return e.append1(dst, msgpcode.Ext8, uint8(l))
	}
	if l <= math.MaxUint16 {
		return e.append2(dst, msgpcode.Ext16, uint16(l))
	}
	return e.append4(dst, msgpcode.Ext32, uint32(l))
}
