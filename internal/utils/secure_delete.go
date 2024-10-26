package utils

import (
	"crypto/rand"
	"fmt"
	"os"
)

func SecureDelete(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	for i := 0; i < 3; i++ {
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek to beginning of file: %w", err)
		}

		randomData := make([]byte, 4096)
		remaining := info.Size()

		for remaining > 0 {

			if _, err := rand.Read(randomData); err != nil {
				return fmt.Errorf("failed to read random data: %w", err)
			}

			writeSize := int64(len(randomData))

			if remaining < writeSize {
				writeSize = remaining
			}

			if _, err := file.Write(randomData[:writeSize]); err != nil {
				return fmt.Errorf("failed to write random data: %w", err)
			}

			remaining -= writeSize
		}
	}

	return os.Remove(path)
}
