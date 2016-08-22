package starbound

import (
	"errors"
)

var (
	ErrDidNotReachLeaf  = errors.New("starbound: did not reach a leaf node")
	ErrInvalidHeader    = errors.New("starbound: invalid header")
	ErrInvalidKeyLength = errors.New("starbound: invalid key length")
	ErrKeyNotFound      = errors.New("starbound: key not found")
)

func getInt(data []byte, n int) int {
	return int(data[n])<<24 | int(data[n+1])<<16 | int(data[n+2])<<8 | int(data[n+3])
}
