package msgpack

import (
	"github.com/nurlybekovnt/msgpack/msgpcode"
)

// DecodeMapLen decodes map length. Length is -1 when map is nil.
func (d *Decoder) DecodeMapLen() (int, error) {
	c, err := d.readCode()
	if err != nil {
		return 0, err
	}

	if msgpcode.IsExt(c) {
		if err = d.skipExtHeader(c); err != nil {
			return 0, err
		}

		c, err = d.readCode()
		if err != nil {
			return 0, err
		}
	}
	return d.mapLen(c)
}

func (d *Decoder) mapLen(c byte) (int, error) {
	if c == msgpcode.Nil {
		return -1, nil
	}
	if c >= msgpcode.FixedMapLow && c <= msgpcode.FixedMapHigh {
		return int(c & msgpcode.FixedMapMask), nil
	}
	if c == msgpcode.Map16 {
		size, err := d.uint16()
		return int(size), err
	}
	if c == msgpcode.Map32 {
		size, err := d.uint32()
		return int(size), err
	}
	return 0, unexpectedCodeError{code: c, hint: "map length"}
}

func (d *Decoder) skipMap(c byte) error {
	n, err := d.mapLen(c)
	if err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if err := d.Skip(); err != nil {
			return err
		}
		if err := d.Skip(); err != nil {
			return err
		}
	}
	return nil
}
