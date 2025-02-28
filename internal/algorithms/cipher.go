package algorithms

type Cipher interface {
	// Encrypt encrypts the plaintext and returns the ciphertext
	Encrypt(plaintext []byte) ([]byte, error)
	// Decrypt decrypts the ciphertext and returns the plaintext
	Decrypt(ciphertext []byte) ([]byte, error)
	// SetNonce sets the nonce for the cipher
	SetNonce(nonce []byte) error
	// GetNonce returns the current nonce
	GetNonce() []byte
}

type Algorithm string

const (
	AES      Algorithm = "AES"
	CHACHA20 Algorithm = "CHACHA20"
)

func NewCipher(algorithm Algorithm, key []byte) (Cipher, error) {
	switch algorithm {
	case AES:
		return NewAESCipher(key)
	case CHACHA20:
		return NewChaCha20Cipher(key)
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}
