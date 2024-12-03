package encryptor

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/constants"
	"github.com/hambosto/go-encryption/internal/encoding"
)

type ChunkProcessor struct {
	serpentCipher  *algorithms.SerpentCipher
	chaCha20Cipher *algorithms.ChaCha20Cipher
	rsEncoder      *encoding.ReedSolomonEncoder
	bufferPool     sync.Pool
	compressPool   sync.Pool
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

	rsEncoder, err := encoding.NewReedSolomonEncoder(constants.DataShards, constants.ParityShards)
	if err != nil {
		return nil, fmt.Errorf("failed to create reed-solomon encoder: %w", err)
	}

	return &ChunkProcessor{
		serpentCipher:  serpentCipher,
		chaCha20Cipher: chaCha20Cipher,
		rsEncoder:      rsEncoder,
		bufferPool: sync.Pool{
			New: func() interface{} {
				buffer := make([]byte, constants.MaxChunkSize)
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

func (cp *ChunkProcessor) processChunk(chunk []byte) ([]byte, error) {
	compressedData, err := cp.compressData(chunk)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}

	paddedData := cp.padData(compressedData)

	serpentEncrypted, err := cp.serpentCipher.Encrypt(paddedData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	chaCha20Encrypted, err := cp.chaCha20Cipher.Encrypt(serpentEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	rsEncoded, err := cp.rsEncoder.Encode(chaCha20Encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	return rsEncoded, nil
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
