package main

import (
	"flag"
	"log"

	"github.com/dulltz/megaconfigmap/pkg/combiner"
)

var (
	megaConfigMapPath = flag.String("megaconfigmap-path", "/megaconfigmap/megaconfigmap.json", "Path of the megaconfigmap")
	shareDir          = flag.String("share-dir", "/data", "Path of the sharing directory among the pod")
)

func main() {
	flag.Parse()
	c, err := combiner.NewCombiner(*megaConfigMapPath, *shareDir)
	if err != nil {
		log.Fatal(err)
	}
	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}
