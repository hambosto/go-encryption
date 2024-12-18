package decryptor

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/config"
	"github.com/hambosto/go-encryption/internal/encoding"
)

const (
	MaxChunkSize          = 1024 * 1024
	MaxEncryptedChunkSize = ((MaxChunkSize + (MaxChunkSize / 10) + 16 + 4 + (4 - 1)) / 4) * (4 + 10)
)

type ChunkProcessor struct {
	serpentCipher  *algorithms.SerpentCipher
	chaCha20Cipher *algorithms.ChaCha20Cipher
	rsDecoder      *encoding.ReedSolomonEncoder
	bufferPool     sync.Pool
	decompressPool sync.Pool
}

func NewChunkProcessor(key []byte) (*ChunkProcessor, error) {
	if len(key) < 64 {
		return nil, fmt.Errorf("encryption key must be at least 64 bytes long")
	}

	serpentCipher, err := algorithms.NewSerpentCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create Serpent cipher: %w", err)
	}

	chaCha20Cipher, err := algorithms.NewChaCha20Cipher(key[32:64])
	if err != nil {
		return nil, fmt.Errorf("failed to create ChaCha20 cipher: %w", err)
	}

	rsDecoder, err := encoding.NewReedSolomonEncoder(config.DataShards, config.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reed-Solomon encoder: %w", err)
	}

	return &ChunkProcessor{
		serpentCipher:  serpentCipher,
		chaCha20Cipher: chaCha20Cipher,
		rsDecoder:      rsDecoder,
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
	decodedData, err := cp.rsDecoder.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("reed-solomon decoding failed: %w", err)
	}

	chaCha20Decrypted, err := cp.chaCha20Cipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("ChaCha20 decryption failed: %w", err)
	}

	serpentDecrypted, err := cp.serpentCipher.Decrypt(chaCha20Decrypted)
	if err != nil {
		return nil, fmt.Errorf("Serpent decryption failed: %w", err)
	}

	return serpentDecrypted, nil
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
