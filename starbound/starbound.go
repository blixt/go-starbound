package starbound

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrDidNotReachLeaf  = errors.New("starbound: did not reach a leaf node")
	ErrInvalidHeader    = errors.New("starbound: invalid header")
	ErrInvalidKeyLength = errors.New("starbound: invalid key length")
	ErrInvalidSBON      = errors.New("starbound: invalid SBON data")
	ErrKeyNotFound      = errors.New("starbound: key not found")
)

const (
	WorldDatabaseName = "World4"
)

type Tile struct {
	ForegroundMaterial    int16
	ForegroundHueShift    uint8
	ForegroundVariant     uint8
	ForegroundMod         int16
	ForegroundModHueShift uint8
	BackgroundMaterial    int16
	BackgroundHueShift    uint8
	BackgroundVariant     uint8
	BackgroundMod         int16
	BackgroundModHueShift uint8
	Liquid                uint8
	LiquidLevel           float32
	LiquidPressure        float32
	LiquidInfinite        uint8 // bool
	Collision             uint8
	DungeonID             uint16
	Biome1, Biome2        uint8
	Indestructible        uint8 // bool
}

// NewWorld creates and initializes a new World using r as the data source.
func NewWorld(r io.ReaderAt) (w *World, err error) {
	db, err := NewBTreeDB5(r)
	if err != nil {
		return
	}
	if db.Name != WorldDatabaseName || db.KeySize != 5 {
		return nil, ErrInvalidHeader
	}
	return &World{db}, nil
}

// A World is a representation of a Starbound world, enabling read access to
// individual regions in the world as well as its metadata.
type World struct {
	*BTreeDB5
}

func (w *World) Get(layer, x, y int) (data []byte, err error) {
	src, err := w.GetReader(layer, x, y)
	if err != nil {
		return
	}
	dst := new(bytes.Buffer)
	_, err = io.Copy(dst, src)
	if err != nil {
		return
	}
	return dst.Bytes(), nil
}

func (w *World) GetReader(layer, x, y int) (r io.Reader, err error) {
	key := []byte{byte(layer), byte(x >> 8), byte(x), byte(y >> 8), byte(y)}
	lr, err := w.BTreeDB5.GetReader(key)
	if err != nil {
		return nil, err
	}
	return zlib.NewReader(lr)
}

func (w *World) GetTiles(x, y int) (t []Tile, err error) {
	r, err := w.GetReader(1, x, y)
	if err != nil {
		return
	}
	// Ignore the first three bytes.
	// TODO: Do something with these bytes?
	discard := make([]byte, 3)
	_, err = io.ReadFull(r, discard)
	if err != nil {
		return
	}
	t = make([]Tile, 1024) // 32x32 tiles in a region
	err = binary.Read(r, binary.BigEndian, t)
	return
}

// Reads a 32-bit integer from the provided buffer and offset.
func getInt(data []byte, n int) int {
	return int(data[n])<<24 | int(data[n+1])<<16 | int(data[n+2])<<8 | int(data[n+3])
}

type logger interface {
	Fatalf(format string, args ...interface{})
}
