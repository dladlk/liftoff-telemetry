package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"sort"

	track "github.com/dladlk/liftoff-auto-drone/track"
	lot_config "github.com/dladlk/liftoff-telemetry/data"
	vector "github.com/dladlk/liftoff-telemetry/vector"
)

func main() {

	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] path_to_telemetry_file\n", os.Args[0])
		flag.PrintDefaults()
	}
	doHelp := flag.Bool("help", false, "Prints help")
	doMinMax := flag.Bool("minmax", false, "Analyze min/max values for each data")

	flag.Parse()
	if *doHelp || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	minMaxMap := map[string]*MinMax{}

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
			if *doMinMax {
				updateMinMax(&d, &minMaxMap)
			}
			if vector.VectorZero(d.MotorRPM[:]) {
				zeroRPMCount++
				continue
			}
			if printCount <= printLimit {
				printCount++
				fmt.Println(vector.VectorPrintTabbed("Attitude", d.Attitude),
					vector.VectorPrint3Short("Degrees", lot_config.AttitudeQuaternionToEulerDegrees(d.Attitude)),
					vector.VectorPrint3Long("Velocity World", d.Velocity),
					vector.VectorPrint3Long("Velocity Local", vector.VelocityWorldSpaceToLocalSpace(d.Velocity, d.Attitude)),
					vector.VectorPrint3Short("Gyro", d.Gyro),
					vector.VectorPrintByDecimal("RPM", [4]float32(d.MotorRPM), 1))
			} else {
				if !*doMinMax {
					break
				}
			}
		}
		fmt.Printf("Printed %d frames, skipped %d\n", printCount, zeroRPMCount)
		if *doMinMax {
			printMinMax(&minMaxMap)
		}
	}

}

type MinMax struct {
	name  string
	index int
	min   float32
	max   float32
}

func (m MinMax) String() string {
	return fmt.Sprintf("%s[%d]\t:\t% 9.6f\t-\t% 9.6f", m.name, m.index, m.min, m.max)
}

func (m *MinMax) update(v float32) {
	if v > m.max {
		m.max = v
	}
	if v < m.min {
		m.min = v
	}
}

func updateMinMax(d *lot_config.Datagram, minMaxMap *map[string]*MinMax) {
	updateMinMaxMap(minMaxMap, "Timestamp", []float32{d.Timestamp})
	updateMinMaxMap(minMaxMap, "Attitude", d.Attitude[:])
	updateMinMaxMap(minMaxMap, "Gyro", d.Gyro[:])
	updateMinMaxMap(minMaxMap, "Input", d.Input[:])
	updateMinMaxMap(minMaxMap, "Position", d.Position[:])
	updateMinMaxMap(minMaxMap, "Velocity", d.Velocity[:])
	updateMinMaxMap(minMaxMap, "MotorRPM", d.MotorRPM[:])
	updateMinMaxMap(minMaxMap, "Battery", d.Battery[:])

}

func updateMinMaxMap(minMaxMap *map[string]*MinMax, name string, f []float32) {
	for i, v := range f {
		key := fmt.Sprintf("%s-%d", name, i)
		minMax, ok := (*minMaxMap)[key]
		if !ok {
			minMax = &MinMax{name: name, index: i, min: math.MaxFloat32}
			(*minMaxMap)[key] = minMax
		}
		minMax.update(v)
	}
}

func printMinMax(minMaxMap *map[string]*MinMax) {
	sortedKeys := make([]string, len(*minMaxMap))
	i := 0
	for key := range *minMaxMap {
		sortedKeys[i] = key
		i++
	}
	sort.Strings(sortedKeys)
	fmt.Println("Min-max of all attributues:")
	for _, key := range sortedKeys {
		fmt.Println((*minMaxMap)[key])
	}
}
