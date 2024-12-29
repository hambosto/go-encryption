package header

type Header struct {
	Salt          Salt
	OriginalSize  OriginalSize
	AesNonce      AesNonce
	ChaCha20Nonce ChaCha20Nonce
}

type HeaderBuilder struct {
	header Header
	err    error
}

func NewHeaderBuilder() *HeaderBuilder {
	return &HeaderBuilder{}
}

func (b *HeaderBuilder) WithSalt(salt []byte) *HeaderBuilder {
	if b.err != nil {
		return b
	}
	b.header.Salt = Salt{Value: salt}
	b.err = b.header.Salt.Validate(salt)
	return b
}

func (b *HeaderBuilder) WithOriginalSize(size uint64) *HeaderBuilder {
	if b.err != nil {
		return b
	}
	b.header.OriginalSize = OriginalSize{Value: size}
	return b
}

func (b *HeaderBuilder) WithAesNonce(nonce []byte) *HeaderBuilder {
	if b.err != nil {
		return b
	}
	b.header.AesNonce = AesNonce{Value: nonce}
	b.err = b.header.AesNonce.Validate(nonce)
	return b
}

func (b *HeaderBuilder) WithChaCha20Nonce(nonce []byte) *HeaderBuilder {
	if b.err != nil {
		return b
	}
	b.header.ChaCha20Nonce = ChaCha20Nonce{Value: nonce}
	b.err = b.header.ChaCha20Nonce.Validate(nonce)
	return b
}

func (b *HeaderBuilder) Build() (Header, error) {
	if b.err != nil {
		return Header{}, b.err
	}
	return b.header, nil
}
