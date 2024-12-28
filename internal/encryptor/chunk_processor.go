package encryptor

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/encoding"
)

const (
	MaxChunkSize = 1024 * 1024
)

type ChunkProcessor struct {
	aesCipher      *algorithms.AESCipher
	chaCha20Cipher *algorithms.ChaCha20Cipher
	encoder        *encoding.Encoder
	bufferPool     sync.Pool
	compressPool   sync.Pool
}

func NewChunkProcessor(key []byte) (*ChunkProcessor, error) {
	if len(key) < 64 {
		return nil, fmt.Errorf("encryption key must be at least 64 bytes long")
	}

	aesCipher, err := algorithms.NewAESCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %w", err)
	}

	chaCha20Cipher, err := algorithms.NewChaCha20Cipher(key[32:64])
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20 cipher: %w", err)
	}

	encoder, err := encoding.New(4, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	return &ChunkProcessor{
		aesCipher:      aesCipher,
		chaCha20Cipher: chaCha20Cipher,
		encoder:        encoder,
		bufferPool: sync.Pool{
			New: func() interface{} {
				buffer := make([]byte, MaxChunkSize)
				return &buffer
			},
		},
		compressPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}, nil
}

func (cp *ChunkProcessor) ProcessChunk(chunk []byte) ([]byte, error) {
	compressedData, err := cp.compressData(chunk)
	if err != nil {
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	paddedData := cp.padData(compressedData)

	aesEncrypted, err := cp.aesCipher.Encrypt(paddedData)
	if err != nil {
		return nil, fmt.Errorf("aes encryption failed: %w", err)
	}

	chaCha20Encrypted, err := cp.chaCha20Cipher.Encrypt(aesEncrypted)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 encryption failed: %w", err)
	}

	encoder, err := cp.encoder.Encode(chaCha20Encrypted)
	if err != nil {
		return nil, fmt.Errorf("Reed-Solomon encoding failed: %w", err)
	}

	return encoder, nil
}

func (cp *ChunkProcessor) compressData(data []byte) ([]byte, error) {
	buffer := cp.compressPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer cp.compressPool.Put(buffer)

	zw, err := zlib.NewWriterLevel(buffer, zlib.BestSpeed)
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib writer: %w", err)
	}

	if _, err := zw.Write(data); err != nil {
		zw.Close()
		return nil, fmt.Errorf("failed to write data to zlib writer: %w", err)
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zlib writer: %w", err)
	}

	result := make([]byte, buffer.Len())
	copy(result, buffer.Bytes())

	return result, nil
}

func (cp *ChunkProcessor) padData(data []byte) []byte {
	sizeHeader := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeHeader, uint32(len(data)))

	alignedSize := (len(data) + 15) & ^15
	if alignedSize > len(data) {
		padding := make([]byte, alignedSize-len(data))
		data = append(data, padding...)
	}

	return append(sizeHeader, data...)
}
