package starbound

import (
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrDidNotReachLeaf  = errors.New("starbound: did not reach a leaf node")
	ErrInvalidData      = errors.New("starbound: data appears to be corrupt")
	ErrInvalidKeyLength = errors.New("starbound: invalid key length")
	ErrKeyNotFound      = errors.New("starbound: key not found")
)

func ReadVersionedJSON(r io.Reader) (v VersionedJSON, err error) {
	// Read the name of the data structure.
	v.Name, err = ReadString(r)
	if err != nil {
		return
	}
	// Read unknown byte which is always 0x01.
	buf := make([]byte, 1)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	if buf[0] != 1 {
		return v, ErrInvalidData
	}
	// Read the data structure version.
	var version int32
	err = binary.Read(r, binary.BigEndian, &version)
	if err != nil {
		return
	}
	v.Version = int(version)
	// Finally, read the JSON-like data itself.
	v.Data, err = ReadDynamic(r)
	return
}

// VersionedJSON represents a JSON-compatible data structure which additionally
// has a name and version associated with it so that the reader may migrate the
// structure based on the name/version.
type VersionedJSON struct {
	Name    string
	Version int
	Data    interface{}
}
