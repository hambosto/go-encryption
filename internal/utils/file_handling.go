package utils

import (
	"fmt"
	"os"
)

func DeleteOriginalFile(inputFile string, deleteType string) error {
	switch deleteType {
	case "Normal delete (faster, but recoverable)":
		return os.Remove(inputFile)
	case "Secure delete (slower, but unrecoverable)":
		return WipeFile(inputFile)
	default:
		return fmt.Errorf("invalid delete type")
	}
}

func PrepareOutputFile(outputFile string) (*os.File, error) {
	output, err := os.Create(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	return output, nil
}

func DeleteEncryptedFile(inputFile string, deleteType string) error {
	switch deleteType {
	case "Normal delete (faster, but recoverable)":
		return os.Remove(inputFile)
	case "Secure delete (slower, but unrecoverable)":
		return WipeFile(inputFile)
	default:
		return fmt.Errorf("invalid delete type")
	}
}
