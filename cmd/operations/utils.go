package operations

import (
	"fmt"
	"os"

	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
	"github.com/hambosto/go-encryption/internal/worker"
)

func (op *Operations) openInputFile(path string) (*os.File, os.FileInfo, error) {
	input, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open input file: %w", err)
	}

	info, err := input.Stat()
	if err != nil {
		input.Close()
		return nil, nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return input, info, nil
}

func (op *Operations) deriveKey(password string) ([]byte, []byte, error) {
	kdf, err := kdf.New(
		kdf.WithMemory(128*1024), // 128 MB
		kdf.WithTimeCost(6),
		kdf.WithThreads(8),
		kdf.WithKeyLength(64),
		kdf.WithSaltLength(32),
	)
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
	worker, err := worker.NewFileProcessor(key, true)
	if err != nil {
		return fmt.Errorf("encryption processor creation failed: %w", err)
	}

	builder, err := header.NewHeaderBuilder().
		WithSalt(salt).
		WithOriginalSize(uint64(fileInfo.Size())).
		WithAesNonce(worker.GetAesNonce()).
		WithChaCha20Nonce(worker.GetChaCha20Nonce()).
		Build()
	if err != nil {
		return fmt.Errorf("header building failed: %w", err)
	}

	if err = header.NewHeaderWriter(header.NewBinaryHeaderIO()).Write(output, builder); err != nil {
		return fmt.Errorf("header writing failed: %w", err)
	}

	if err = worker.Process(input, output, fileInfo.Size()); err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	return nil
}

func (op *Operations) performDecryption(input *os.File, output *os.File, key []byte, fileHeader header.Header) error {
	worker, err := worker.NewFileProcessor(key, false)
	if err != nil {
		return fmt.Errorf("decryption processor creation failed: %w", err)
	}

	if err := worker.SetAesNonce(fileHeader.AesNonce.Value); err != nil {
		return fmt.Errorf("AES nonce setting failed: %w", err)
	}

	if err := worker.SetChaCha20Nonce(fileHeader.ChaCha20Nonce.Value); err != nil {
		return fmt.Errorf("ChaCha20 nonce setting failed: %w", err)
	}

	if err := worker.Process(input, output, int64(fileHeader.OriginalSize.Value)); err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	return nil
}
