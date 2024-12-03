package decryption

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hambosto/go-encryption/internal/utils"
)

func prepareOutputFile(outputFile string) (*os.File, error) {
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	output, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	return output, nil
}

func deleteEncryptedFile(inputFile string, deleteType string) error {
	switch deleteType {
	case "Normal delete (faster, but recoverable)":
		return os.Remove(inputFile)
	case "Secure delete (slower, but unrecoverable)":
		return utils.SecureDelete(inputFile)
	default:
		return fmt.Errorf("invalid delete type")
	}
}
