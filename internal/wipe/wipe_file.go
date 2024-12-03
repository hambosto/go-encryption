package wipe

import (
	"fmt"
	"os"
)

func WipeFile(path string) error {
	file, err := openFile(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	for i := 0; i < 3; i++ { // Overwrite 3 times
		if err := overwriteFile(file, info.Size()); err != nil {
			return fmt.Errorf("failed to overwrite file: %w", err)
		}
	}

	return os.Remove(path)
}
