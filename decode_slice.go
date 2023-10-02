package msgpack

import (
	"fmt"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

// DecodeArrayLen decodes array length. Length is -1 when array is nil.
func (d *Decoder) DecodeArrayLen() (int, error) {
	c, err := d.readCode()
	if err != nil {
		return 0, err
	}
	return d.arrayLen(c)
}

func (d *Decoder) arrayLen(c byte) (int, error) {
	if c == msgpcode.Nil {
		return -1, nil
	} else if c >= msgpcode.FixedArrayLow && c <= msgpcode.FixedArrayHigh {
		return int(c & msgpcode.FixedArrayMask), nil
	}
	switch c {
	case msgpcode.Array16:
		n, err := d.uint16()
		return int(n), err
	case msgpcode.Array32:
		n, err := d.uint32()
		return int(n), err
	}
	return 0, fmt.Errorf("msgpack: invalid code=%x decoding array length", c)
}

func (d *Decoder) skipSlice(c byte) error {
	n, err := d.arrayLen(c)
	if err != nil {
		return err
	}

	for i := 0; i < n; i++ {
		if err := d.Skip(); err != nil {
			return err
		}
	}

	return nil
}
