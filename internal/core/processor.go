package core

import (
	"errors"
	"os"

	"github.com/hambosto/go-encryption/internal/crypto"
	"github.com/hambosto/go-encryption/internal/crypto/compression"
	"github.com/hambosto/go-encryption/internal/crypto/encoding"
	"github.com/hambosto/go-encryption/internal/crypto/padding"
	"github.com/hambosto/go-encryption/internal/progress"
)

// FileProcessor handles the encryption and decryption of files.
type FileProcessor struct {
	crypto   *crypto.CryptoService
	progress *progress.ProgressReporter
}

// NewFileProcessor creates a new instance of FileProcessor with the provided crypto service and progress reporter.
func NewFileProcessor(crypto *crypto.CryptoService, progress *progress.ProgressReporter) *FileProcessor {
	return &FileProcessor{
		crypto:   crypto,
		progress: progress,
	}
}

// Encrypt encrypts the data from the input file and writes it to the output file using the provided password.
// It performs compression, key derivation, padding, encryption, and error correction.
func (fp *FileProcessor) Encrypt(inputPath, outputPath, password string) error {
	// Read input file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	fp.progress.Start("Starting encryption process")

	// Compress data
	fp.progress.Update("Compressing file data", 0.2)
	compressedData, err := compression.ZlibCompress(data)
	if err != nil {
		return err
	}

	// Generate salt
	fp.progress.Update("Generating encryption salt", 0.3)
	salt, err := fp.crypto.GenerateSalt()
	if err != nil {
		return err
	}

	// Derive encryption key from password and salt
	fp.progress.Update("Deriving encryption key from password", 0.4)
	key, err := fp.crypto.DeriveKey(password, salt)
	if err != nil {
		return err
	}

	// Pad the compressed data
	fp.progress.Update("Padding compressed data for encryption", 0.5)
	paddedData, err := padding.Pad(compressedData, 16)
	if err != nil {
		return err
	}

	// Encrypt the padded data
	fp.progress.Update("Encrypting the padded data", 0.6)
	encryptedData, err := fp.crypto.EncryptData(paddedData, key)
	if err != nil {
		return err
	}

	// Initialize Reed-Solomon encoding for error correction
	fp.progress.Update("Initializing Reed-Solomon encoding", 0.7)
	encoding, err := encoding.NewReedSolomonCodec(encoding.Config{
		DataShards:   fp.crypto.RequiredShares,
		ParityShards: fp.crypto.TotalShares,
	})
	if err != nil {
		return err
	}

	// Apply Reed-Solomon encoding to the encrypted data
	fp.progress.Update("Applying Reed-Solomon error correction", 0.8)
	encodedData, err := encoding.Encode(encryptedData)
	if err != nil {
		return err
	}

	// Write the salt and encoded data to the output file
	fp.progress.Update("Writing encrypted file to disk", 0.9)
	if err := fp.writeEncryptedFile(outputPath, salt, encodedData); err != nil {
		return err
	}

	fp.progress.Complete("Encryption successfully completed")
	return nil
}

// Decrypt decrypts the data from the input file and writes the original data to the output file using the provided password.
// It performs error correction, key derivation, decryption, unpadding, and decompression.
func (fp *FileProcessor) Decrypt(inputPath, outputPath, password string) error {
	// Read encrypted file
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	fp.progress.Start("Starting decryption process")

	// Extract salt and encoded data from the encrypted file
	if len(data) < fp.crypto.SaltSize {
		return errors.New("invalid encrypted file")
	}
	salt := data[:fp.crypto.SaltSize]
	encodedData := data[fp.crypto.SaltSize:]

	// Derive decryption key from password and salt
	fp.progress.Update("Deriving decryption key from password", 0.2)
	key, err := fp.crypto.DeriveKey(password, salt)
	if err != nil {
		return err
	}

	// Initialize Reed-Solomon decoding for error correction
	fp.progress.Update("Initializing Reed-Solomon decoding", 0.3)
	encoding, err := encoding.NewReedSolomonCodec(encoding.Config{
		DataShards:   fp.crypto.RequiredShares,
		ParityShards: fp.crypto.TotalShares,
	})
	if err != nil {
		return err
	}

	// Apply Reed-Solomon decoding to the encoded data
	fp.progress.Update("Applying Reed-Solomon decoding", 0.4)
	decodedData, err := encoding.Decode(encodedData)
	if err != nil {
		return err
	}

	// Decrypt the data
	fp.progress.Update("Decrypting the encoded data", 0.6)
	decryptedData, err := fp.crypto.DecryptData(decodedData, key)
	if err != nil {
		return err
	}

	// Remove padding from decrypted data
	fp.progress.Update("Removing padding from decrypted data", 0.7)
	unpaddedData, err := padding.Unpad(decryptedData, 16)
	if err != nil {
		return err
	}

	// Decompress the decrypted data
	fp.progress.Update("Decompressing decrypted data", 0.8)
	decompressedData, err := compression.ZlibDecompress(unpaddedData)
	if err != nil {
		return err
	}

	// Write the decompressed data to the output file
	fp.progress.Update("Writing decrypted file to disk", 0.9)
	if err := os.WriteFile(outputPath, decompressedData, 0o0600); err != nil {
		return err
	}

	fp.progress.Complete("Decryption successfully completed")
	return nil
}

// writeEncryptedFile writes the salt and encoded data to the specified file path.
func (fp *FileProcessor) writeEncryptedFile(path string, salt []byte, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the salt to the file
	if _, err := f.Write(salt); err != nil {
		return err
	}
	// Write the encoded data to the file
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}
