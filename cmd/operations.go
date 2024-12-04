package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hambosto/go-encryption/internal/decryptor"
	"github.com/hambosto/go-encryption/internal/encryptor"
	"github.com/hambosto/go-encryption/internal/header"
	"github.com/hambosto/go-encryption/internal/kdf"
	"github.com/hambosto/go-encryption/internal/utils"
)

func RunEncryption(inputFile string) error {
	outputFile := inputFile + ".enc"

	if err := utils.ValidateInputFile(inputFile); err != nil {
		return err
	}

	if err := utils.ValidateOutputFile(outputFile); err != nil {
		overwrite, err := utils.PromptOverwrite(outputFile)
		if err != nil {
			return fmt.Errorf("failed to ask for overwrite: %w", err)
		}
		if !overwrite {
			return fmt.Errorf("encryption cancelled")
		}
	}

	password, err := utils.PromptPassword()
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

	output, err := utils.PrepareOutputFile(outputFile)
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

	deleteOriginal, deleteType, err := utils.PromptDeleteOriginal(inputFile)
	if err != nil {
		return fmt.Errorf("failed to ask for delete original: %w", err)
	}

	if deleteOriginal {
		if err := utils.DeleteOriginalFile(inputFile, deleteType); err != nil {
			return fmt.Errorf("failed to delete original file: %w", err)
		}
	}

	fmt.Printf("File %s encrypted successfully\n", outputFile)
	return nil
}

func RunDecryption(inputFile string) error {
	if err := utils.ValidateInputFile(inputFile); err != nil {
		return err
	}

	outputFile := strings.TrimSuffix(inputFile, ".enc")

	if err := utils.ValidateOutputFile(outputFile); err != nil {
		overwrite, err := utils.PromptOverwrite(outputFile)
		if err != nil {
			return fmt.Errorf("failed to ask for overwrite: %w", err)
		}
		if !overwrite {
			return fmt.Errorf("decryption cancelled")
		}
	}

	password, err := utils.PromptPassword()
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

	output, err := utils.PrepareOutputFile(outputFile)
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

	deleteEncrypted, deleteType, err := utils.PromptDeleteEncrypted(inputFile)
	if err != nil {
		return fmt.Errorf("failed to ask for delete encrypted: %w", err)
	}

	if deleteEncrypted {
		if err := utils.DeleteEncryptedFile(inputFile, deleteType); err != nil {
			return fmt.Errorf("failed to delete encrypted file: %w", err)
		}
	}

	fmt.Printf("File %s decrypted successfully\n", outputFile)
	return nil
}
