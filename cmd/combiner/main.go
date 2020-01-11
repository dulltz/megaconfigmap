package main

import (
	"flag"
	"log"

	"github.com/dulltz/megaconfigmap/pkg/combiner"
)

func main() {
	var megaConfigMapName = flag.String("megaconfigmap", "", "Name of the megaconfigmap")
	var shareDir = flag.String("share-dir", "/data", "Path of the sharing directory among the pod")
	flag.Parse()

	log.Println("megaconfigmap:", *megaConfigMapName)
	log.Println("share-dir:", *shareDir)
	if len(*megaConfigMapName) == 0 {
		log.Fatal("please specify --megaconfigmap")
	}
	if len(*shareDir) == 0 {
		log.Fatal("please specify --share-dir")
	}

	c, err := combiner.NewCombiner(*megaConfigMapName, *shareDir)
	if err != nil {
		log.Fatal(err)
	}
	err = c.Run()
	if err != nil {
		log.Fatal(err)
	}
}
