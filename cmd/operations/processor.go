package operations

import (
	"fmt"

	"github.com/hambosto/go-encryption/cmd/filemanager"
)

type CryptoProcessor struct {
	fileManager *filemanager.FileManager
	userPrompt  *filemanager.UserPrompt
}

func NewCryptoProcessor(fileManager *filemanager.FileManager, userPrompt *filemanager.UserPrompt) *CryptoProcessor {
	return &CryptoProcessor{
		fileManager: fileManager,
		userPrompt:  userPrompt,
	}
}

func (cp *CryptoProcessor) Process(config CryptoConfig) error {
	if err := cp.validateOperation(config); err != nil {
		return err
	}

	switch config.Operation {
	case OperationEncrypt:
		return cp.handleEncryption(config)
	case OperationDecrypt:
		return cp.handleDecryption(config)
	default:
		return fmt.Errorf("unsupported operation: %s", config.Operation)
	}
}
