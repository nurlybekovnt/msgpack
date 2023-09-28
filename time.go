package msgpack

import (
	"encoding/binary"
	"time"
)

var timeExtID int8 = -1

func (e *Encoder) AppendTime(dst []byte, tm time.Time) []byte {
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

func (e *Encoder) appendTimeHdr(dst []byte, timeLen int) []byte {
	dst = e.appendExtLen(dst, timeLen)
	return append(dst, byte(timeExtID))
}
