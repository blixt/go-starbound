package starbound

import (
	"io"
)

func ReadVarint(r io.Reader) (v int64, err error) {
	buf := make([]byte, 1)
	for {
		if _, err = r.Read(buf); err != nil {
			return
		}
		if buf[0]&0x80 == 0 {
			return v<<7 | int64(buf[0]), nil
		}
		v = v<<7 | int64(buf[0]&0x7F)
	}
}
