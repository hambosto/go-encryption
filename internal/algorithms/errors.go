package algorithms

import "errors"

var (
	ErrEmptyPlaintext        = errors.New("plaintext cannot be empty")
	ErrInvalidKeySize        = errors.New("invalid key size")
	ErrInvalidNonceSize      = errors.New("invalid nonce size")
	ErrUnsupportedAlgorithm  = errors.New("unsupported encryption algorithm")
	ErrCipherCreationFailed  = errors.New("failed to create cipher")
	ErrNonceGenerationFailed = errors.New("failed to generate nonce")
	ErrDecryptionFailed      = errors.New("decryption failed")
	ErrCiphertextTooShort    = errors.New("ciphertext too short")
)
