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
	if err := validateInputFile(inputFile); err != nil {
		return err
	}

	outputFile := strings.TrimSuffix(inputFile, ".enc")

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

	password, err := promptPassword()
	if err != nil {
		return fmt.Errorf("failed to ask for password: %w", err)
	}

	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	fileHeader, err := header.Read(input)
	if err != nil {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	key, err := kdf.Derive([]byte(password), fileHeader.Salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	output, err := prepareOutputFile(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	decrypt, err := decryptor.NewFileDecryptor(key)
	if err != nil {
		return fmt.Errorf("failed to create decryptor: %w", err)
	}

	err = decrypt.SetNonce(fileHeader.SerpentNonce, fileHeader.ChaCha20Nonce)
	if err != nil {
		return fmt.Errorf("failed to set nonce: %w", err)
	}

	fmt.Printf("Decrypting %s...\n", inputFile)
	if err = decrypt.Decrypt(input, output, int64(fileHeader.OriginalSize)); err != nil {
		output.Close()
		os.Remove(outputFile)
		return fmt.Errorf("failed to decrypt file: %w", err)
	}

	deleteEncrypted, deleteType, err := promptDeleteEncrypted(inputFile)
	if err != nil {
		return fmt.Errorf("failed to ask for delete encrypted: %w", err)
	}

	if deleteEncrypted {
		if err := deleteEncryptedFile(inputFile, deleteType); err != nil {
			return fmt.Errorf("failed to delete encrypted file: %w", err)
		}
	}

	fmt.Printf("File %s decrypted successfully\n", outputFile)
	return nil
}
