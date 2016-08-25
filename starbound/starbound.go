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
	ErrInvalidData      = errors.New("starbound: data appears to be corrupt")
	ErrInvalidKeyLength = errors.New("starbound: invalid key length")
	ErrKeyNotFound      = errors.New("starbound: key not found")
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
	if db.Name != "World4" || db.KeySize != 5 {
		return nil, ErrInvalidData
	}
	return &World{BTreeDB5: db}, nil
}

// A World is a representation of a Starbound world, enabling read access to
// individual regions in the world as well as its metadata.
type World struct {
	*BTreeDB5
	Metadata      VersionedJSON
	Width, Height int
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

func (w *World) GetEntities(x, y int) (e []VersionedJSON, err error) {
	r, err := w.GetReader(2, x, y)
	if err != nil {
		return
	}
	n, err := ReadVaruint(r)
	if err != nil {
		return
	}
	e = make([]VersionedJSON, n)
	for i := uint64(0); i < n; i++ {
		e[i], err = ReadVersionedJSON(r)
		if err != nil {
			break
		}
	}
	return
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

func (w *World) ReadMetadata() error {
	r, err := w.GetReader(0, 0, 0)
	if err != nil {
		return err
	}
	wh := make([]int32, 2)
	err = binary.Read(r, binary.BigEndian, wh)
	if err != nil {
		return err
	}
	w.Width, w.Height = int(wh[0]), int(wh[1])
	w.Metadata, err = ReadVersionedJSON(r)
	return err
}

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

// Reads a 32-bit integer from the provided buffer and offset.
func getInt(data []byte, n int) int {
	return int(data[n])<<24 | int(data[n+1])<<16 | int(data[n+2])<<8 | int(data[n+3])
}

type logger interface {
	Fatalf(format string, args ...interface{})
}
