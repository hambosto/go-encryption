package decryption

import (
	"fmt"
	"os"
	"strings"

	"github.com/hambosto/go-encryption/internal/decryptor"
	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
)

func RunDecryption(inputFile string) error {
	// Validate input file
	if err := validateInputFile(inputFile); err != nil {
		return err
	}

	// Prepare output file name
	outputFile := strings.TrimSuffix(inputFile, ".enc")

	// Validate output file
	if err := validateOutputFile(outputFile); err != nil {
		// Prompt for overwrite if file exists
		overwrite, err := promptOverwrite(outputFile)
		if err != nil {
			return fmt.Errorf("failed to ask for overwrite: %w", err)
		}
		if !overwrite {
			return fmt.Errorf("decryption cancelled")
		}
	}

	// Prompt for password
	password, err := promptPassword()
	if err != nil {
		return fmt.Errorf("failed to ask for password: %w", err)
	}

	// Open input file
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	// Read file header
	fileHeader, err := header.Read(input)
	if err != nil {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	// Derive key
	key, err := kdf.Derive([]byte(password), fileHeader.Salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	// Prepare output file
	output, err := prepareOutputFile(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	// Create decryptor
	decrypt, err := decryptor.NewFileDecryptor(key)
	if err != nil {
		return fmt.Errorf("failed to create decryptor: %w", err)
	}

	// Set nonce
	err = decrypt.SetNonce(fileHeader.SerpentNonce, fileHeader.ChaCha20Nonce)
	if err != nil {
		return fmt.Errorf("failed to set nonce: %w", err)
	}

	// Decrypt file
	fmt.Printf("Decrypting %s...\n", inputFile)
	if err = decrypt.Decrypt(input, output, int64(fileHeader.OriginalSize)); err != nil {
		output.Close()
		os.Remove(outputFile)
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	// Prompt for encrypted file deletion
	deleteEncrypted, deleteType, err := promptDeleteEncrypted(inputFile)
	if err != nil {
		return fmt.Errorf("failed to ask for delete encrypted: %w", err)
	}

	// Delete encrypted file if requested
	if deleteEncrypted {
		if err := deleteEncryptedFile(inputFile, deleteType); err != nil {
			return fmt.Errorf("failed to delete encrypted file: %w", err)
		}
	}

	fmt.Printf("File %s decrypted successfully\n", outputFile)
	return nil
}
