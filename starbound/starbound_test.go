package starbound

import (
	"os"
	"testing"
)

func BenchmarkWorldMetadata(b *testing.B) {
	file, err := os.Open("../test.world")
	if err != nil {
		b.Fatalf("failed to open world file: %v", err)
	}
	w, err := NewWorld(file)
	if err != nil {
		b.Fatalf("failed to open world file: %v", err)
	}
	for i := 0; i < b.N; i++ {
		w.Get(0, 0, 0)
	}
}
