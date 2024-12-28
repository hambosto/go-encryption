package header

type FileHeader struct {
	Salt          []byte
	OriginalSize  uint64
	AesNonce      []byte
	ChaCha20Nonce []byte
}
