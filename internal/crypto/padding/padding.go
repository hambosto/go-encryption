package padding

import (
	"errors"
)

// Pad adds PKCS#7 padding to the input buffer.
// It returns the padded buffer and any potential error.
func Pad(inputBuffer []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, errors.New("pkcs7: block size must be positive")
	}

	inputLength := len(inputBuffer)
	paddingLength := blockSize - (inputLength % blockSize)
	paddedBuffer := make([]byte, inputLength+paddingLength)
	copy(paddedBuffer, inputBuffer)

	// Fill the padding bytes with the padding length
	paddingByte := byte(paddingLength)
	for i := 0; i < paddingLength; i++ {
		paddedBuffer[inputLength+i] = paddingByte
	}
	return paddedBuffer, nil
}

// Unpad removes PKCS#7 padding from the input buffer.
// It returns the original buffer and any potential error.
func Unpad(paddedData []byte, blockSize int) ([]byte, error) {
	if len(paddedData) == 0 {
		return nil, errors.New("pkcs7: input buffer is empty")
	}

	if len(paddedData)%blockSize != 0 {
		return nil, errors.New("pkcs7: padded value wasn't in correct size")
	}

	paddingLength := int(paddedData[len(paddedData)-1])
	if paddingLength < 1 || paddingLength > blockSize {
		return nil, errors.New("pkcs7: invalid padding length")
	}

	outputLength := len(paddedData) - paddingLength
	return paddedData[:outputLength], nil
}

