package operations

import "fmt"

func (cp *CryptoProcessor) validateOperation(config CryptoConfig) error {
	if err := cp.fileManager.Validate(config.InputPath, true); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	if err := cp.fileManager.Validate(config.OutputPath, false); err != nil {
		overwrite, promptErr := cp.userPrompt.ConfirmOverwrite(config.OutputPath)
		if promptErr != nil {
			return fmt.Errorf("overwrite prompt failed: %w", promptErr)
		}
		if !overwrite {
			return fmt.Errorf("%s cancelled by user", config.Operation)
		}
	}

	return nil
}
