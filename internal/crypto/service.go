package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"errors"
	"io"

	"github.com/aead/serpent"
	"github.com/hambosto/go-encryption/internal/crypto/random"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"
)

// CryptoService provides encryption and decryption functionalities.
// It contains settings for key derivation and share requirements.
type CryptoService struct {
	SaltSize       int    // Size of the salt in bytes
	KeySize        int    // Size of the derived key in bytes
	NonceSize      int    // Size of the nonce for encryption
	MemoryCost     uint32 // Memory cost for Argon2 key derivation
	TimeCost       uint32 // Time cost for Argon2 key derivation
	Parallelism    uint8  // Parallelism for Argon2 key derivation
	RequiredShares int    // Required shares for Reed-Solomon encoding
	TotalShares    int    // Total shares for Reed-Solomon encoding
}

// NewCryptoService creates a new instance of CryptoService with default settings.
func NewCryptoService() *CryptoService {
	return &CryptoService{
		SaltSize:       32,
		KeySize:        32,
		NonceSize:      24,
		MemoryCost:     64 * 1024, // 64MB
		TimeCost:       3,
		Parallelism:    4,
		RequiredShares: 16,
		TotalShares:    48,
	}
}

// DeriveKey derives a cryptographic key from a password and salt using Argon2 and HKDF.
// It returns the derived key or an error if the process fails.
func (cs *CryptoService) DeriveKey(password string, salt []byte) ([]byte, error) {
	// Use Argon2id for password hashing
	key := argon2.IDKey(
		[]byte(password),
		salt,
		cs.TimeCost,
		cs.MemoryCost,
		cs.Parallelism,
		uint32(cs.KeySize),
	)

	// Use HKDF to expand the key
	h := hkdf.New(sha3.New256, key, salt, []byte("secure-encrypt-v1"))
	finalKey := make([]byte, cs.KeySize)
	if _, err := io.ReadFull(h, finalKey); err != nil {
		return nil, err
	}

	return finalKey, nil
}

// GenerateSalt generates a random salt of the configured size.
// It returns the generated salt or an error if the process fails.
func (cs *CryptoService) GenerateSalt() ([]byte, error) {
	salt := make([]byte, cs.KeySize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return random.Generate(cs.KeySize)
}

// EncryptData encrypts the given data using a combination of ChaCha20 and Serpent encryption.
// It returns the encrypted data along with the nonce or an error if the process fails.
func (cs *CryptoService) EncryptData(data []byte, key []byte) ([]byte, error) {
	// Generate nonce
	nonce, err := random.Generate(cs.NonceSize)
	if err != nil {
		return nil, err
	}

	// First layer: ChaCha20 encryption
	chacha, err := chacha20.NewUnauthenticatedCipher(key[:32], nonce)
	if err != nil {
		return nil, err
	}

	chachaEncrypted := make([]byte, len(data))
	chacha.XORKeyStream(chachaEncrypted, data)

	// Second layer: Serpent encryption
	serpentBlock, err := serpent.NewCipher(key)
	if err != nil {
		return nil, err
	}

	serpentEncrypted := make([]byte, len(chachaEncrypted))
	for i := 0; i < len(chachaEncrypted); i += serpent.BlockSize {
		end := i + serpent.BlockSize
		if end > len(chachaEncrypted) {
			end = len(chachaEncrypted)
		}
		serpentBlock.Encrypt(serpentEncrypted[i:end], chachaEncrypted[i:end])
	}

	// Compute BLAKE2b hash for data integrity
	hash, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	hash.Write(serpentEncrypted)
	checksum := hash.Sum(nil)

	// Combine nonce, encrypted data, and checksum
	result := make([]byte, cs.NonceSize+len(serpentEncrypted)+32)
	copy(result[:cs.NonceSize], nonce)
	copy(result[cs.NonceSize:], serpentEncrypted)
	copy(result[cs.NonceSize+len(serpentEncrypted):], checksum)

	return result, nil
}

// DecryptData decrypts the given encrypted data using the specified key.
// It returns the original plaintext data or an error if the process fails.
func (cs *CryptoService) DecryptData(data []byte, key []byte) ([]byte, error) {
	if len(data) < cs.NonceSize+32 {
		return nil, errors.New("encrypted data too short")
	}

	// Extract components from the encrypted data
	nonce := data[:cs.NonceSize]
	encryptedData := data[cs.NonceSize : len(data)-32]
	storedChecksum := data[len(data)-32:]

	// Verify checksum to ensure data integrity
	hash, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	hash.Write(encryptedData)
	calculatedChecksum := hash.Sum(nil)
	if !hmac.Equal(calculatedChecksum, storedChecksum) {
		return nil, errors.New("data integrity check failed")
	}

	// First layer: Serpent decryption
	serpentBlock, err := serpent.NewCipher(key)
	if err != nil {
		return nil, err
	}

	serpentDecrypted := make([]byte, len(encryptedData))
	for i := 0; i < len(encryptedData); i += serpent.BlockSize {
		end := i + serpent.BlockSize
		if end > len(encryptedData) {
			end = len(encryptedData)
		}
		serpentBlock.Decrypt(serpentDecrypted[i:end], encryptedData[i:end])
	}

	// Second layer: ChaCha20 decryption
	chacha, err := chacha20.NewUnauthenticatedCipher(key[:32], nonce)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(serpentDecrypted))
	chacha.XORKeyStream(plaintext, serpentDecrypted)

	return plaintext, nil
}

