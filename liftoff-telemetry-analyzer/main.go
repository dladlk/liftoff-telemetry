package main

import (
	"flag"
	"fmt"
	"os"

	track "github.com/dladlk/liftoff-auto-drone/track"
)

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] path_to_telemetry_file\n", os.Args[0])
		flag.PrintDefaults()
	}
	doHelp := flag.Bool("help", false, "Prints help")

	flag.Parse()
	if *doHelp || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	for _, path := range flag.Args() {
		t := track.Track{}
		err := t.Open(path)
		if err != nil {
			fmt.Printf("Failed to read track file %s: %v", path, err)
			os.Exit(1)
		}
	}
}
