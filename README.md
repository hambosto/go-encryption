# Go Encryption

A secure and efficient file encryption tool written in Go that implements multi-layer encryption with parallel processing capabilities.

[![Go Report Card](https://goreportcard.com/badge/github.com/hambosto/go-encryption)](https://goreportcard.com/report/github.com/hambosto/go-encryption)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Multi-layer Encryption**: Combines multiple encryption algorithms for enhanced security
  - AES-GCM (Authenticated Encryption with Associated Data)
  - ChaCha20 stream cipher
  - Reed-Solomon error correction encoding
- **Data Compression**: Utilizes LZ4 compression for efficient storage
- **Parallel Processing**: Leverages multi-core systems for faster encryption/decryption
- **Progress Tracking**: Real-time progress bar for file operations
- **Error Recovery**: Built-in error correction using Reed-Solomon encoding
- **Interactive CLI**: User-friendly command-line interface with file selection
- **Cross-platform**: Supports Windows, macOS, and Linux

⚠️ **Important File Size Notice**: Encrypted files will be significantly larger than the original files due to chunk-based encryption. For example, a 26MB unencrypted file will become approximately 96MB after encryption. Please ensure you have sufficient storage space available before encrypting large files.

## Installation

### From Releases

1. Go to the [Releases](https://github.com/hambosto/go-encryption/releases) page
2. Download the latest binary for your operating system.
3. Make the file executable (Unix-based systems):
   ```bash
   chmod +x go-encryption*
   ```

### From Source

Requirements:
- Go 1.19 or higher
- Git

```bash
# Clone the repository
git clone https://github.com/hambosto/go-encryption.git

# Change to project directory
cd go-encryption

# Build the project
go build -o go-encryption

# (Optional) Install globally
go install
```

## Usage

1. Run the application:
   ```bash
   ./go-encryption
   ```

2. Select operation:
   - Choose between `Encrypt` or `Decrypt` using arrow keys

3. Select file:
   - Navigate through available files using arrow keys
   - For encryption: shows all non-encrypted files
   - For decryption: shows only `.enc` files

The program will process the selected file and display progress in real-time.

### Encrypted File Format

- Encrypted files are saved with the `.enc` extension
- Original filename is preserved when decrypting
- Files are processed in chunks for efficient memory usage

## Security Features

### Encryption Layers

1. **AES-GCM** (Authenticated Encryption with Associated Data)
   - 256-bit key
   - Provides confidentiality, integrity, and authenticity
   - Galois/Counter Mode (GCM) ensures secure and efficient encryption

2. **ChaCha20** (Stream Cipher)
   - 256-bit key
   - Modern, high-performance cipher

3. **Reed-Solomon** (Error Correction)
   - Adds redundancy for error recovery
   - Helps protect against data corruption

### Additional Security Measures

- Unique nonces for each encryption layer
- Secure memory handling with buffer pools
- Padding and alignment for block cipher security
- Size header encryption

## Technical Details

### Performance Features

- **Parallel Processing**: Utilizes all available CPU cores
- **Buffer Pools**: Reduces memory allocations
- **Chunked Processing**: Handles large files efficiently
- **Compressed Output**: Reduces encrypted file size

## Building from Source

To build for specific platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o go-encryption.exe

# macOS
GOOS=darwin GOARCH=amd64 go build -o go-encryption

# Linux
GOOS=linux GOARCH=amd64 go build -o go-encryption
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Uses [progressbar](https://github.com/schollz/progressbar) for progress tracking
- Uses [survey](https://github.com/AlecAivazis/survey) for interactive CLI

## Security Notice

While this tool implements strong encryption algorithms, please note:
- Keep your encryption keys secure
- Back up important files before encryption
- Use strong, unique keys for each file
- Store nonces securely for decryption

