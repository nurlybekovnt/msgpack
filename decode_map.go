package msgpack

import (
	"github.com/nurlybekovnt/msgpack/msgpcode"
)

func (d *Decoder) decodeMapDefault() (interface{}, error) {
	if d.mapDecoder != nil {
		return d.mapDecoder(d)
	}
	return d.DecodeMap()
}

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

func (d *Decoder) decodeMapStringStringPtr(ptr *map[string]string) error {
	size, err := d.DecodeMapLen()
	if err != nil {
		return err
	}
	if size == -1 {
		*ptr = nil
		return nil
	}

	m := *ptr
	if m == nil {
		*ptr = make(map[string]string, min(size, maxMapSize))
		m = *ptr
	}

	for i := 0; i < size; i++ {
		mk, err := d.DecodeString()
		if err != nil {
			return err
		}
		mv, err := d.DecodeString()
		if err != nil {
			return err
		}
		m[mk] = mv
	}

	return nil
}

func (d *Decoder) decodeMapStringInterfacePtr(ptr *map[string]interface{}) error {
	m, err := d.DecodeMap()
	if err != nil {
		return err
	}
	*ptr = m
	return nil
}

// DecodeMap decodes a map with string keys and interface{} values.
//
// Enabling the UnsafeDecoding flag may improve decoding speed but could lead to
// potential memory issues as the strings and byte slices reference the
// underlying decoding byte slice. Exercise caution when modifying decoded data
// in this mode.
func (d *Decoder) DecodeMap() (map[string]interface{}, error) {
	n, err := d.DecodeMapLen()
	if err != nil {
		return nil, err
	}

	if n == -1 {
		return nil, nil
	}

	m := make(map[string]interface{}, min(n, maxMapSize))

	for i := 0; i < n; i++ {
		mk, err := d.DecodeString()
		if err != nil {
			return nil, err
		}
		mv, err := d.decodeInterfaceCond()
		if err != nil {
			return nil, err
		}
		m[mk] = mv
	}

	return m, nil
}

// DecodeUntypedMap decodes a map with interface{} keys and values.
//
// Enabling the UnsafeDecoding flag may improve decoding speed but could lead to
// potential memory issues as the strings and byte slices reference the
// underlying decoding byte slice. Exercise caution when modifying decoded data
// in this mode.
func (d *Decoder) DecodeUntypedMap() (map[interface{}]interface{}, error) {
	n, err := d.DecodeMapLen()
	if err != nil {
		return nil, err
	}

	if n == -1 {
		return nil, nil
	}

	m := make(map[interface{}]interface{}, min(n, maxMapSize))

	for i := 0; i < n; i++ {
		mk, err := d.decodeInterfaceCond()
		if err != nil {
			return nil, err
		}

		mv, err := d.decodeInterfaceCond()
		if err != nil {
			return nil, err
		}

		m[mk] = mv
	}

	return m, nil
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
