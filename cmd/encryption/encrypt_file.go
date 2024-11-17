package encryption

import (
	"bytes"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hambosto/go-encryption/internal/encryptor"
	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
	"github.com/hambosto/go-encryption/internal/utils"
)

func RunEncryption(inputFile string) error {
	outputFile := inputFile + ".enc"

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
			return fmt.Errorf("encryption cancelled")
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

	confirmPassword := ""
	promptConfirmPassword := &survey.Password{
		Message: "Confirm password:",
	}
	err = survey.AskOne(promptConfirmPassword, &confirmPassword)
	if err != nil {
		return fmt.Errorf("failed to ask for password confirmation: %w", err)
	}

	if !bytes.Equal([]byte(password), []byte(confirmPassword)) {
		return fmt.Errorf("passwords do not match")
	}

	salt, err := kdf.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	key, err := kdf.Derive([]byte(password), salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer input.Close()

	fileInfo, err := input.Stat()
	if err != nil {
		return fmt.Errorf("failed to get input file info: %w", err)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("input file is empty")
	}

	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer output.Close()

	encrypt, err := encryptor.NewFileEncryptor(key)
	if err != nil {
		return fmt.Errorf("failed to create encryptor: %w", err)
	}

	serpentNonce, chaCha20Nonce := encrypt.GetNonce()

	fileHeader := header.FileHeader{
		Salt:          salt,
		OriginalSize:  uint64(fileInfo.Size()),
		SerpentNonce:  serpentNonce,
		ChaCha20Nonce: chaCha20Nonce,
	}

	err = header.Write(output, fileHeader)
	if err != nil {
		return fmt.Errorf("failed to write file header: %w", err)
	}

	fmt.Printf("Encrypting %s...\n", inputFile)
	if err := encrypt.Encrypt(input, output, fileInfo.Size()); err != nil {
		output.Close()
		os.Remove(outputFile)
		return fmt.Errorf("failed to encrypt file: %w", err)
	}

	deleteOriginal := false
	promptDeleteOriginal := &survey.Confirm{
		Message: fmt.Sprintf("Delete original file %s?", inputFile),
		Default: false,
	}
	err = survey.AskOne(promptDeleteOriginal, &deleteOriginal)
	if err != nil {
		return fmt.Errorf("failed to ask for delete original: %w", err)
	}

	if deleteOriginal {
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
				return fmt.Errorf("failed to delete original file: %w", err)
			}
		case "Secure delete (slower, but unrecoverable)":
			if err := utils.SecureDelete(inputFile); err != nil {
				return fmt.Errorf("failed to securely delete original file: %w", err)
			}
		}
	}

	fmt.Printf("File %s encrypted successfully\n", outputFile)
	return nil
}
