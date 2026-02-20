package main

import (
	"flag"
	"fmt"
	"os"

	track "github.com/dladlk/liftoff-auto-drone/track"
	vector "github.com/dladlk/liftoff-auto-drone/vector"
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

		printLimit := 100
		zeroRPMCount := 0
		printCount := 0
		for _, d := range t.List {
			if vector.VectorZero(d.MotorRPM[:]) {
				zeroRPMCount++
				continue
			}
			fmt.Println(vector.VectorPrint("Attitude", d.Attitude), vector.VectorPrintGyro("Gyro", d.Gyro), vector.VectorPrintByDecimal("RPM", [4]float32(d.MotorRPM), 1))
			printCount++
			if printCount >= printLimit {
				break
			}
		}
		fmt.Printf("Printed %d frames, skipped %d", printCount, zeroRPMCount)
	}

}
