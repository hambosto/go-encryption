package operations

import (
	"fmt"
	"os"
)

func (cp *CryptoProcessor) handleEncryption(config CryptoConfig) error {
	// Open input file
	input, inputInfo, err := cp.openInputFile(config.InputPath)
	if err != nil {
		return err
	}
	defer input.Close()

	// Create output file
	output, err := cp.fileManager.CreateOutput(config.OutputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	// Get password if not provided
	password := config.Password
	if password == "" {
		password, err = cp.userPrompt.GetPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	// Derive key and generate salt
	key, salt, err := cp.deriveKey(password)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypting %s...\n", config.InputPath)

	if err = cp.performEncryption(input, output, inputInfo, key, salt); err != nil {
		output.Close()
		os.Remove(config.OutputPath)
		return err
	}

	if err = cp.handleCleanup(config.InputPath, true); err != nil {
		return err
	}

	fmt.Printf("File %s encrypted successfully\n", config.OutputPath)
	return nil
}
