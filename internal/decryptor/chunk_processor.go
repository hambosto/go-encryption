package decryptor

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/encoding"
)

const (
	MaxChunkSize          = 1024 * 1024
	EncryptionOverhead    = (MaxChunkSize / 10) + 16 + 4
	Padding               = 4 - 1
	MaxEncryptedChunkSize = ((MaxChunkSize + EncryptionOverhead + Padding) / 4) * (4 + 10)
)

type ChunkProcessor struct {
	aesCipher      *algorithms.AESCipher
	chaCha20Cipher *algorithms.ChaCha20Cipher
	encoder        *encoding.Encoder
	bufferPool     sync.Pool
	decompressPool sync.Pool
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
				buffer := make([]byte, MaxEncryptedChunkSize)
				return &buffer
			},
		},
		decompressPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}, nil
}

func (cp *ChunkProcessor) ProcessChunk(chunk []byte) ([]byte, error) {
	decodedData, err := cp.encoder.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("reed-solomon decoding failed: %w", err)
	}

	chaCha20Decrypted, err := cp.chaCha20Cipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 decryption failed: %w", err)
	}

	aesDecrypted, err := cp.aesCipher.Decrypt(chaCha20Decrypted)
	if err != nil {
		return nil, fmt.Errorf("aes decryption failed: %w", err)
	}

	return aesDecrypted, nil
}

func (cp *ChunkProcessor) DecompressData(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid data: insufficient bytes for size information")
	}

	compressedSize := binary.BigEndian.Uint32(data[:4])
	if compressedSize > uint32(len(data)-4) {
		return nil, fmt.Errorf("invalid compressed data size: expected %d, got %d", compressedSize, len(data)-4)
	}

	compressedData := data[4 : 4+compressedSize]

	buffer := cp.decompressPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer cp.decompressPool.Put(buffer)

	zr, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zr.Close()

	if _, err := io.Copy(buffer, zr); err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	result := make([]byte, buffer.Len())
	copy(result, buffer.Bytes())

	return result, nil
}
