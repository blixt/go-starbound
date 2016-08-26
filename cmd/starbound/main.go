package main

import (
	"flag"
	"log"
	"path"
)

var base = flag.String("dir", ".", "Starbound directory")
var world = flag.String("world", "", "world filename")

func check(err error, action string) {
	if err == nil {
		return
	}
	log.Fatalf("failed to %s: %v\n", action, err)
}

func main() {
	flag.Parse()
	if len(*world) == 0 {
		flag.Usage()
		return
	}

	g := newGame(*base)

	var err error
	err = g.LoadAsset("assets/packed.pak")
	check(err, "load asset file")
	err = g.OpenWorld(path.Join("storage/universe", *world))
	check(err, "load world")
	err = g.Run()
	check(err, "start game")
}
