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

func (d *Decoder) decodeStringSlicePtr(ptr *[]string) error {
	n, err := d.DecodeArrayLen()
	if err != nil {
		return err
	}
	if n == -1 {
		return nil
	}

	ss := makeStrings(*ptr, n)
	for i := 0; i < n; i++ {
		s, err := d.DecodeString()
		if err != nil {
			return err
		}
		ss = append(ss, s)
	}
	*ptr = ss

	return nil
}

func makeStrings(s []string, n int) []string {
	if n > sliceAllocLimit {
		n = sliceAllocLimit
	}

	if s == nil {
		return make([]string, 0, n)
	}

	if cap(s) >= n {
		return s[:0]
	}

	s = s[:cap(s)]
	s = append(s, make([]string, n-len(s))...)
	return s[:0]
}

// DecodeSlice decodes a slice of interface{}.
//
// Enabling the UnsafeDecoding flag may improve decoding speed but could lead to
// potential memory issues as the strings and byte slices reference the
// underlying decoding byte slice. Exercise caution when modifying decoded data
// in this mode.
func (d *Decoder) DecodeSlice() ([]interface{}, error) {
	c, err := d.readCode()
	if err != nil {
		return nil, err
	}
	return d.decodeSlice(c)
}

func (d *Decoder) decodeSlice(c byte) ([]interface{}, error) {
	n, err := d.arrayLen(c)
	if err != nil {
		return nil, err
	}
	if n == -1 {
		return nil, nil
	}

	s := make([]interface{}, 0, min(n, sliceAllocLimit))
	for i := 0; i < n; i++ {
		v, err := d.decodeInterfaceCond()
		if err != nil {
			return nil, err
		}
		s = append(s, v)
	}

	return s, nil
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
