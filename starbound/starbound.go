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
	// Read version flag (boolean).
	buf := make([]byte, 1)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}
	if buf[0] == 0 {
		// This data structure has no version.
		v.HasVersion = false
		v.Version = -1
	} else {
		v.HasVersion = true
		// Read the data structure version.
		var version int32
		err = binary.Read(r, binary.BigEndian, &version)
		if err != nil {
			return
		}
		v.Version = int(version)
	}
	// Finally, read the JSON-like data itself.
	v.Data, err = ReadDynamic(r)
	return
}

// VersionedJSON represents a JSON-compatible data structure which additionally
// has a name and version associated with it so that the reader may migrate the
// structure based on the name/version.
type VersionedJSON struct {
	Name       string
	HasVersion bool
	Version    int
	Data       interface{}
}

// Gets the list at the specified key path if there is one; otherwise, nil.
func (v *VersionedJSON) List(keys ...string) []interface{} {
	if val, ok := v.Value(keys...).([]interface{}); ok {
		return val
	} else {
		return nil
	}
}

// Gets the map at the specified key path if there is one; otherwise, nil.
func (v *VersionedJSON) Map(keys ...string) map[string]interface{} {
	if val, ok := v.Value(keys...).(map[string]interface{}); ok {
		return val
	} else {
		return nil
	}
}

// Gets the value at a specific key path if there is one; otherwise, nil.
func (v *VersionedJSON) Value(keys ...string) interface{} {
	val := v.Data
	for _, key := range keys {
		m, ok := val.(map[string]interface{})
		if !ok {
			return nil
		}
		val = m[key]
	}
	return val
}
