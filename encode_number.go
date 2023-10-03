package msgpack

import (
	"math"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

// AppendUint8 appends an uint8 in 2 bytes preserving type of the number.
func (e Encoder) AppendUint8(dst []byte, n uint8) []byte {
	return e.append1(dst, msgpcode.Uint8, n)
}

// AppendUint16 appends an uint16 in 3 bytes preserving type of the number.
func (e Encoder) AppendUint16(dst []byte, n uint16) []byte {
	return e.append2(dst, msgpcode.Uint16, n)
}

// AppendUint32 appends an uint16 in 5 bytes preserving type of the number.
func (e Encoder) AppendUint32(dst []byte, n uint32) []byte {
	return e.append4(dst, msgpcode.Uint32, n)
}

// AppendUint64 appends an uint16 in 9 bytes preserving type of the number.
func (e Encoder) AppendUint64(dst []byte, n uint64) []byte {
	return e.append8(dst, msgpcode.Uint64, n)
}

func (e Encoder) appendUint64Cond(dst []byte, n uint64) []byte {
	if e.flags&useCompactIntsFlag != 0 {
		return e.AppendUint(dst, n)
	}
	return e.AppendUint64(dst, n)
}

// AppendInt8 appends an int8 in 2 bytes preserving type of the number.
func (e Encoder) AppendInt8(dst []byte, n int8) []byte {
	return e.append1(dst, msgpcode.Int8, uint8(n))
}

// AppendInt16 appends an int16 in 3 bytes preserving type of the number.
func (e Encoder) AppendInt16(dst []byte, n int16) []byte {
	return e.append2(dst, msgpcode.Int16, uint16(n))
}

// AppendInt32 appends an int32 in 5 bytes preserving type of the number.
func (e Encoder) AppendInt32(dst []byte, n int32) []byte {
	return e.append4(dst, msgpcode.Int32, uint32(n))
}

// AppendInt64 appends an int64 in 9 bytes preserving type of the number.
func (e Encoder) AppendInt64(dst []byte, n int64) []byte {
	return e.append8(dst, msgpcode.Int64, uint64(n))
}

func (e Encoder) appendInt64Cond(dst []byte, n int64) []byte {
	if e.flags&useCompactIntsFlag != 0 {
		return e.AppendInt(dst, n)
	}
	return e.AppendInt64(dst, n)
}

// AppendUnsignedNumber appends an uint64 in 1, 2, 3, 5, or 9 bytes. Type of the
// number is lost during encoding.
func (e Encoder) AppendUint(dst []byte, n uint64) []byte {
	if n <= math.MaxInt8 {
		return append(dst, byte(n))
	}
	if n <= math.MaxUint8 {
		return e.AppendUint8(dst, uint8(n))
	}
	if n <= math.MaxUint16 {
		return e.AppendUint16(dst, uint16(n))
	}
	if n <= math.MaxUint32 {
		return e.AppendUint32(dst, uint32(n))
	}
	return e.AppendUint64(dst, n)
}

// AppendNumber appends an int64 in 1, 2, 3, 5, or 9 bytes. Type of the number
// is lost during encoding.
func (e Encoder) AppendInt(dst []byte, n int64) []byte {
	if n >= 0 {
		return e.AppendUint(dst, uint64(n))
	}
	if n >= int64(int8(msgpcode.NegFixedNumLow)) {
		return append(dst, byte(n))
	}
	if n >= math.MinInt8 {
		return e.AppendInt8(dst, int8(n))
	}
	if n >= math.MinInt16 {
		return e.AppendInt16(dst, int16(n))
	}
	if n >= math.MinInt32 {
		return e.AppendInt32(dst, int32(n))
	}
	return e.AppendInt64(dst, n)
}

func (e Encoder) AppendFloat32(dst []byte, n float32) []byte {
	if e.flags&useCompactFloatsFlag != 0 {
		if float32(int64(n)) == n {
			return e.AppendInt(dst, int64(n))
		}
	}
	return e.append4(dst, msgpcode.Float, math.Float32bits(n))
}

func (e Encoder) AppendFloat64(dst []byte, n float64) []byte {
	if e.flags&useCompactFloatsFlag != 0 {
		// Both NaN and Inf convert to int64(-0x8000000000000000)
		// If n is NaN then it never compares true with any other value
		// If n is Inf then it doesn't convert from int64 back to +/-Inf
		// In both cases the comparison works.
		if float64(int64(n)) == n {
			return e.AppendInt(dst, int64(n))
		}
	}
	return e.append8(dst, msgpcode.Double, math.Float64bits(n))
}

func (e Encoder) append1(dst []byte, code byte, n uint8) []byte {
	return append(dst, code, n)
}

func (e Encoder) append2(dst []byte, code byte, n uint16) []byte {
	return append(dst, code, byte(n>>8), byte(n))
}

func (e Encoder) append4(dst []byte, code byte, n uint32) []byte {
	return append(dst,
		code,
		byte(n>>24),
		byte(n>>16),
		byte(n>>8),
		byte(n),
	)
}

func (e Encoder) append8(dst []byte, code byte, n uint64) []byte {
	return append(dst,
		code,
		byte(n>>56),
		byte(n>>48),
		byte(n>>40),
		byte(n>>32),
		byte(n>>24),
		byte(n>>16),
		byte(n>>8),
		byte(n),
	)
}
