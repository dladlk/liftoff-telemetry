package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	lot_config "github.com/dladlk/liftoff-telemetry/data"
)

type Track struct {
	path   string
	fields []lot_config.StreamDataType
	List   []lot_config.Datagram
	minTs  float32
	maxTs  float32
}

func (t *Track) Open(path string) error {
	t.path = path

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	header, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return errors.New("Failed to read first line as header")
	}
	header = strings.TrimSpace(header)
	t.fields = lot_config.ParseStreamDataTypeFormats(strings.Split(header, ","))

	blockLength := lot_config.CalculateBlockLength(t.fields)

	fmt.Printf("File header: %s, block length: %d\r\n", header, blockLength)

	buffer := make([]byte, blockLength)

	blocks := 0

	for {
		n, err := io.ReadFull(reader, buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		if n != int(blockLength) {
			return fmt.Errorf("Expected to read %d bytes, but read only %d", blockLength, n)
		}
		blocks++
		datagram := lot_config.Datagram{}
		datagram.ParseDatagram(bytes.NewReader(buffer[:n]), &t.fields)

		t.List = append(t.List, datagram)
	}
	t.minTs = t.List[0].Timestamp
	t.maxTs = t.List[len(t.List)-1].Timestamp
	fmt.Printf("Loaded %d blocks, min ts %.2f sec, max ts %.2f sec\n", blocks, t.minTs, t.maxTs)

	return nil
}
