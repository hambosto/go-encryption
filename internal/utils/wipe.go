package utils

import (
	"crypto/rand"
	"fmt"
	"os"
)

func WipeFile(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info for %s: %w", path, err)
	}
	fileSize := info.Size()

	const overwritePasses = 3
	for pass := 0; pass < overwritePasses; pass++ {
		if err := secureOverwrite(file, fileSize); err != nil {
			return fmt.Errorf("secure overwrite pass %d failed for %s: %w", pass+1, path, err)
		}
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to remove file %s: %w", path, err)
	}

	return nil
}

func secureOverwrite(file *os.File, fileSize int64) error {
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to file start: %w", err)
	}

	const bufferSize = 4096
	randomData := make([]byte, bufferSize)
	remaining := fileSize

	for remaining > 0 {
		if _, err := rand.Read(randomData); err != nil {
			return fmt.Errorf("failed to generate cryptographically secure random data: %w", err)
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

	return file.Sync()
}
