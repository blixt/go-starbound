package starbound

import (
	"os"
	"testing"
)

func getDB(log logger) *BTreeDB5 {
	file, err := os.Open("../test.world")
	if err != nil {
		log.Fatalf("failed to open world file: %v", err)
	}
	db, err := NewBTreeDB5(file)
	if err != nil {
		log.Fatalf("failed to read world: %v", err)
	}
	return db
}

func TestHeader(t *testing.T) {
	db := getDB(t)
	if db.Name != "World4" {
		t.Errorf("incorrect database name: %v", db.Name)
	}
}

func TestInvalidKeyLength(t *testing.T) {
	db := getDB(t)
	_, err := db.Get([]byte("\x00\x00\x00\x00"))
	if err != ErrInvalidKeyLength {
		t.Errorf("expected invalid key length, got: %v", err)
	}
}

func TestMissingKey(t *testing.T) {
	db := getDB(t)
	data, err := db.Get([]byte("\x00\x00\x00\x00\x01"))
	if data != nil {
		t.Error("data should be <nil>")
	}
	if err != ErrKeyNotFound {
		t.Errorf("expected key error, got: %v", err)
	}
}

func BenchmarkHeader(b *testing.B) {
	file, err := os.Open("../test.world")
	if err != nil {
		b.Fatalf("failed to open world file: %v", err)
	}
	for i := 0; i < b.N; i++ {
		NewBTreeDB5(file)
	}
}

func BenchmarkLookupFail(b *testing.B) {
	db := getDB(b)
	for i := 0; i < b.N; i++ {
		_, err := db.Get([]byte("\x04\x03\x02\x01\x00"))
		if err != ErrKeyNotFound {
			b.Fatalf("expected ErrKeyNotFound, but got: %v", err)
		}
	}
}

func BenchmarkLookupSuccess(b *testing.B) {
	db := getDB(b)
	for i := 0; i < b.N; i++ {
		_, err := db.Get([]byte("\x00\x00\x00\x00\x00"))
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
