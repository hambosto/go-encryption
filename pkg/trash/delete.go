package trash

import (
	"fmt"
	"os"
)

func SecureDelete(path string, passes int) error {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open file for secure deletion: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	for pass := 0; pass < passes; pass++ {
		if err := OverwriteWithRandom(file, info.Size()); err != nil {
			return fmt.Errorf("secure overwrite pass %d failed: %w", pass+1, err)
		}
	}

	return os.Remove(path)
}
