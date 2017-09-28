package starbound

import (
	"bytes"
	"encoding/binary"
	"io"
)

type IndexEntry struct {
	Offset, Size int64
}

func NewSBAsset6(r io.ReaderAt) (a *SBAsset6, err error) {
	buf := make([]byte, 16)
	// Read the initial header which points at a metadata section.
	_, err = r.ReadAt(buf, 0)
	if err != nil {
		return
	}
	if !bytes.Equal(buf[:8], []byte("SBAsset6")) {
		return nil, ErrInvalidData
	}
	a = &SBAsset6{r: r}
	a.metadata = int64(binary.BigEndian.Uint64(buf[8:]))
	// Read the metadata section of the asset file.
	buf5 := buf[:5]
	_, err = r.ReadAt(buf5, a.metadata)
	if err != nil {
		return
	}
	if !bytes.Equal(buf5, []byte("INDEX")) {
		return nil, ErrInvalidData
	}
	rr := &readerAtReader{r: r, off: a.metadata + 5}
	a.Metadata, err = ReadMap(rr)
	if err != nil {
		return
	}
	c, err := ReadVaruint(rr)
	if err != nil {
		return
	}
	a.FileCount = int(c)
	a.index = rr.off
	return
}

type SBAsset6 struct {
	FileCount int
	Index     map[string]IndexEntry
	Metadata  map[string]interface{}

	r        io.ReaderAt
	metadata int64
	index    int64
}

func (a *SBAsset6) GetReader(path string) (r io.Reader, err error) {
	entry, ok := a.Index[path]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return io.NewSectionReader(a.r, entry.Offset, entry.Size), nil
}

func (a *SBAsset6) ReadIndex() error {
	a.Index = make(map[string]IndexEntry, a.FileCount)
	buf := make([]byte, 16)
	r := &readerAtReader{r: a.r, off: a.index}
	for i := 0; i < a.FileCount; i++ {
		// Read the path, which will be used as the index key.
		key, err := ReadString(r)
		if err != nil {
			return err
		}
		// Read the offset and size of the file.
		_, err = r.Read(buf)
		if err != nil {
			return err
		}
		var entry IndexEntry
		entry.Offset = int64(binary.BigEndian.Uint64(buf))
		entry.Size = int64(binary.BigEndian.Uint64(buf[8:]))
		// Update the map.
		a.Index[key] = entry
	}
	return nil
}
