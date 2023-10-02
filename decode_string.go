package msgpack

import (
	"fmt"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

func (d *Decoder) bytesLen(c byte) (int, error) {
	if c == msgpcode.Nil {
		return -1, nil
	}

	if msgpcode.IsFixedString(c) {
		return int(c & msgpcode.FixedStrMask), nil
	}

	switch c {
	case msgpcode.Str8, msgpcode.Bin8:
		n, err := d.uint8()
		return int(n), err
	case msgpcode.Str16, msgpcode.Bin16:
		n, err := d.uint16()
		return int(n), err
	case msgpcode.Str32, msgpcode.Bin32:
		n, err := d.uint32()
		return int(n), err
	}

	return 0, fmt.Errorf("msgpack: invalid code=%x decoding string/bytes length", c)
}

func (d *Decoder) DecodeString() (string, error) {
	c, err := d.readCode()
	if err != nil {
		return "", err
	}
	return d.string(c)
}

func (d *Decoder) string(c byte) (string, error) {
	n, err := d.bytesLen(c)
	if err != nil {
		return "", err
	}
	return d.stringWithLen(n)
}

func (d *Decoder) stringWithLen(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}

	b, err := d.readN(n)
	if err != nil {
		return "", err
	}

	if d.unsafeDecoding() {
		return bytesToString(b), nil
	} else {
		return string(b), nil
	}
}

func (d *Decoder) DecodeBytesLen() (int, error) {
	c, err := d.readCode()
	if err != nil {
		return 0, err
	}
	return d.bytesLen(c)
}

func (d *Decoder) DecodeBytes() ([]byte, error) {
	c, err := d.readCode()
	if err != nil {
		return nil, err
	}
	return d.bytes(c)
}

func (d *Decoder) bytes(c byte) ([]byte, error) {
	n, err := d.bytesLen(c)
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
	}

	b, err := d.readN(n)
	if err != nil {
		return nil, err
	}

	if d.unsafeDecoding() {
		return b, nil
	} else {
		bb := make([]byte, len(b))
		copy(bb, b)
		return bb, nil
	}
}

func (d *Decoder) skipBytes(c byte) error {
	n, err := d.bytesLen(c)
	if err != nil {
		return err
	}
	if n <= 0 {
		return nil
	}
	return d.skipN(n)
}
