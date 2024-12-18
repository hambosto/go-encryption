package utils

import (
	"fmt"
	"os"
)

func ValidateInputFile(inputFile string) error {
	fileInfo, err := os.Stat(inputFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist")
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("input file is empty")
	}

	return nil
}

func ValidateOutputFile(outputFile string) error {
	_, err := os.Stat(outputFile)
	if err == nil {
		return fmt.Errorf("output file already exists")
	}
	return nil
}
