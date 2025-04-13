package trash

import (
	"crypto/rand"
	"fmt"
	"os"
)

func OverwriteWithRandom(file *os.File, size int64) error {
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to file start: %w", err)
	}

	const bufferSize = 4096
	buffer := make([]byte, bufferSize)
	remaining := size

	for remaining > 0 {
		writeSize := min(remaining, int64(len(buffer)))

		if _, err := rand.Read(buffer[:writeSize]); err != nil {
			return fmt.Errorf("failed to generate random data: %w", err)
		}

		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return fmt.Errorf("failed to write random data: %w", err)
		}

		remaining -= writeSize
	}

	return file.Sync()
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
