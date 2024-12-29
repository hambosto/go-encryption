package header

import (
	"fmt"
)

type Salt struct {
	Value []byte
}

func (s Salt) Size() int { return SaltSize }
func (s Salt) Validate(data []byte) error {
	if len(data) != SaltSize {
		return fmt.Errorf("invalid salt size: got %d, want %d", len(data), SaltSize)
	}
	return nil
}

type OriginalSize struct {
	Value uint64
}

func (s OriginalSize) Size() int { return OriginalSizeBytes }
func (s OriginalSize) Validate(data []byte) error {
	if len(data) != OriginalSizeBytes {
		return fmt.Errorf("invalid size bytes: got %d, want %d", len(data), OriginalSizeBytes)
	}
	return nil
}

type AesNonce struct {
	Value []byte
}

func (n AesNonce) Size() int { return AesNonceSize }
func (n AesNonce) Validate(data []byte) error {
	if len(data) != AesNonceSize {
		return fmt.Errorf("invalid AES nonce size: got %d, want %d", len(data), AesNonceSize)
	}
	return nil
}

type ChaCha20Nonce struct {
	Value []byte
}

func (n ChaCha20Nonce) Size() int { return ChaCha20NonceSize }
func (n ChaCha20Nonce) Validate(data []byte) error {
	if len(data) != ChaCha20NonceSize {
		return fmt.Errorf("invalid ChaCha20 nonce size: got %d, want %d", len(data), ChaCha20NonceSize)
	}
	return nil
}
