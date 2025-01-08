package operations

import (
	"fmt"

	"github.com/hambosto/go-encryption/cmd/filemanager"
)

type Operations struct {
	fileManager *filemanager.FileManager
	userPrompt  *filemanager.UserPrompt
}

func NewOperation(fileManager *filemanager.FileManager, userPrompt *filemanager.UserPrompt) *Operations {
	return &Operations{
		fileManager: fileManager,
		userPrompt:  userPrompt,
	}
}

func (op *Operations) Process(config OperationConfig) error {
	if err := op.validateOperation(config); err != nil {
		return err
	}

	switch config.Operation {
	case OperationEncrypt:
		return op.handleEncryption(config)
	case OperationDecrypt:
		return op.handleDecryption(config)
	default:
		return fmt.Errorf("unsupported operation: %s", config.Operation)
	}
}
