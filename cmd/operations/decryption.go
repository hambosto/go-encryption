package operations

import (
	"fmt"
	"os"

	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
)

func (cp *CryptoProcessor) handleDecryption(config CryptoConfig) error {
	input, _, err := cp.openInputFile(config.InputPath)
	if err != nil {
		return err
	}
	defer input.Close()

	reader := header.NewHeaderReader(header.NewBinaryHeaderIO())
	fileHeader, err := reader.Read(input)
	if err != nil {
		return fmt.Errorf("header reading failed: %w", err)
	}

	password := config.Password
	if password == "" {
		password, err = cp.userPrompt.GetPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	key, err := kdf.Derive([]byte(password), fileHeader.Salt.Value)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	output, err := cp.fileManager.CreateOutput(config.OutputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	fmt.Printf("Decrypting %s...\n", config.InputPath)

	if err = cp.performDecryption(input, output, key, fileHeader); err != nil {
		output.Close()
		os.Remove(config.OutputPath)
		return err
	}

	if err = cp.handleCleanup(config.InputPath, false); err != nil {
		return err
	}

	fmt.Printf("File %s decrypted successfully\n", config.OutputPath)
	return nil
}
