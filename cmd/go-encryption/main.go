package main

import (
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hambosto/go-encryption/internal/core"
	"github.com/hambosto/go-encryption/internal/crypto"
	"github.com/hambosto/go-encryption/internal/progress"
	"github.com/mbndr/figlet4go"
	"github.com/spf13/cobra"
)

func main() {
	// Initialize figlet for ASCII art rendering
	ascii := figlet4go.NewAsciiRender()
	renderOptions := figlet4go.NewRenderOptions()
	renderOptions.FontColor = []figlet4go.Color{
		figlet4go.ColorGreen,
		figlet4go.ColorYellow,
		figlet4go.ColorCyan,
		figlet4go.ColorBlue,
		figlet4go.ColorCyan,
	}

	// Create root command (parent command for encrypt and decrypt)
	rootCmd := &cobra.Command{
		Use:   "go-encryption",
		Short: "File encryption and decryption tool",
	}

	// Shared flags for input/output file paths and stealth mode
	var inputFile, outputFile string
	var stealthMode bool
	var password string

	// Password prompt configuration
	passwordPrompt := &survey.Password{
		Message: "Enter your password:",
	}

	// Encrypt command definition
	encryptCmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt a file",
		Run: func(cmd *cobra.Command, args []string) {
			// Check if input file is provided
			if inputFile == "" {
				log.Fatal("Input file is required")
			}

			// Print banner for encryption
			renderBanner, err := ascii.RenderOpts("Encryption", renderOptions)
			if err != nil {
				log.Fatalf("Failed to render banner: %v\n", err)
				os.Exit(1)
			}
			fmt.Print(renderBanner)

			// Prompt for the encryption password
			if err = survey.AskOne(passwordPrompt, &password); err != nil {
				log.Fatalf("Failed to get password: %v", err)
			}

			// Set default output file if not specified
			if outputFile == "" {
				outputFile = inputFile + ".encrypted"
			}

			// Initialize components for encryption
			progress := progress.NewProgressReporter(stealthMode)
			cryptoService := crypto.NewCryptoService()
			processor := core.NewFileProcessor(cryptoService, progress)

			// Encrypt the file
			err = processor.Encrypt(inputFile, outputFile, password)
			if err != nil {
				log.Fatalf("\nEncryption failed: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Decrypt command definition
	decryptCmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt a file",
		Run: func(cmd *cobra.Command, args []string) {
			// Check if input file is provided
			if inputFile == "" {
				log.Fatal("Input file is required")
			}

			// Print banner for decryption
			renderBanner, err := ascii.RenderOpts("Decryption", renderOptions)
			if err != nil {
				log.Fatalf("Failed to render banner: %v\n", err)
				os.Exit(1)
			}
			fmt.Print(renderBanner)

			// Prompt for the decryption password
			if err = survey.AskOne(passwordPrompt, &password); err != nil {
				log.Fatalf("Failed to get password: %v", err)
			}

			// Set default output file if not specified
			if outputFile == "" {
				outputFile = inputFile + ".decrypted"
			}

			// Initialize components for decryption
			progress := progress.NewProgressReporter(stealthMode)
			cryptoService := crypto.NewCryptoService()
			processor := core.NewFileProcessor(cryptoService, progress)

			// Decrypt the file
			err = processor.Decrypt(inputFile, outputFile, password)
			if err != nil {
				log.Fatalf("\nDecryption failed: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Adding flags to the encrypt command
	encryptCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file path (required)")
	encryptCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	encryptCmd.Flags().BoolVarP(&stealthMode, "stealth", "s", false, "Enable stealth mode")

	// Adding flags to the decrypt command
	decryptCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file path (required)")
	decryptCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	decryptCmd.Flags().BoolVarP(&stealthMode, "stealth", "s", false, "Enable stealth mode")

	// Mark the input flag as required for both commands
	encryptCmd.MarkFlagRequired("input")
	decryptCmd.MarkFlagRequired("input")

	// Add encrypt and decrypt commands to the root command
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Command execution failed: %v", err)
		os.Exit(1)
	}
}
