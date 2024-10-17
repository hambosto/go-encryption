package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"reflect"

	"github.com/hambosto/go-encryption/internal/crypto"
	"github.com/hambosto/go-encryption/internal/crypto/compression"
	"github.com/hambosto/go-encryption/internal/crypto/encoding"
	"github.com/hambosto/go-encryption/internal/crypto/padding"
)

func main() {
	// Create a new instance of CryptoService
	cs := crypto.NewCryptoService()

	// Define a password and some data to encrypt
	password := "my_secret_password"
	message := "this is test message"
	data := []byte(message)
	fmt.Printf("Original Data: %s\n", message)

	// Store the original data type
	var originalDataType string
	if reflect.TypeOf(message).Kind() == reflect.String {
		originalDataType = "string"
	} else {
		originalDataType = "[]byte"
	}

	// Step 1: Compress the data
	compressed, err := compression.ZlibCompress(data)
	if err != nil {
		log.Fatalf("Failed to compress data: %v", err)
	}
	fmt.Printf("Compressed Data (hex): %s\n", hex.EncodeToString(compressed))

	// Step 2: Generate a salt
	salt, err := cs.GenerateSalt()
	if err != nil {
		log.Fatalf("Failed to generate salt: %v", err)
	}
	fmt.Printf("Generated Salt (hex): %s\n", hex.EncodeToString(salt))

	// Step 3: Derive a key from the password and salt
	key, err := cs.DeriveKey(password, salt)
	if err != nil {
		log.Fatalf("Failed to derive key: %v", err)
	}
	fmt.Printf("Derived Key (hex): %s\n", hex.EncodeToString(key))

	// Step 4: Pad the compressed data
	padded, err := padding.Pad(compressed, 16)
	if err != nil {
		log.Fatalf("Failed to add padding: %v", err)
	}
	fmt.Printf("Padded Data (hex): %s\n", hex.EncodeToString(padded))

	// Step 5: Initialize Reed Solomon Codec
	rsCodec, err := encoding.NewReedSolomonCodec(encoding.Config{
		DataShards:   16,
		ParityShards: 48,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Reed Solomon Codec: %v", err)
	}

	// Step 6: Encode the padded data
	encoded, err := rsCodec.Encode(padded)
	if err != nil {
		log.Fatalf("Failed to encode data: %v", err)
	}
	fmt.Printf("Encoded Data (hex): %s\n", hex.EncodeToString(encoded))

	// Step 7: Encrypt the encoded data
	encryptedData, err := cs.EncryptData(encoded, key)
	if err != nil {
		log.Fatalf("Failed to encrypt data: %v", err)
	}
	fmt.Printf("Encrypted Data (hex): %s\n", hex.EncodeToString(encryptedData))

	// Step 8: Decrypt the data
	decryptedData, err := cs.DecryptData(encryptedData, key)
	if err != nil {
		log.Fatalf("Failed to decrypt data: %v", err)
	}
	fmt.Printf("Decrypted Data (hex): %s\n", hex.EncodeToString(decryptedData))

	// Step 9: Decode the decrypted data
	decoded, err := rsCodec.Decode(decryptedData)
	if err != nil {
		log.Fatalf("Failed to decode data: %v", err)
	}
	fmt.Printf("Decoded Data (hex): %s\n", hex.EncodeToString(decoded))

	// Step 10: Unpad the decoded data
	unpadded, err := padding.Unpad(decoded, 16)
	if err != nil {
		log.Fatalf("Failed to remove padding: %v", err)
	}
	fmt.Printf("Unpadded Data (hex): %s\n", hex.EncodeToString(unpadded))

	// Step 11: Decompress the unpadded data
	decompressed, err := compression.ZlibDecompress(unpadded)
	if err != nil {
		log.Fatalf("Failed to decompress data: %v", err)
	}
	fmt.Printf("Decompressed Data: %s\n", decompressed)

	// Check original data type and result data type
	var resultDataType string
	if reflect.TypeOf(decompressed).Kind() == reflect.Slice && reflect.TypeOf(decompressed).Elem().Kind() == reflect.Uint8 {
		resultDataType = "[]byte"
	} else {
		resultDataType = "string"
	}

	// Display the original data type
	fmt.Printf("\nOriginal Data Type: %s\n", originalDataType)
	fmt.Printf("Result Data Type: %s\n", resultDataType)

	// Verify that the decrypted data matches the original data
	if string(decompressed) != string(data) {
		log.Fatal("Decrypted data does not match original data")
	} else {
		fmt.Println("Data decrypted successfully and matches the original data.")
	}

	// Show original and result data
	fmt.Printf("\nOriginal Data: %s\n", message)
	fmt.Printf("Result Data: %s\n", decompressed)
}
