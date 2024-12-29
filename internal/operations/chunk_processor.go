package operations

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/hambosto/go-encryption/internal/algorithms"
	"github.com/hambosto/go-encryption/internal/encoding"
)

const ChunkSize = 1024 * 1024

type ChunkProcessor struct {
	aesCipher      *algorithms.AESCipher
	chaCha20Cipher *algorithms.ChaCha20Cipher
	encoder        *encoding.Encoder
	isEncryption   bool
}

func NewChunkProcessor(key []byte, isEncryption bool) (*ChunkProcessor, error) {
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
		isEncryption:   isEncryption,
	}, nil
}

func (cp *ChunkProcessor) ProcessChunk(chunk []byte) ([]byte, error) {
	if cp.isEncryption {
		return cp.encrypt(chunk)
	}

	return cp.decrypt(chunk)
}

func (cp *ChunkProcessor) encrypt(chunk []byte) ([]byte, error) {
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

func (cp *ChunkProcessor) decrypt(chunk []byte) ([]byte, error) {
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

	zlibDecompressed, err := cp.decompressData(aesDecrypted)
	if err != nil {
		return nil, fmt.Errorf("zlib decompression failed: %w", err)
	}

	return zlibDecompressed, nil
}

func (cp *ChunkProcessor) compressData(data []byte) ([]byte, error) {
	var buffer bytes.Buffer

	zw, err := zlib.NewWriterLevel(&buffer, zlib.BestSpeed)
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

	return buffer.Bytes(), nil
}

func (cp *ChunkProcessor) decompressData(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("invalid data: insufficient bytes for size information or invalid padding")
	}

	compressedSize := binary.BigEndian.Uint32(data)
	if compressedSize > uint32(len(data)-4) {
		return nil, fmt.Errorf("invalid compressed data size: expected %d, got %d", compressedSize, len(data)-4)
	}

	compressedData := data[4 : 4+compressedSize]
	var buffer bytes.Buffer

	zr, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("failed to create zlib reader: %w", err)
	}
	defer zr.Close()

	if _, err := io.Copy(&buffer, zr); err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}

	return buffer.Bytes(), nil
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
