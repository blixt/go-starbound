package starbound

import (
	"os"
	"testing"
)

func getWorld(log logger) *World {
	file, err := os.Open("../test.world")
	if err != nil {
		log.Fatalf("failed to open world file: %v", err)
	}
	w, err := NewWorld(file)
	if err != nil {
		log.Fatalf("failed to read world: %v", err)
	}
	return w
}

func BenchmarkWorldMetadata(b *testing.B) {
	w := getWorld(b)
	for i := 0; i < b.N; i++ {
		w.Get(0, 0, 0)
	}
}

func BenchmarkWorldTilesFail(b *testing.B) {
	w := getWorld(b)
	for i := 0; i < b.N; i++ {
		_, err := w.GetTiles(123, 456)
		if err != ErrKeyNotFound {
			b.Fatalf("expected ErrKeyNotFound, but got: %v", err)
		}
	}
}

func BenchmarkWorldTilesSuccess(b *testing.B) {
	w := getWorld(b)
	for i := 0; i < b.N; i++ {
		_, err := w.GetTiles(30, 21)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
