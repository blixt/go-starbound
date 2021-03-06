package starbound

import (
	"bytes"
	"io"
)

var (
	BTreeDB5BlockFree  = []byte("FF")
	BTreeDB5BlockIndex = []byte("II")
	BTreeDB5BlockLeaf  = []byte("LL")
)

func NewBTreeDB5(r io.ReaderAt) (db *BTreeDB5, err error) {
	db = &BTreeDB5{r: r}
	header := make([]byte, 67)
	n, err := r.ReadAt(header, 0)
	if n != len(header) || err != nil {
		return nil, ErrInvalidData
	}
	if !bytes.Equal(header[:8], []byte("BTreeDB5")) {
		return nil, ErrInvalidData
	}
	db.BlockSize = getInt(header, 8)
	db.Name = string(bytes.TrimRight(header[12:28], "\x00"))
	db.KeySize = getInt(header, 28)
	db.Swap = (header[32] == 1)
	db.freeBlock1 = getInt(header, 33)
	db.unknown1 = getInt(header, 37)
	db.freeBlock1End = getInt(header, 41)
	db.rootBlock1 = getInt(header, 45)
	db.rootBlock1IsLeaf = (header[49] == 1)
	db.freeBlock2 = getInt(header, 50)
	db.unknown2 = getInt(header, 54)
	db.freeBlock2End = getInt(header, 58)
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
	freeBlock1End          int
	freeBlock2End          int
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

func (db *BTreeDB5) GetReader(key []byte) (r io.Reader, err error) {
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
		if !bytes.Equal(bufType, BTreeDB5BlockIndex) {
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
	lr := NewLeafReader(db, block)
	if _, err = lr.Read(bufBlock); err != nil {
		return
	}
	keyCount := getInt(buf, 0)
	for i := 0; i < keyCount; i += 1 {
		if _, err = lr.Read(bufKey); err != nil {
			return
		}
		var n uint64
		if n, err = ReadVaruint(lr); err != nil {
			return
		}
		// Is this the key you're looking for?
		if bytes.Equal(bufKey, key) {
			// Key found. Return a reader for the value.
			return io.LimitReader(lr, int64(n)), nil
		}
		// This isn't the key you're looking for.
		err = lr.Skip(int(n))
		if err != nil {
			return
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
	return 512 + int64(block*db.BlockSize)
}

func NewLeafReader(db *BTreeDB5, block int) *LeafReader {
	l := &LeafReader{
		db:   db,
		buf4: make([]byte, 4),
		cur:  db.blockOffset(block),
	}
	l.buf2 = l.buf4[:2]
	return l
}

type LeafReader struct {
	db         *BTreeDB5
	buf2, buf4 []byte
	cur, end   int64
}

// Reads n bytes from the LeafReader, jumping to the next leaf when necessary.
func (l *LeafReader) Read(p []byte) (n int, err error) {
	off, n, err := l.step(len(p))
	if err != nil {
		return
	}
	n, err = l.db.r.ReadAt(p[:n], off)
	return
}

// Increments the LeafReader pointer by n bytes. This may require several reads
// from the underlying ReaderAt as the LeafReader reaches the boundary of the
// current leaf.
func (l *LeafReader) Skip(n int) error {
	for n > 0 {
		if _, m, err := l.step(n); err != nil {
			return err
		} else {
			n -= m
		}
	}
	return nil
}

func (l *LeafReader) step(max int) (off int64, n int, err error) {
	// We're at the end of the block – move to the next one.
	if l.cur == l.end {
		l.db.r.ReadAt(l.buf4, l.cur)
		l.cur = l.db.blockOffset(getInt(l.buf4, 0))
		l.end = 0
	}
	// We haven't verified that the current block is a leaf yet.
	if l.end == 0 {
		if _, err = l.db.r.ReadAt(l.buf2, l.cur); err != nil {
			return
		}
		if !bytes.Equal(l.buf2, BTreeDB5BlockLeaf) {
			return 0, 0, ErrDidNotReachLeaf
		}
		l.end = l.cur + int64(l.db.BlockSize-4)
		l.cur += 2
	}
	// Move the current pointer forward.
	off, n = l.cur, max
	if l.cur+int64(n) > l.end {
		n = int(l.end - l.cur)
	}
	l.cur += int64(n)
	return
}
