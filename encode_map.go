package msgpack

import (
	"math"
	"sort"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

func (e Encoder) appendMapStringString(dst []byte, m map[string]string) []byte {
	if m == nil {
		return e.AppendNil(dst)
	}

	dst = e.AppendMapLen(dst, len(m))

	for mk, mv := range m {
		dst = e.AppendString(dst, mk)
		dst = e.AppendString(dst, mv)
	}

	return dst
}

func (e Encoder) AppendMap(dst []byte, m map[string]interface{}) []byte {
	if m == nil {
		return e.AppendNil(dst)
	}
	dst = e.AppendMapLen(dst, len(m))
	for mk, mv := range m {
		dst = e.AppendString(dst, mk)
		dst = e.Append(dst, mv)
	}
	return dst
}

func AppendMap(dst []byte, m map[string]interface{}) []byte {
	return DefaultEncoder.AppendMap(dst, m)
}

func (e Encoder) AppendMapSorted(dst []byte, m map[string]interface{}) []byte {
	if m == nil {
		return e.AppendNil(dst)
	}
	dst = e.AppendMapLen(dst, len(m))

	keys := make([]string, 0, len(m))

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		dst = e.AppendString(dst, k)
		dst = e.Append(dst, m[k])
	}

	return dst
}

func AppendMapSorted(dst []byte, m map[string]interface{}) []byte {
	return DefaultEncoder.AppendMapSorted(dst, m)
}

func (e Encoder) appendSortedMapStringString(dst []byte, m map[string]string) []byte {
	if m == nil {
		return e.AppendNil(dst)
	}

	dst = e.AppendMapLen(dst, len(m))

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		dst = e.AppendString(dst, k)
		dst = e.AppendString(dst, m[k])
	}

	return dst
}

func (e Encoder) AppendMapLen(dst []byte, l int) []byte {
	if l < 16 {
		return e.appendCode(dst, msgpcode.FixedMapLow|byte(l))
	}
	if l <= math.MaxUint16 {
		return e.append2(dst, msgpcode.Map16, uint16(l))
	}
	return e.append4(dst, msgpcode.Map32, uint32(l))
}

func AppendMapLen(dst []byte, l int) []byte {
	return DefaultEncoder.AppendMapLen(dst, l)
}
