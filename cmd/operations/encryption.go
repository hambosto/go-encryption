package operations

import (
	"fmt"
	"os"
)

func (op *Operations) handleEncryption(config OperationConfig) error {
	input, inputInfo, err := op.openInputFile(config.InputPath)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := op.fileManager.CreateOutput(config.OutputPath)
	if err != nil {
		return err
	}
	defer output.Close()

	password := config.Password
	if password == "" {
		password, err = op.userPrompt.GetPassword()
		if err != nil {
			return fmt.Errorf("password prompt failed: %w", err)
		}
	}

	key, salt, err := op.deriveKey(password)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypting %s...\n", config.InputPath)

	if err = op.performEncryption(input, output, inputInfo, key, salt); err != nil {
		output.Close()
		os.Remove(config.OutputPath)
		return err
	}

	if err = op.handleCleanup(config.InputPath, true); err != nil {
		return err
	}

	fmt.Printf("File %s encrypted successfully\n", config.OutputPath)
	return nil
}
