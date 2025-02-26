package operations

import (
	"fmt"
	"os"

	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
)

func (op *Operations) handleDecryption(config OperationConfig) error {
	input, _, err := op.openInputFile(config.InputPath)
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
		password, err = op.userPrompt.GetPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	kdf := kdf.NewWithDefaults()
	key, err := kdf.DeriveKey([]byte(password), fileHeader.Salt.Value)
	if err != nil {
		return fmt.Errorf("key derivation failed: %w", err)
	}

	output, err := op.fileManager.CreateOutput(config.OutputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	fmt.Printf("Decrypting %s...\n", config.InputPath)

	if err = op.performDecryption(input, output, key, fileHeader); err != nil {
		output.Close()
		os.Remove(config.OutputPath)
		return err
	}

	if err = op.handleCleanup(config.InputPath, false); err != nil {
		return err
	}

	fmt.Printf("File %s decrypted successfully\n", config.OutputPath)
	return nil
}
