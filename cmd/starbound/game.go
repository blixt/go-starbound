package main

import (
	"image"
	"image/color"
	_ "image/png"
	"path"

	"github.com/blixt/go-starbound/starbound"
	"github.com/hajimehoshi/ebiten"
	"golang.org/x/exp/mmap"
)

func newGame(base string) *game {
	return &game{base: base}
}

type game struct {
	base   string
	cx, cy float64
	asset  *starbound.SBAsset6
	world  *starbound.World

	cursor, dirt, temp *ebiten.Image
}

func (g *game) LoadAsset(p string) error {
	file, err := mmap.Open(path.Join(g.base, p))
	if err != nil {
		return err
	}
	g.asset, err = starbound.NewSBAsset6(file)
	if err == nil {
		g.asset.ReadIndex()
	}
	return err
}

func (g *game) OpenWorld(p string) error {
	file, err := mmap.Open(path.Join(g.base, p))
	if err != nil {
		return err
	}
	g.world, err = starbound.NewWorld(file)
	if err != nil {
		return err
	}
	err = g.world.ReadMetadata()
	if err != nil {
		return err
	}
	start := g.world.Metadata.List("playerStart")
	g.cx = start[0].(float64)
	g.cy = start[1].(float64)
	return nil
}

func (g *game) Run() error {
	var err error
	g.cursor, err = g.image("/cursors/cursors.png")
	if err != nil {
		return err
	}
	g.dirt, err = g.image("/tiles/materials/dirt.png")
	if err != nil {
		return err
	}
	g.temp, err = g.region(int(g.cx/32), int(g.cy/32))
	if err != nil {
		return err
	}
	return ebiten.Run(g.tick, 400, 300, 2, "Go Starbound")
}

type slice struct {
	X0, Y0, X1, Y1 int
}

func (s slice) Len() int                       { return 1 }
func (s slice) Dst(i int) (x0, y0, x1, y1 int) { return 0, 0, s.X1 - s.X0, s.Y1 - s.Y0 }
func (s slice) Src(i int) (x0, y0, x1, y1 int) { return s.X0, s.Y0, s.X1, s.Y1 }

func (g *game) image(path string) (img *ebiten.Image, err error) {
	r, err := g.asset.GetReader(path)
	if err != nil {
		return
	}
	i, _, err := image.Decode(r)
	if err != nil {
		return
	}
	return ebiten.NewImageFromImage(i, ebiten.FilterNearest)
}

func (g *game) region(rx, ry int) (img *ebiten.Image, err error) {
	img, err = ebiten.NewImage(256, 256, ebiten.FilterNearest)
	if err != nil {
		return
	}
	img.Fill(color.White)
	tiles, err := g.world.GetTiles(rx, ry)
	if err != nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.ImageParts = slice{4, 12, 12, 20}
	var x, y int
	var tx, ty float64
	ty = 248
	for _, tile := range tiles {
		op.GeoM.SetElement(0, 2, tx)
		op.GeoM.SetElement(1, 2, ty)
		if tile.ForegroundMaterial == 8 {
			img.DrawImage(g.dirt, op)
		}
		x += 1
		tx += 8
		if x == 32 {
			x = 0
			tx = 0
			y += 1
			ty -= 8
		}
	}
	return
}

func (g *game) tick(screen *ebiten.Image) error {
	screen.DrawImage(g.temp, nil)

	op := &ebiten.DrawImageOptions{}
	op.ImageParts = slice{32, 0, 48, 16}
	x, y := ebiten.CursorPosition()
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(g.cursor, op)

	return nil
}
