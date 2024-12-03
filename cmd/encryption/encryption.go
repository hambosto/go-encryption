package encryption

import (
	"fmt"
	"os"

	"github.com/hambosto/go-encryption/internal/encryptor"
	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
)

func RunEncryption(inputFile string) error {
	outputFile := inputFile + ".enc"

	if err := validateInputFile(inputFile); err != nil {
		return err
	}

	if err := validateOutputFile(outputFile); err != nil {
		overwrite, err := promptOverwrite(outputFile)
		if err != nil {
			return fmt.Errorf("failed to ask for overwrite: %w", err)
		}
		if !overwrite {
			return fmt.Errorf("encryption cancelled")
		}
	}

	password, err := promptPassword()
	if err != nil {
		return err
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

	output, err := prepareOutputFile(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	salt, err := kdf.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	key, err := kdf.Derive([]byte(password), salt)
	if err != nil {
		return fmt.Errorf("failed to derive key: %w", err)
	}

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
	if err = encrypt.Encrypt(input, output, fileInfo.Size()); err != nil {
		output.Close()
		os.Remove(outputFile)
		return fmt.Errorf("failed to encrypt file: %w", err)
	}

	deleteOriginal, deleteType, err := promptDeleteOriginal(inputFile)
	if err != nil {
		return fmt.Errorf("failed to ask for delete original: %w", err)
	}

	if deleteOriginal {
		if err := deleteOriginalFile(inputFile, deleteType); err != nil {
			return fmt.Errorf("failed to delete original file: %w", err)
		}
	}

	fmt.Printf("File %s encrypted successfully\n", outputFile)
	return nil
}
