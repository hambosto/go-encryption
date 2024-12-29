package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
	"github.com/hambosto/go-encryption/internal/operations"
	"github.com/hambosto/go-encryption/internal/utils"
)

type FileOperationConfig struct {
	InputFile  string
	OutputFile string
	Password   string
}

func validateFileOperation(config FileOperationConfig, isEncryption bool) error {
	if err := utils.ValidateInputFile(config.InputFile); err != nil {
		return err
	}

	if err := utils.ValidateOutputFile(config.OutputFile); err != nil {
		overwrite, err := utils.PromptOverwrite(config.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to ask for overwrite: %w", err)
		}
		if !overwrite {
			operation := "encryption"
			if !isEncryption {
				operation = "decryption"
			}
			return fmt.Errorf("%s cancelled", operation)
		}
	}

	return nil
}

func prepareInputFile(inputFile string) (*os.File, os.FileInfo, error) {
	input, err := os.Open(inputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open input file: %w", err)
	}

	fileInfo, err := input.Stat()
	if err != nil {
		input.Close()
		return nil, nil, fmt.Errorf("failed to get input file info: %w", err)
	}

	return input, fileInfo, nil
}

func deriveEncryptionKey(password string) ([]byte, []byte, error) {
	salt, err := kdf.GenerateSalt()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key, err := kdf.Derive([]byte(password), salt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive key: %w", err)
	}

	return key, salt, nil
}

func performEncryption(input *os.File, output *os.File, fileInfo os.FileInfo, key []byte, salt []byte) error {
	operations, err := operations.NewFileProcessor(key, true)
	if err != nil {
		return fmt.Errorf("failed to create encryptor: %w", err)
	}

	// aesNonce, chaCha20Nonce := operations.GetNonce()

	aesNonce := operations.GetAesNonce()
	chaCha20Nonce := operations.GetChaCha20Nonce()

	builder, err := header.NewHeaderBuilder().
		WithSalt(salt).
		WithOriginalSize(uint64(fileInfo.Size())).
		WithAesNonce(aesNonce).
		WithChaCha20Nonce(chaCha20Nonce).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build header: %w", err)
	}

	writer := header.NewHeaderWriter(header.NewBinaryHeaderIO())
	if err = writer.Write(output, builder); err != nil {
		return fmt.Errorf("failed to write file header: %w", err)
	}

	if err = operations.Process(input, output, fileInfo.Size()); err != nil {
		return fmt.Errorf("failed to encrypt file: %w", err)
	}

	return nil
}

func performDecryption(input *os.File, output *os.File, key []byte, fileHeader header.Header) error {
	operations, err := operations.NewFileProcessor(key, false)
	if err != nil {
		return fmt.Errorf("failed to create decryptor: %w", err)
	}

	if operations.SetAesNonce(fileHeader.AesNonce.Value); err != nil {
		return fmt.Errorf("failed to set AES nonce: %w", err)
	}

	if operations.SetChaCha20Nonce(fileHeader.ChaCha20Nonce.Value); err != nil {
		return fmt.Errorf("failed to set ChaCha20 nonce: %w", err)
	}

	if err = operations.Process(input, output, int64(fileHeader.OriginalSize.Value)); err != nil {
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	return nil
}

func handleFileCleanup(inputFile string, isEncryption bool) error {
	var deleteFile bool
	var deleteType string
	var err error

	if isEncryption {
		deleteFile, deleteType, err = utils.PromptDeleteOriginal(inputFile)
	} else {
		deleteFile, deleteType, err = utils.PromptDeleteEncrypted(inputFile)
	}

	if err != nil {
		return fmt.Errorf("failed to ask about file deletion: %w", err)
	}

	if deleteFile {
		var deleteFunc func(string, string) error
		if isEncryption {
			deleteFunc = utils.DeleteOriginalFile
		} else {
			deleteFunc = utils.DeleteEncryptedFile
		}

		if err := deleteFunc(inputFile, deleteType); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	return nil
}

func RunEncryption(inputFile string) error {
	outputFile := inputFile + ".enc"

	config := FileOperationConfig{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}
	if err := validateFileOperation(config, true); err != nil {
		return err
	}

	password, err := utils.PromptPassword()
	if err != nil {
		return err
	}

	input, fileInfo, err := prepareInputFile(inputFile)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := utils.PrepareOutputFile(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	key, salt, err := deriveEncryptionKey(password)
	if err != nil {
		return err
	}

	fmt.Printf("Encrypting %s...\n", inputFile)

	if err = performEncryption(input, output, fileInfo, key, salt); err != nil {
		output.Close()
		os.Remove(outputFile)
		return err
	}

	if err = handleFileCleanup(inputFile, true); err != nil {
		return err
	}

	fmt.Printf("File %s encrypted successfully\n", outputFile)
	return nil
}

func RunDecryption(inputFile string) error {
	outputFile := strings.TrimSuffix(inputFile, ".enc")

	config := FileOperationConfig{
		InputFile:  inputFile,
		OutputFile: outputFile,
	}
	if err := validateFileOperation(config, false); err != nil {
		return err
	}

	password, err := utils.PromptPassword()
	if err != nil {
		return fmt.Errorf("failed to ask for password: %w", err)
	}

	input, _, err := prepareInputFile(inputFile)
	if err != nil {
		return err
	}
	defer input.Close()

	reader := header.NewHeaderReader(header.NewBinaryHeaderIO())
	fileHeader, err := reader.Read(input)
	if err != nil {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	key, err := kdf.Derive([]byte(password), fileHeader.Salt.Value)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	output, err := utils.PrepareOutputFile(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	fmt.Printf("Decrypting %s...\n", inputFile)

	if err = performDecryption(input, output, key, fileHeader); err != nil {
		output.Close()
		os.Remove(outputFile)
		return err
	}

	if err = handleFileCleanup(inputFile, false); err != nil {
		return err
	}

	fmt.Printf("File %s decrypted successfully\n", outputFile)
	return nil
}
