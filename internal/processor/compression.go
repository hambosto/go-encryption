package processor

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

func (p *Processor) compressData(data []byte) ([]byte, error) {
	var buffer bytes.Buffer

	zw, err := zlib.NewWriterLevel(&buffer, zlib.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer: %w", err)
	}

	if _, err := zw.Write(data); err != nil {
		zw.Close()
		return nil, fmt.Errorf("failed to write data to zlib writer: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %w", err)
	}

	return buffer.Bytes(), nil
}

func (p *Processor) decompressData(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid data: insufficient bytes for size information or invalid padding")
	}

	compressedSize := binary.BigEndian.Uint32(data)
	if compressedSize > uint32(len(data)-4) {
		return nil, fmt.Errorf("invalid compressed data size: expected %d, got %d", compressedSize, len(data)-4)
	}

	compressedData := data[4 : 4+compressedSize]
	var buffer bytes.Buffer

	zr, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zr.Close()

	if _, err := io.Copy(&buffer, zr); err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return buffer.Bytes(), nil
}

func (p *Processor) padData(data []byte) []byte {
	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(len(data)))

	alignedSize := (len(data) + 15) & ^15
	if alignedSize > len(data) {
		padding := make([]byte, alignedSize-len(data))
		data = append(data, padding...)
	}

	return append(sizeHeader, data...)
}
