# Go Advanced Encryption Tool

A high-security, feature-rich command-line file encryption and decryption tool built in Go. This tool implements military-grade encryption standards with multiple encryption layers, advanced integrity checks, and secure data handling features comparable to PicoCrypt.

## 🔐 Security Features

### Multi-Layer Encryption
- Serpent encryption algorithm for the first encryption layer
- ChaCha20 for the second encryption layer
- Enforced maximum security parameters
- No option for weaker encryption methods

### Advanced Key Derivation
- Argon2id implementation for password hashing
- HKDF (HMAC-based Key Derivation Function) for key expansion
- Secure random number generation for keys and nonces
- Protection against brute-force attacks

### Data Integrity & Protection
- BLAKE2b cryptographic hash function for file integrity verification
- Reed-Solomon error correction encoding
- Encrypted metadata to protect file information
- Automatic secure data shredding (optional)

### Additional Security Measures
- Integrated compression before encryption
- Stealth mode operation
- Secure memory handling
- Protected against timing attacks
- Zero-knowledge architecture

## ⚡ Performance Features

- Progress reporting with detailed status updates
- Parallel processing for large files
- Efficient memory management
- Optimized compression algorithms

## 📦 Installation

### Using Pre-built Binaries (Recommended)

1. Visit the [Releases](https://github.com/hambosto/go-encryption/releases) page
2. Download the appropriate binary for your system:

**For macOS:**
- Apple Silicon (M1/M2): `go-encryption-<version>-darwin-arm64`
- Intel: `go-encryption-<version>-darwin-amd64`

**For Windows:**
- 64-bit: `go-encryption-<version>-windows-amd64.exe`
 
**For Linux:**
- 64-bit: `go-encryption-<version>-linux-amd64`

3. For Unix-based systems (macOS, Linux), make the file executable:
   ```bash
   chmod +x go-encryption-*
   ```

### Building from Source

If you prefer to build from source, ensure you have  Go 1.19 or higher installed:

```bash
# Clone the repository
git clone https://github.com/hambosto/go-encryption.git

# Navigate to the project directory
cd go-encryption

# Install dependencies
go mod download

# Build the project
CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static" -s -w' -o go-encryption ./cmd/go-encryption
```

## 💻 Usage

### Basic Commands

```bash
# Encrypt a file with maximum security
./go-encryption encrypt -i <input-file>

# Decrypt a file
./go-encryption decrypt -i <input-file>
```

### Command Flags

Both `encrypt` and `decrypt` commands support:

- `-i, --input`: Input file path (required)
- `-o, --output`: Output file path (optional)
- `-s, --stealth`: Enable stealth mode (Disable Output)


## 🔨 Technical Implementation

### Encryption Process
1. File preparation and integrity checking
2. Data compression using optimized algorithms
3. First-layer encryption using Serpent
4. Second-layer encryption using ChaCha20
5. Reed-Solomon encoding for error correction
6. Metadata encryption and protection

### Key Derivation Process
```
Input Password → Argon2id → HKDF Expansion → Multiple Encryption Keys
```

### Security Parameters
- Argon2id: Memory=1GB, Iterations=4, Parallelism=8
- ChaCha20: 256-bit keys, 96-bit nonces
- Serpent: 256-bit keys
- BLAKE2b: 512-bit output
- Reed-Solomon: 255,223 encoding

## ⚠️ Security Considerations

1. Password Requirements:
   - Minimum 12 characters
   - Mix of uppercase, lowercase, numbers, and symbols
   - No common dictionary words

2. File Handling:
   - Temporary files are encrypted in memory
   - Sensitive data is wiped from memory after use

3. Error Handling:
   - Detailed error messages without exposing sensitive information
   - Graceful handling of corruption and tampering attempts
   - Verification of file integrity before and after operations

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by PicoCrypt's security model
- Thanks to the Go cryptography community
- Contributors to all dependent libraries

---

**Note**: This tool is designed for users requiring high-security file encryption. For basic encryption needs, simpler alternatives may be more appropriate.
