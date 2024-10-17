package compression

import (
	"bytes"
	"compress/zlib"
	"io"
)

// ZlibCompress compresses data using zlib compression algorithm.
// It takes a byte slice of data and returns a compressed byte slice
// along with any potential error encountered during compression.
func ZlibCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer      // Buffer to hold compressed data
	w := zlib.NewWriter(&buf) // Create a new zlib writer

	// Write data to the zlib writer
	if _, err := w.Write(data); err != nil {
		return nil, err // Return error if writing fails
	}

	// Close the writer to flush any remaining data
	if err := w.Close(); err != nil {
		return nil, err // Return error if closing fails
	}

	return buf.Bytes(), nil // Return the compressed data
}

// ZlibDecompress decompresses zlib-compressed data and returns the original bytes.
// It takes a byte slice of compressed data and returns the decompressed byte slice
// along with any potential error encountered during decompression.
func ZlibDecompress(compressed []byte) ([]byte, error) {
	// Create a new zlib reader from the compressed data
	r, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err // Return error if creating reader fails
	}
	defer r.Close() // Ensure the reader is closed after use

	// Read all decompressed data from the zlib reader
	return io.ReadAll(r) // Return the decompressed data
}

