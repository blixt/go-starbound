package starbound

import (
	"io"
)

// Reads a 32-bit integer from the provided buffer and offset.
func getInt(data []byte, n int) int {
	return int(data[n])<<24 | int(data[n+1])<<16 | int(data[n+2])<<8 | int(data[n+3])
}

// Allows for different concrete types that behave like loggers.
type logger interface {
	Fatalf(format string, args ...interface{})
}

// Implements the Reader interface for a ReaderAt type.
type readerAtReader struct {
	r   io.ReaderAt
	off int64
}

func (r *readerAtReader) Read(p []byte) (n int, err error) {
	// TODO: Unset err if we read data and ReaderAt is still readable.
	n, err = r.r.ReadAt(p, r.off)
	r.off += int64(n)
	return
}
