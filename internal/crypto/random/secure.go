package random

import (
	"crypto/rand"
	"encoding/binary"
	"io"
)

// Generate generates cryptographically secure random bytes
func Generate(size int) ([]byte, error) {
	bytes := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// GenerateInt64 generates a cryptographically secure random int64
func GenerateInt64() (int64, error) {
	var buf [8]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(buf[:])), nil
}

// GenerateUint64 generates a cryptographically secure random uint64
func GenerateUint64() (uint64, error) {
	var buf [8]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(buf[:]), nil
}

// Shuffle cryptographically securely shuffles a byte slice in place
func Shuffle(data []byte) error {
	n := len(data)
	if n <= 1 {
		return nil
	}

	for i := n - 1; i > 0; i-- {
		randInt, err := GenerateInt64()
		if err != nil {
			return err
		}
		j := int(uint(randInt) % uint(i+1))
		data[i], data[j] = data[j], data[i]
	}
	return nil
}
