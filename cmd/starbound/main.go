package main

import (
	"fmt"
	"os"

	"github.com/blixt/go-starbound/starbound"
	"golang.org/x/exp/mmap"
)

func main() {
	file, err := mmap.Open("../../test.world")
	if err != nil {
		fmt.Printf("failed to open world: %v\n", err)
		os.Exit(1)
	}
	db, err := starbound.NewBTreeDB5(file)
	if err != nil {
		fmt.Printf("failed to open world: %v\n", err)
		os.Exit(1)
	}
	value, err := db.Get([]byte("\x00\x00\x00\x00\x00"))
	fmt.Printf("metadata size: %d\n", len(value))
}
