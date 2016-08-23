package starbound

import (
	"bytes"
	"io"
)

const (
	BTreeDB5HeaderSize = 512
)

var (
	BlockFree  = []byte("FF")
	BlockIndex = []byte("II")
	BlockLeaf  = []byte("LL")

	HeaderSignature = []byte("BTreeDB5")
)

func NewBTreeDB5(r io.ReaderAt) (db *BTreeDB5, err error) {
	db = &BTreeDB5{r: r}
	header := make([]byte, 67)
	n, err := r.ReadAt(header, 0)
	if n != len(header) || err != nil {
		return nil, ErrInvalidHeader
	}
	if !bytes.Equal(header[:8], HeaderSignature) {
		return nil, ErrInvalidHeader
	}
	db.BlockSize = getInt(header, 8)
	db.Name = string(bytes.TrimRight(header[12:28], "\x00"))
	db.KeySize = getInt(header, 28)
	db.Swap = (header[32] == 1)
	db.freeBlock1 = getInt(header, 33)
	// Skip 3 bytes...
	db.unknown1 = getInt(header, 40)
	// Skip 1 byte...
	db.rootBlock1 = getInt(header, 45)
	db.rootBlock1IsLeaf = (header[49] == 1)
	db.freeBlock2 = getInt(header, 50)
	// Skip 3 bytes...
	db.unknown2 = getInt(header, 57)
	// Skip 1 byte...
	db.rootBlock2 = getInt(header, 62)
	db.rootBlock2IsLeaf = (header[66] == 1)
	return
}

type BTreeDB5 struct {
	Name      string
	BlockSize int
	KeySize   int
	Swap      bool

	r io.ReaderAt

	freeBlock1, freeBlock2 int
	rootBlock1, rootBlock2 int
	rootBlock1IsLeaf       bool
	rootBlock2IsLeaf       bool
	unknown1, unknown2     int
}

func (db *BTreeDB5) FreeBlock() int {
	if !db.Swap {
		return db.freeBlock1
	} else {
		return db.freeBlock2
	}
}

func (db *BTreeDB5) Get(key []byte) (data []byte, err error) {
	if len(key) != db.KeySize {
		return nil, ErrInvalidKeyLength
	}
	bufSize := 11
	if db.KeySize > bufSize {
		bufSize = db.KeySize
	}
	buf := make([]byte, bufSize)
	bufBlock := buf[:4]
	bufHead := buf[:11]
	bufKey := buf[:db.KeySize]
	bufType := buf[:2]
	block := db.RootBlock()
	offset := db.blockOffset(block)
	entrySize := db.KeySize + 4
	// Traverse the B-tree until we reach a leaf.
	for {
		if _, err = db.r.ReadAt(bufHead, offset); err != nil {
			return
		}
		if !bytes.Equal(bufType, BlockIndex) {
			break
		}
		offset += 11
		// Binary search for the key.
		lo, hi := 0, getInt(buf, 3)
		block = getInt(buf, 7)
		for lo < hi {
			mid := (lo + hi) / 2
			if _, err = db.r.ReadAt(bufKey, offset+int64(entrySize*mid)); err != nil {
				return
			}
			if bytes.Compare(key, bufKey) < 0 {
				hi = mid
			} else {
				lo = mid + 1
			}
		}
		if lo > 0 {
			// A candidate leaf/index was found in the current index. Get the block index.
			db.r.ReadAt(bufBlock, offset+int64(entrySize*(lo-1)+db.KeySize))
			block = getInt(buf, 0)
		}
		offset = db.blockOffset(block)
	}
	// Scan leaves for the key, then read the data.
	r := NewLeafReader(db, block)
	if _, err = r.Read(bufBlock); err != nil {
		return
	}
	keyCount := getInt(buf, 0)
	for i := 0; i < keyCount; i += 1 {
		if _, err = r.Read(bufKey); err != nil {
			return
		}
		var n int64
		if n, err = ReadVarint(r); err != nil {
			return
		}
		// TODO: Allow skipping without reading.
		temp := make([]byte, n)
		if _, err = io.ReadFull(r, temp); err != nil {
			return
		}
		if bytes.Equal(bufKey, key) {
			return temp, nil
		}
	}
	return nil, ErrKeyNotFound
}

func (db *BTreeDB5) RootBlock() int {
	if !db.Swap {
		return db.rootBlock1
	} else {
		return db.rootBlock2
	}
}

func (db *BTreeDB5) blockOffset(block int) int64 {
	return BTreeDB5HeaderSize + int64(block*db.BlockSize)
}

func NewLeafReader(db *BTreeDB5, block int) *LeafReader {
	return &LeafReader{
		db:  db,
		cur: db.blockOffset(block),
	}
}

type LeafReader struct {
	db       *BTreeDB5
	cur, end int64
}

func (l *LeafReader) Read(p []byte) (n int, err error) {
	buf := make([]byte, 4)
	if l.end == 0 {
		if _, err = l.db.r.ReadAt(buf[:2], l.cur); err != nil {
			return
		}
		if !bytes.Equal(buf[:2], BlockLeaf) {
			return 0, ErrDidNotReachLeaf
		}
		l.end = l.cur + int64(l.db.BlockSize-4)
		l.cur += 2
	}
	want := int64(len(p))
	if l.cur+want > l.end {
		want = l.end - l.cur
	}
	n, err = l.db.r.ReadAt(p[:want], l.cur)
	l.cur += int64(n)
	if l.cur == l.end {
		l.db.r.ReadAt(buf, l.cur)
		l.cur = l.db.blockOffset(getInt(buf, 0))
		l.end = 0
	}
	return
}
