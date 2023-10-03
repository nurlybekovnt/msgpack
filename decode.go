package msgpack

import (
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

const (
	looseInterfaceDecodingFlag uint32 = 1 << iota
	unsafeDecodingFlag
)

const (
	sliceAllocLimit = 1e4
	maxMapSize      = 1e6
)

var decPool = sync.Pool{
	New: func() interface{} {
		return NewDecoder(nil)
	},
}

func GetDecoder() *Decoder {
	return decPool.Get().(*Decoder)
}

func PutDecoder(dec *Decoder) {
	*dec = Decoder{}
	decPool.Put(dec)
}

type Decoder struct {
	flags uint32
	data  []byte
	i     int // current reading index

	mapDecoder func(*Decoder) (interface{}, error)
}

// Reset resets the Decoder to be decoding from b.
func (d *Decoder) Reset(b []byte) {
	d.data = b
	d.i = 0
}

// NewDecoder returns a new Decoder decoding from b.
func NewDecoder(b []byte) *Decoder { return &Decoder{data: b} }

func (d *Decoder) SetMapDecoder(fn func(*Decoder) (interface{}, error)) {
	d.mapDecoder = fn
}

// UseLooseInterfaceDecoding causes decoder to use DecodeInterfaceLoose
// to decode msgpack value into Go interface{}.
func (d *Decoder) UseLooseInterfaceDecoding(on bool) {
	if on {
		d.flags |= looseInterfaceDecodingFlag
	} else {
		d.flags &= ^looseInterfaceDecodingFlag
	}
}

// UseUnsafeDecoding causes the decoder to reference the underlying data when
// decoding strings and byte slices to avoid memory allocation and provide
// faster decoding speed. However, this approach requires careful use as it can
// lead to memory corruption and memory leaks. When the flag is off, the decoder
// allocates new memory for strings and byte slices. That is the default
// behaviour.
func (d *Decoder) UseUnsafeDecoding(on bool) {
	if on {
		d.flags |= unsafeDecodingFlag
	} else {
		d.flags &= ^unsafeDecodingFlag
	}
}

func (d *Decoder) unsafeDecoding() bool {
	return d.flags&unsafeDecodingFlag != 0
}

// Decode decodes the Msgpack-encoded data and stores the result in the value
// pointed to by v. If v is nil or not a pointer, Decode returns an error.
//
// Enabling the UnsafeDecoding flag may improve decoding speed but could lead to
// potential memory issues as the strings and byte slices reference the
// underlying decoding byte slice. Exercise caution when modifying decoded data
// in this mode.
func (d *Decoder) Decode(v interface{}) error {
	var err error
	switch v := v.(type) {
	case *string:
		if v != nil {
			*v, err = d.DecodeString()
			return err
		}
	case *[]byte:
		if v != nil {
			return d.decodeBytesPtr(v)
		}
	case *int:
		if v != nil {
			*v, err = d.DecodeInt()
			return err
		}
	case *int8:
		if v != nil {
			*v, err = d.DecodeInt8()
			return err
		}
	case *int16:
		if v != nil {
			*v, err = d.DecodeInt16()
			return err
		}
	case *int32:
		if v != nil {
			*v, err = d.DecodeInt32()
			return err
		}
	case *int64:
		if v != nil {
			*v, err = d.DecodeInt64()
			return err
		}
	case *uint:
		if v != nil {
			*v, err = d.DecodeUint()
			return err
		}
	case *uint8:
		if v != nil {
			*v, err = d.DecodeUint8()
			return err
		}
	case *uint16:
		if v != nil {
			*v, err = d.DecodeUint16()
			return err
		}
	case *uint32:
		if v != nil {
			*v, err = d.DecodeUint32()
			return err
		}
	case *uint64:
		if v != nil {
			*v, err = d.DecodeUint64()
			return err
		}
	case *bool:
		if v != nil {
			*v, err = d.DecodeBool()
			return err
		}
	case *float32:
		if v != nil {
			*v, err = d.DecodeFloat32()
			return err
		}
	case *float64:
		if v != nil {
			*v, err = d.DecodeFloat64()
			return err
		}
	case *[]string:
		return d.decodeStringSlicePtr(v)
	case *map[string]string:
		return d.decodeMapStringStringPtr(v)
	case *map[string]interface{}:
		return d.decodeMapStringInterfacePtr(v)
	case *time.Duration:
		if v != nil {
			vv, err := d.DecodeInt64()
			*v = time.Duration(vv)
			return err
		}
	case *time.Time:
		if v != nil {
			*v, err = d.DecodeTime()
			return err
		}
	}

	return fmt.Errorf("msgpack: Decode(unsupported %T)", v)
}

func (d *Decoder) DecodeMulti(v ...interface{}) error {
	for _, vv := range v {
		if err := d.Decode(vv); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) decodeInterfaceCond() (interface{}, error) {
	if d.flags&looseInterfaceDecodingFlag != 0 {
		return d.DecodeInterfaceLoose()
	}
	return d.DecodeInterface()
}

func (d *Decoder) DecodeNil() error {
	c, err := d.readCode()
	if err != nil {
		return err
	}
	if c != msgpcode.Nil {
		return fmt.Errorf("msgpack: invalid code=%x decoding nil", c)
	}
	return nil
}

func (d *Decoder) DecodeBool() (bool, error) {
	c, err := d.readCode()
	if err != nil {
		return false, err
	}
	return d.bool(c)
}

func (d *Decoder) bool(c byte) (bool, error) {
	if c == msgpcode.Nil {
		return false, nil
	}
	if c == msgpcode.False {
		return false, nil
	}
	if c == msgpcode.True {
		return true, nil
	}
	return false, fmt.Errorf("msgpack: invalid code=%x decoding bool", c)
}

func (d *Decoder) DecodeDuration() (time.Duration, error) {
	n, err := d.DecodeInt64()
	if err != nil {
		return 0, err
	}
	return time.Duration(n), nil
}

// DecodeInterface decodes value into interface. It returns following types:
//   - nil,
//   - bool,
//   - int8, int16, int32, int64,
//   - uint8, uint16, uint32, uint64,
//   - float32 and float64,
//   - string,
//   - []byte,
//   - slices of any of the above,
//   - maps of any of the above.
//
// DecodeInterface should be used only when you don't know the type of value you
// are decoding. For example, if you are decoding number it is better to use
// DecodeInt64 for negative numbers and DecodeUint64 for positive numbers.
//
// Enabling the UnsafeDecoding flag may improve decoding speed but could lead to
// potential memory issues as the strings and byte slices reference the
// underlying decoding byte slice. Exercise caution when modifying decoded data
// in this mode.
func (d *Decoder) DecodeInterface() (interface{}, error) {
	c, err := d.readCode()
	if err != nil {
		return nil, err
	}

	if msgpcode.IsFixedNum(c) {
		return int8(c), nil
	}
	if msgpcode.IsFixedMap(c) {
		err = d.unreadByte()
		if err != nil {
			return nil, err
		}
		return d.decodeMapDefault()
	}
	if msgpcode.IsFixedArray(c) {
		return d.decodeSlice(c)
	}
	if msgpcode.IsFixedString(c) {
		return d.string(c)
	}

	switch c {
	case msgpcode.Nil:
		return nil, nil
	case msgpcode.False, msgpcode.True:
		return d.bool(c)
	case msgpcode.Float:
		return d.float32(c)
	case msgpcode.Double:
		return d.float64(c)
	case msgpcode.Uint8:
		return d.uint8()
	case msgpcode.Uint16:
		return d.uint16()
	case msgpcode.Uint32:
		return d.uint32()
	case msgpcode.Uint64:
		return d.uint64()
	case msgpcode.Int8:
		return d.int8()
	case msgpcode.Int16:
		return d.int16()
	case msgpcode.Int32:
		return d.int32()
	case msgpcode.Int64:
		return d.int64()
	case msgpcode.Bin8, msgpcode.Bin16, msgpcode.Bin32:
		return d.bytes(c)
	case msgpcode.Str8, msgpcode.Str16, msgpcode.Str32:
		return d.string(c)
	case msgpcode.Array16, msgpcode.Array32:
		return d.decodeSlice(c)
	case msgpcode.Map16, msgpcode.Map32:
		err = d.unreadByte()
		if err != nil {
			return nil, err
		}
		return d.decodeMapDefault()
		// case msgpcode.FixExt1, msgpcode.FixExt2, msgpcode.FixExt4, msgpcode.FixExt8, msgpcode.FixExt16,
		// 	msgpcode.Ext8, msgpcode.Ext16, msgpcode.Ext32:
		// 	return d.decodeInterfaceExt(c)
	}

	return 0, fmt.Errorf("msgpack: unknown code %x decoding interface{}", c)
}

// DecodeInterfaceLoose is like DecodeInterface except that:
//   - int8, int16, and int32 are converted to int64,
//   - uint8, uint16, and uint32 are converted to uint64,
//   - float32 is converted to float64.
//   - []byte is converted to string.
//
// Enabling the UnsafeDecoding flag may improve decoding speed but could lead to
// potential memory issues as the strings and byte slices reference the
// underlying decoding byte slice. Exercise caution when modifying decoded data
// in this mode.
func (d *Decoder) DecodeInterfaceLoose() (interface{}, error) {
	c, err := d.readCode()
	if err != nil {
		return nil, err
	}

	if msgpcode.IsFixedNum(c) {
		return int64(int8(c)), nil
	}
	if msgpcode.IsFixedMap(c) {
		err = d.unreadByte()
		if err != nil {
			return nil, err
		}
		return d.decodeMapDefault()
	}
	if msgpcode.IsFixedArray(c) {
		return d.decodeSlice(c)
	}
	if msgpcode.IsFixedString(c) {
		return d.string(c)
	}

	switch c {
	case msgpcode.Nil:
		return nil, nil
	case msgpcode.False, msgpcode.True:
		return d.bool(c)
	case msgpcode.Float, msgpcode.Double:
		return d.float64(c)
	case msgpcode.Uint8, msgpcode.Uint16, msgpcode.Uint32, msgpcode.Uint64:
		return d.uint(c)
	case msgpcode.Int8, msgpcode.Int16, msgpcode.Int32, msgpcode.Int64:
		return d.int(c)
	case msgpcode.Str8, msgpcode.Str16, msgpcode.Str32,
		msgpcode.Bin8, msgpcode.Bin16, msgpcode.Bin32:
		return d.string(c)
	case msgpcode.Array16, msgpcode.Array32:
		return d.decodeSlice(c)
	case msgpcode.Map16, msgpcode.Map32:
		err = d.unreadByte()
		if err != nil {
			return nil, err
		}
		return d.decodeMapDefault()
		// case msgpcode.FixExt1, msgpcode.FixExt2, msgpcode.FixExt4, msgpcode.FixExt8, msgpcode.FixExt16,
		// 	msgpcode.Ext8, msgpcode.Ext16, msgpcode.Ext32:
		// 	return d.decodeInterfaceExt(c)
	}

	return 0, fmt.Errorf("msgpack: unknown code %x decoding interface{}", c)
}

// Skip skips next value.
func (d *Decoder) Skip() error {
	c, err := d.readCode()
	if err != nil {
		return err
	}

	if msgpcode.IsFixedNum(c) {
		return nil
	}
	if msgpcode.IsFixedMap(c) {
		return d.skipMap(c)
	}
	if msgpcode.IsFixedArray(c) {
		return d.skipSlice(c)
	}
	if msgpcode.IsFixedString(c) {
		return d.skipBytes(c)
	}

	switch c {
	case msgpcode.Nil, msgpcode.False, msgpcode.True:
		return nil
	case msgpcode.Uint8, msgpcode.Int8:
		return d.skipN(1)
	case msgpcode.Uint16, msgpcode.Int16:
		return d.skipN(2)
	case msgpcode.Uint32, msgpcode.Int32, msgpcode.Float:
		return d.skipN(4)
	case msgpcode.Uint64, msgpcode.Int64, msgpcode.Double:
		return d.skipN(8)
	case msgpcode.Bin8, msgpcode.Bin16, msgpcode.Bin32:
		return d.skipBytes(c)
	case msgpcode.Str8, msgpcode.Str16, msgpcode.Str32:
		return d.skipBytes(c)
	case msgpcode.Array16, msgpcode.Array32:
		return d.skipSlice(c)
	case msgpcode.Map16, msgpcode.Map32:
		return d.skipMap(c)
	case msgpcode.FixExt1, msgpcode.FixExt2, msgpcode.FixExt4, msgpcode.FixExt8, msgpcode.FixExt16,
		msgpcode.Ext8, msgpcode.Ext16, msgpcode.Ext32:
		return d.skipExt(c)
	}

	return fmt.Errorf("msgpack: unknown code %x", c)
}

// PeekCode returns the next MessagePack code without advancing the reader.
// Subpackage msgpack/msgpcode defines the list of available msgpcode.
func (d *Decoder) PeekCode() (byte, error) {
	c, err := d.readByte()
	if err != nil {
		return 0, err
	}
	return c, d.unreadByte()
}

// readByte reads and returns the next byte from the input or any error
// encountered. If readByte returns an error, no input byte was consumed, and
// the returned byte value is undefined.
func (d *Decoder) readByte() (byte, error) {
	if d.i >= len(d.data) {
		return 0, io.EOF
	}
	b := d.data[d.i]
	d.i++
	return b, nil
}

// unreadByte complements readByte.
func (d *Decoder) unreadByte() error {
	if d.i <= 0 {
		return errors.New("msgpack: unread byte at beginning of slice")
	}
	d.i--
	return nil
}

func (d *Decoder) readCode() (byte, error) {
	return d.readByte()
}

func (d *Decoder) readN(n int) ([]byte, error) {
	if d.i+n > len(d.data) {
		return nil, io.ErrShortBuffer
	}
	b := d.data[d.i : d.i+n]
	d.i += n
	return b, nil
}
