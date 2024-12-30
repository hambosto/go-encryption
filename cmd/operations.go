package cmd

import (
	"fmt"
	"strings"

	"github.com/hambosto/go-encryption/cmd/filemanager"
	"github.com/hambosto/go-encryption/cmd/operations"
)

func RunEncryption(input string) error {
	fileManager := filemanager.NewFileManager(3)
	userPrompt := filemanager.NewUserPrompt(fileManager)
	processor := operations.NewCryptoProcessor(fileManager, userPrompt)

	config := operations.CryptoConfig{
		InputPath:  input,
		OutputPath: input + ".enc",
		Operation:  operations.OperationEncrypt,
	}

	if err := processor.Process(config); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return nil
}

func RunDecryption(input string) error {
	fileManager := filemanager.NewFileManager(3)
	userPrompt := filemanager.NewUserPrompt(fileManager)
	processor := operations.NewCryptoProcessor(fileManager, userPrompt)

	config := operations.CryptoConfig{
		InputPath:  input,
		OutputPath: strings.TrimSuffix(input, ".enc"),
		Operation:  operations.OperationDecrypt,
	}

	if err := processor.Process(config); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	return nil
}
