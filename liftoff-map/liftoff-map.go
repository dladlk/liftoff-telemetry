package main

import (
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

type MinMax struct {
	min float32
	max float32
}

func (t *MinMax) add(v float32) {
	if v < t.min {
		t.min = v
	}
	if v > t.max {
		t.max = v
	}
}

func main() {

	f, err := os.Open("data.csv")
	if err != nil {
		log.Fatal(err)
	}
	// Ensure the file is closed when the function returns.
	defer f.Close()

	// 2. Create a new CSV reader using the opened file.
	csvReader := csv.NewReader(f)

	xMinMax := MinMax{min: math.MaxFloat32}
	yMinMax := MinMax{min: math.MaxFloat32}

	var path [][]float32
	// 3. Loop indefinitely, reading one record at a time.
	for {
		// Read one record (a slice of strings) from the CSV file.
		rec, err := csvReader.Read()

		// Check if the end of the file has been reached.
		if err == io.EOF {
			break // Exit the loop
		}
		// Handle any other potential errors during reading.
		if err != nil {
			log.Fatal(err)
		}

		// 4. Process the read line (record).
		// Each record is a slice of strings, where each element is a field.
		//fmt.Printf("Record: %v\n", rec[3])
		position := strings.Split(rec[3][1:len(rec[3])-1], " ")
		x, _ := strconv.ParseFloat(position[0], 32)
		y, _ := strconv.ParseFloat(position[2], 32)
		//fmt.Printf("%.5f - %.5f\n", x, y)
		row := []float32{float32(x), float32(y)}

		if row[0] == 0 && row[1] == 0 {
			continue
		}

		path = append(path, row)
		xMinMax.add(row[0])
		yMinMax.add(row[1])
	}

	log.Printf("Loaded %d rows, x MinMax %+v, y MinMax %+v", len(path), xMinMax, yMinMax)

	const padding = 10
	const scale = 5
	const flipX = false
	const flipY = true
	xSize := int(math.Round(float64(xMinMax.max-xMinMax.min)))*scale + 2*padding
	ySize := int(math.Round(float64(yMinMax.max-yMinMax.min)))*scale + 2*padding

	fmt.Printf("Size %dx%d", xSize, ySize)

	// Background canvas
	width, height := xSize, ySize
	destinationRect := image.Rect(0, 0, width, height)
	// Use image.NewRGBA to create a writable image
	dst := image.NewRGBA(destinationRect)

	// Fill the destination with a background color (e.g., white)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	draw.Draw(dst, dst.Bounds(), &image.Uniform{C: white}, image.Point{X: 0, Y: 0}, draw.Src)

	startSrc := &image.Uniform{C: color.RGBA{R: 0, G: 255, B: 0, A: 255}}
	tripSrc := &image.Uniform{C: color.RGBA{R: 0, G: 0, B: 255, A: 255}}
	endSrc := &image.Uniform{C: color.RGBA{R: 255, G: 0, B: 0, A: 255}}

	var src *image.Uniform

	for i := range path {
		pos := path[i]
		con := []int{int(pos[0]-xMinMax.min)*scale + padding, int(pos[1]-yMinMax.min)*scale + padding}
		if flipX {
			con[0] = width - con[0]
		}
		if flipY {
			con[1] = height - con[1]
		}

		// fmt.Printf("%v %v\n", con[0], con[1])
		// Draw the blue square at (50, 50) on the destination image

		size := 1
		if i < 2 || i > len(path)-2 {
			size = 3
		}

		drawRect := image.Rect(con[0]-size, con[1]-size, con[0]+size, con[1]+size) // Destination rectangle for the source

		// 4. Use draw.Draw to perform the composition
		// draw.Over blends the images based on alpha; draw.Src copies the source directly

		switch {
		case i < 10:
			src = startSrc
		case i > len(path)-10:
			src = endSrc
		default:
			src = tripSrc
		}

		draw.Draw(dst, drawRect, src, image.Point{X: 0, Y: 0}, draw.Over)
	}

	// 5. Save the result to a file (e.g., PNG)
	outputFile, err := os.Create("path.png")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, dst); err != nil {
		panic(err)
	}
	openFile(outputFile)

}

func openFile(outputFile *os.File) {
	url := outputFile.Name()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	default: // linux
		cmd = exec.Command("xdg-open", url)
	}
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start a separate process to open generated report in browser: %v", err)
	}

}
