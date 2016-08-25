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

func BenchmarkWorldEntities(b *testing.B) {
	w := getWorld(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := w.GetEntities(30, 21)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWorldMetadata(b *testing.B) {
	w := getWorld(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := w.ReadMetadata()
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkWorldTiles(b *testing.B) {
	w := getWorld(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := w.GetTiles(30, 21)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
