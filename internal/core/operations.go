package core

import (
	"fmt"
	"os"

	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
	"github.com/hambosto/go-encryption/internal/worker"
)

type OperationType string

const (
	OperationEncrypt OperationType = "encryption"
	OperationDecrypt OperationType = "decryption"
	Encrypt          OperationType = "Encrypt"
	Decrypt          OperationType = "Decrypt"
	encExtension                   = ".enc"
)

type OperationConfig struct {
	InputPath  string
	OutputPath string
	Password   string
	Operation  OperationType
}

type Operations struct {
	fileManager FileManagerInterface
	userPrompt  PromptInterface
}

func NewOperation(fileManager FileManagerInterface, userPrompt PromptInterface) *Operations {
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

func (op *Operations) deriveKey(password string) ([]byte, []byte, error) {
	kdf, err := kdf.NewBuilder().WithMemory(128).WithIterations(6).WithParallelism(8).WithKeyLength(64).WithSaltLength(32).Build()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create KDF: %v", err)
	}

	salt, err := kdf.GenerateSalt()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key, err := kdf.DeriveKey([]byte(password), salt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive key: %w", err)
	}

	return key, salt, nil
}

func (op *Operations) handleCleanup(path string, isEncryption bool) error {
	shouldDelete, deleteType, err := op.userPrompt.ConfirmDelete(
		path,
		fmt.Sprintf("Delete %s file", map[bool]string{true: "original", false: "encrypted"}[isEncryption]),
	)
	if err != nil {
		return fmt.Errorf("deletion prompt failed: %w", err)
	}

	if shouldDelete {
		if err := op.fileManager.Delete(path, deleteType); err != nil {
			return fmt.Errorf("file deletion failed: %w", err)
		}
	}

	return nil
}

func (op *Operations) performEncryption(input *os.File, output *os.File, fileInfo os.FileInfo, key []byte, salt []byte) error {
	processor, err := worker.NewWorkerStream(key, true)
	if err != nil {
		return fmt.Errorf("encryption processor creation failed: %w", err)
	}

	headerBuilder, err := header.NewHeaderBuilder().WithSalt(salt).WithOriginalSize(uint64(fileInfo.Size())).WithAesNonce(processor.GetAESNonce()).WithChaCha20Nonce(processor.GetChaCha20Nonce()).Build()
	if err != nil {
		return fmt.Errorf("header building failed: %w", err)
	}

	if err = header.NewHeaderWriter(header.NewBinaryHeaderIO()).Write(output, headerBuilder); err != nil {
		return fmt.Errorf("header writing failed: %w", err)
	}

	if err = processor.Process(input, output, fileInfo.Size()); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return nil
}

func (op *Operations) performDecryption(input *os.File, output *os.File, key []byte, fileHeader header.Header) error {
	processor, err := worker.NewWorkerStream(key, false)
	if err != nil {
		return fmt.Errorf("decryption processor creation failed: %w", err)
	}

	if err := processor.SetAESNonce(fileHeader.AesNonce.Value); err != nil {
		return fmt.Errorf("AES nonce setting failed: %w", err)
	}

	if err := processor.SetChaCha20Nonce(fileHeader.ChaCha20Nonce.Value); err != nil {
		return fmt.Errorf("ChaCha20 nonce setting failed: %w", err)
	}

	if err := processor.Process(input, output, int64(fileHeader.OriginalSize.Value)); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	return nil
}

func (op *Operations) handleEncryption(config OperationConfig) error {
	input, inputInfo, err := op.fileManager.OpenInputFile(config.InputPath)
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

func (op *Operations) handleDecryption(config OperationConfig) error {
	input, _, err := op.fileManager.OpenInputFile(config.InputPath)
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

	kdfBuilder, err := kdf.NewBuilder().WithMemory(128).WithIterations(6).WithParallelism(8).WithKeyLength(64).WithSaltLength(32).Build()
	if err != nil {
		return fmt.Errorf("failed to create KDF: %v", err)
	}

	key, err := kdfBuilder.DeriveKey([]byte(password), fileHeader.Salt.Value)
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
