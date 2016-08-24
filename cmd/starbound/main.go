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
	w, err := starbound.NewWorld(file)
	if err != nil {
		fmt.Printf("failed to open world: %v\n", err)
		os.Exit(1)
	}
	t, err := w.GetTiles(30, 21)
	if err != nil {
		fmt.Printf("failed to get region: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("region:", t)
}
