package decryption

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hambosto/go-encryption/internal/decryptor"
	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
	"github.com/hambosto/go-encryption/internal/utils"
)

func RunDecryption(inputFile string) error {
	if !strings.HasSuffix(inputFile, ".enc") {
		return fmt.Errorf("input file must have .enc extension")
	}

	outputFile := strings.TrimSuffix(inputFile, ".enc")

	var err error

	_, err = os.Stat(inputFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist")
	}

	_, err = os.Stat(outputFile)
	if err == nil {
		overwrite := false
		promptOverwrite := &survey.Confirm{
			Message: fmt.Sprintf("Output file %s already exists. Overwrite?", outputFile),
			Default: false,
		}
		err = survey.AskOne(promptOverwrite, &overwrite)
		if err != nil {
			return fmt.Errorf("failed to ask for overwrite: %w", err)
		}

		if !overwrite {
			return fmt.Errorf("decryption cancelled")
		}
	}

	password := ""
	promptPassword := &survey.Password{
		Message: "Enter password:",
	}
	err = survey.AskOne(promptPassword, &password)
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

	outputDir := filepath.Dir(outputFile)
	if err = os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
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

	deleteEncrypted := false
	promptDeleteEncrypted := &survey.Confirm{
		Message: fmt.Sprintf("Delete encrypted file %s?", inputFile),
		Default: false,
	}
	err = survey.AskOne(promptDeleteEncrypted, &deleteEncrypted)
	if err != nil {
		return fmt.Errorf("failed to ask for delete encrypted: %w", err)
	}

	if deleteEncrypted {
		deleteType := ""
		promptDeleteType := &survey.Select{
			Message: "Select delete type:",
			Options: []string{
				"Normal delete (faster, but recoverable)",
				"Secure delete (slower, but unrecoverable)",
			},
			Default: "Normal delete (faster, but recoverable)",
		}
		err = survey.AskOne(promptDeleteType, &deleteType)
		if err != nil {
			return fmt.Errorf("failed to ask for delete type: %w", err)
		}

		switch deleteType {
		case "Normal delete (faster, but recoverable)":
			if err := os.Remove(inputFile); err != nil {
				return fmt.Errorf("failed to delete encrypted file: %w", err)
			}
		case "Secure delete (slower, but unrecoverable)":
			if err := utils.SecureDelete(inputFile); err != nil {
				return fmt.Errorf("failed to securely delete encrypted file: %w", err)
			}
		}
	}

	fmt.Printf("File %s decrypted successfully\n", outputFile)
	return nil
}
