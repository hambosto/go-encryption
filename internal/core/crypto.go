package core

import (
	"os"
	"strings"
)

type FileManagerInterface interface {
	Delete(path string, deleteType DeleteType) error
	CreateOutput(path string) (*os.File, error)
	Validate(path string, shouldExist bool) error
	OpenInputFile(path string) (*os.File, os.FileInfo, error)
}

type PromptInterface interface {
	ConfirmOverwrite(path string) (bool, error)
	GetPassword() (string, error)
	ConfirmDelete(path string, prompt string) (bool, DeleteType, error)
	GetOperation() (OperationType, error)
	SelectFile(files []string) (string, error)
}

type Processor struct {
	fileManager FileManagerInterface
	userPrompt  PromptInterface
	operation   *Operations
}

func NewProcessor(fileManager FileManagerInterface, userPrompt PromptInterface) *Processor {
	return &Processor{
		fileManager: fileManager,
		userPrompt:  userPrompt,
		operation:   NewOperation(fileManager, userPrompt),
	}
}

func (p *Processor) ProcessFile(input string, op OperationType) error {
	config := OperationConfig{
		InputPath:  input,
		OutputPath: determineOutputPath(input, op),
		Operation:  mapOperationType(op),
	}

	if err := p.operation.Process(config); err != nil {
		return err
	}

	return nil
}

func determineOutputPath(input string, op OperationType) string {
	if op == Encrypt {
		return input + encExtension
	}
	return strings.TrimSuffix(input, encExtension)
}

func mapOperationType(op OperationType) OperationType {
	if op == Encrypt {
		return OperationEncrypt
	}
	return OperationDecrypt
}
