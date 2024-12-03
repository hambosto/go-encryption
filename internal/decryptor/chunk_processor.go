package decryptor

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/constants"
	"github.com/hambosto/go-encryption/internal/encoding"
)

type ChunkProcessor struct {
	serpentCipher  *algorithms.SerpentCipher
	chaCha20Cipher *algorithms.ChaCha20Cipher
	rsDecoder      *encoding.ReedSolomonEncoder
	bufferPool     sync.Pool
	decompressPool sync.Pool
}

func NewChunkProcessor(key []byte) (*ChunkProcessor, error) {
	serpentCipher, err := algorithms.NewSerpentCipher(key[:32])
	if err != nil {
		return nil, fmt.Errorf("failed to create serpent cipher: %w", err)
	}

	chaCha20Cipher, err := algorithms.NewChaCha20Cipher(key[32:])
	if err != nil {
		return nil, fmt.Errorf("failed to create chacha20 cipher: %w", err)
	}

	rsDecoder, err := encoding.NewReedSolomonEncoder(constants.DataShards, constants.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reed solomon decoder: %w", err)
	}

	return &ChunkProcessor{
		serpentCipher:  serpentCipher,
		chaCha20Cipher: chaCha20Cipher,
		rsDecoder:      rsDecoder,
		bufferPool: sync.Pool{
			New: func() interface{} {
				buffer := make([]byte, constants.MaxEncryptedChunkSize)
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

func (cp *ChunkProcessor) processChunk(chunk []byte) ([]byte, error) {
	decodedData, err := cp.rsDecoder.Decode(chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %w", err)
	}

	chaCha20Decrypted, err := cp.chaCha20Cipher.Decrypt(decodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	serpentDecrypted, err := cp.serpentCipher.Decrypt(chaCha20Decrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return serpentDecrypted, nil
}

func (cp *ChunkProcessor) decompressData(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid data size: must be at least 4 bytes")
	}

	compressedSize := binary.BigEndian.Uint32(data[:4])
	if compressedSize > uint32(len(data)-4) {
		return nil, fmt.Errorf("invalid data size: must be at least %d bytes", compressedSize)
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
		return nil, fmt.Errorf("failed to decompress data: %w", err)
	}

	result := make([]byte, buffer.Len())
	copy(result, buffer.Bytes())

	return result, nil
}
