package msgpack

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/nurlybekovnt/msgpack/msgpcode"
)

var timeExtID int8 = -1

func (e Encoder) AppendTime(dst []byte, tm time.Time) []byte {
	secs := uint64(tm.Unix())
	if secs>>34 == 0 {
		data := uint64(tm.Nanosecond())<<34 | secs

		if data&0xffffffff00000000 == 0 {
			dst = e.appendTimeHdr(dst, 4)
			return binary.BigEndian.AppendUint32(dst, uint32(data))
		}

		dst = e.appendTimeHdr(dst, 8)
		return binary.BigEndian.AppendUint64(dst, data)
	}

	dst = e.appendTimeHdr(dst, 12)
	dst = binary.BigEndian.AppendUint32(dst, uint32(tm.Nanosecond()))
	return binary.BigEndian.AppendUint64(dst, secs)
}

func AppendTime(dst []byte, tm time.Time) []byte {
	return DefaultEncoder.AppendTime(dst, tm)
}

func (e Encoder) appendTimeHdr(dst []byte, timeLen int) []byte {
	dst = e.appendExtLen(dst, timeLen)
	return append(dst, byte(timeExtID))
}

func (d *Decoder) DecodeTime() (time.Time, error) {
	c, err := d.readCode()
	if err != nil {
		return time.Time{}, err
	}

	// Legacy format.
	if c == msgpcode.FixedArrayLow|2 {
		sec, err := d.DecodeInt64()
		if err != nil {
			return time.Time{}, err
		}

		nsec, err := d.DecodeInt64()
		if err != nil {
			return time.Time{}, err
		}

		return time.Unix(sec, nsec), nil
	}

	if msgpcode.IsString(c) {
		s, err := d.string(c)
		if err != nil {
			return time.Time{}, err
		}
		return time.Parse(time.RFC3339Nano, s)
	}

	extID, extLen, err := d.extHeader(c)
	if err != nil {
		return time.Time{}, err
	}

	// NodeJS seems to use extID 13.
	if extID != timeExtID && extID != 13 {
		return time.Time{}, fmt.Errorf("msgpack: invalid time ext id=%d", extID)
	}

	tm, err := d.decodeTime(extLen)
	if err != nil {
		return tm, err
	}

	if tm.IsZero() {
		// Zero time does not have timezone information.
		return tm.UTC(), nil
	}
	return tm, nil
}

func (d *Decoder) decodeTime(extLen int) (time.Time, error) {
	b, err := d.readN(extLen)
	if err != nil {
		return time.Time{}, err
	}

	switch len(b) {
	case 4:
		sec := binary.BigEndian.Uint32(b)
		return time.Unix(int64(sec), 0), nil
	case 8:
		sec := binary.BigEndian.Uint64(b)
		nsec := int64(sec >> 34)
		sec &= 0x00000003ffffffff
		return time.Unix(int64(sec), nsec), nil
	case 12:
		nsec := binary.BigEndian.Uint32(b)
		sec := binary.BigEndian.Uint64(b[4:])
		return time.Unix(int64(sec), int64(nsec)), nil
	default:
		err = fmt.Errorf("msgpack: invalid ext len=%d decoding time", extLen)
		return time.Time{}, err
	}
}
