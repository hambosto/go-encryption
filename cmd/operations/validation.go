package operations

import (
	"fmt"
)

func (op *Operations) validateOperation(config OperationConfig) error {
	if err := op.validatePath(config.InputPath, true); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	if err := op.validatePath(config.OutputPath, false); err != nil {
		overwrite, promptErr := op.userPrompt.ConfirmOverwrite(config.OutputPath)
		if promptErr != nil {
			return fmt.Errorf("overwrite prompt failed: %w", promptErr)
		}
		if !overwrite {
			return fmt.Errorf("%s cancelled by user", config.Operation)
		}
	}

	return nil
}

func (op *Operations) validatePath(path string, isInput bool) error {
	return op.fileManager.Validate(path, isInput)
}
