package starbound

import (
	"encoding/binary"
	"io"
)

func ReadBytes(r io.Reader) (p []byte, err error) {
	n, err := ReadVaruint(r)
	if err != nil {
		return
	}
	p = make([]byte, n)
	_, err = io.ReadFull(r, p)
	return
}

func ReadDynamic(r io.Reader) (v interface{}, err error) {
	buf := make([]byte, 1)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	switch buf[0] {
	case 1:
		return nil, nil
	case 2:
		var d float64
		err = binary.Read(r, binary.BigEndian, &d)
		return d, err
	case 3:
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return
		}
		return buf[0] != 0, nil
	case 4:
		return ReadVarint(r)
	case 5:
		return ReadString(r)
	case 6:
		return ReadList(r)
	case 7:
		return ReadMap(r)
	}
	return nil, ErrInvalidData
}

func ReadList(r io.Reader) (l []interface{}, err error) {
	n, err := ReadVaruint(r)
	if err != nil {
		return
	}
	l = make([]interface{}, n)
	for i := uint64(0); i < n; i++ {
		l[i], err = ReadDynamic(r)
		if err != nil {
			break
		}
	}
	return
}

func ReadMap(r io.Reader) (m map[string]interface{}, err error) {
	n, err := ReadVaruint(r)
	if err != nil {
		return
	}
	m = make(map[string]interface{}, n)
	for i := uint64(0); i < n; i++ {
		var key string
		if key, err = ReadString(r); err == nil {
			m[key], err = ReadDynamic(r)
		}
		if err != nil {
			break
		}
	}
	return
}

func ReadString(r io.Reader) (s string, err error) {
	b, err := ReadBytes(r)
	return string(b), err
}

func ReadVarint(r io.Reader) (v int64, err error) {
	uv, err := ReadVaruint(r)
	if err != nil {
		return
	}
	if uv&1 == 1 {
		return -int64(uv>>1) - 1, nil
	} else {
		return int64(uv >> 1), nil
	}
}

func ReadVaruint(r io.Reader) (v uint64, err error) {
	buf := make([]byte, 1)
	for {
		if _, err = r.Read(buf); err != nil {
			return
		}
		if buf[0]&0x80 == 0 {
			return v<<7 | uint64(buf[0]), nil
		}
		v = v<<7 | uint64(buf[0]&0x7F)
	}
}
