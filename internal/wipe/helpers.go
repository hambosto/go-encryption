package wipe

import (
	"crypto/rand"
	"fmt"
	"os"
)

func openFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY, 0)
}

func overwriteFile(file *os.File, size int64) error {
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to the beginning of file: %w", err)
	}

	randomData := make([]byte, 4096)
	remaining := size

	for remaining > 0 {
		if _, err := rand.Read(randomData); err != nil {
			return fmt.Errorf("failed to generate random data: %w", err)
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
	return nil
}
