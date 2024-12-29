package header

type FileHeader struct {
	Salt          []byte
	OriginalSize  uint64
	AesNonce      []byte
	ChaCha20Nonce []byte
}

type FileHeaderBuilder struct {
	salt          []byte
	originalSize  uint64
	aesNonce      []byte
	chaCha20Nonce []byte
}

func NewFileHeaderBuilder() *FileHeaderBuilder {
	return &FileHeaderBuilder{}
}

func (b *FileHeaderBuilder) SetSalt(salt []byte) *FileHeaderBuilder {
	b.salt = salt
	return b
}

func (b *FileHeaderBuilder) SetOriginalSize(originalSize uint64) *FileHeaderBuilder {
	b.originalSize = originalSize
	return b
}

func (b *FileHeaderBuilder) SetAesNonce(aesNonce []byte) *FileHeaderBuilder {
	b.aesNonce = aesNonce
	return b
}

func (b *FileHeaderBuilder) SetChaCha20Nonce(chaCha20Nonce []byte) *FileHeaderBuilder {
	b.chaCha20Nonce = chaCha20Nonce
	return b
}

func (b *FileHeaderBuilder) Build() FileHeader {
	return FileHeader{
		Salt:          b.salt,
		OriginalSize:  b.originalSize,
		AesNonce:      b.aesNonce,
		ChaCha20Nonce: b.chaCha20Nonce,
	}
}
