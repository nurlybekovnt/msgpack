package msgpack

import (
	"fmt"
	"math"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

func (e Encoder) AppendExtHeader(dst []byte, extID int8, extLen int) []byte {
	dst = e.appendExtLen(dst, extLen)
	return append(dst, byte(extID))
}

func AppendExtHeader(dst []byte, extID int8, extLen int) []byte {
	return DefaultEncoder.AppendExtHeader(dst, extID, extLen)
}

func (e Encoder) appendExtLen(dst []byte, l int) []byte {
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

func (d *Decoder) DecodeExtHeader() (extID int8, extLen int, err error) {
	c, err := d.readCode()
	if err != nil {
		return
	}
	return d.extHeader(c)
}

func (d *Decoder) extHeader(c byte) (int8, int, error) {
	extLen, err := d.parseExtLen(c)
	if err != nil {
		return 0, 0, err
	}

	extID, err := d.readCode()
	if err != nil {
		return 0, 0, err
	}

	return int8(extID), extLen, nil
}

func (d *Decoder) parseExtLen(c byte) (int, error) {
	switch c {
	case msgpcode.FixExt1:
		return 1, nil
	case msgpcode.FixExt2:
		return 2, nil
	case msgpcode.FixExt4:
		return 4, nil
	case msgpcode.FixExt8:
		return 8, nil
	case msgpcode.FixExt16:
		return 16, nil
	case msgpcode.Ext8:
		n, err := d.uint8()
		return int(n), err
	case msgpcode.Ext16:
		n, err := d.uint16()
		return int(n), err
	case msgpcode.Ext32:
		n, err := d.uint32()
		return int(n), err
	default:
		return 0, fmt.Errorf("msgpack: invalid code=%x decoding ext len", c)
	}
}

func (d *Decoder) skipExt(c byte) error {
	n, err := d.parseExtLen(c)
	if err != nil {
		return err
	}
	return d.skipN(n + 1)
}

func (d *Decoder) skipExtHeader(c byte) error {
	// Read ext type.
	_, err := d.readCode()
	if err != nil {
		return err
	}
	// Read ext body len.
	for i := 0; i < extHeaderLen(c); i++ {
		_, err := d.readCode()
		if err != nil {
			return err
		}
	}
	return nil
}

func extHeaderLen(c byte) int {
	switch c {
	case msgpcode.Ext8:
		return 1
	case msgpcode.Ext16:
		return 2
	case msgpcode.Ext32:
		return 4
	}
	return 0
}
