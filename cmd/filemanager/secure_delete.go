package filemanager

import (
	"crypto/rand"
	"fmt"
	"os"
)

func (fm *FileManager) secureDelete(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file for secure deletion: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	for pass := 0; pass < fm.overwritePasses; pass++ {
		if err := fm.overwriteWithRandom(file, info.Size()); err != nil {
			return fmt.Errorf("secure overwrite pass %d failed: %w", pass+1, err)
		}
	}

	return os.Remove(path)
}

func (fm *FileManager) overwriteWithRandom(file *os.File, size int64) error {
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek to file start: %w", err)
	}

	const bufferSize = 4096
	buffer := make([]byte, bufferSize)
	remaining := size

	for remaining > 0 {
		writeSize := int64(len(buffer))
		if remaining < writeSize {
			writeSize = remaining
		}

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
