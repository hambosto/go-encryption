package decryption

import (
	"fmt"
	"os"
	"strings"
)

func validateInputFile(inputFile string) error {
	if !strings.HasSuffix(inputFile, ".enc") {
		return fmt.Errorf("input file must have .enc extension")
	}

	_, err := os.Stat(inputFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist")
	}

	return nil
}

func validateOutputFile(outputFile string) error {
	_, err := os.Stat(outputFile)
	if err == nil {
		return fmt.Errorf("output file already exists")
	}
	return nil
}
