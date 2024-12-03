package header

type FileHeader struct {
	Salt          []byte
	OriginalSize  uint64
	SerpentNonce  []byte
	ChaCha20Nonce []byte
}
