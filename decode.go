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
	unsafeDecodingFlag uint32 = 1 << iota
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
}

// Reset resets the Decoder to be decoding from b.
func (d *Decoder) Reset(b []byte) {
	d.data = b
	d.i = 0
}

// NewDecoder returns a new Decoder decoding from b.
func NewDecoder(b []byte) *Decoder { return &Decoder{data: b} }

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
