package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hambosto/go-encryption/cmd"
)

type Config struct {
	Version      string
	EncExtension string
}

type Operation string

const (
	Encrypt Operation = "Encrypt"
	Decrypt Operation = "Decrypt"
)

type FileProcessor struct {
	config Config
}

func NewFileProcessor() *FileProcessor {
	return &FileProcessor{
		config: Config{
			Version:      "1.0",
			EncExtension: ".enc",
		},
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	processor := NewFileProcessor()

	if err := clearTerminal(); err != nil {
		return fmt.Errorf("failed to clear terminal: %w", err)
	}

	processor.displayLogo()

	operation, err := promptOperation()
	if err != nil {
		return fmt.Errorf("failed to get operation: %w", err)
	}

	files, err := processor.findEligibleFiles(operation)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No eligible files found.")
		return nil
	}

	selectedFile, err := selectFile(files)
	if err != nil {
		return fmt.Errorf("failed to select file: %w", err)
	}

	return processor.processFile(operation, selectedFile)
}

func (fp *FileProcessor) displayLogo() {
	logo := fmt.Sprintf(`
 ▗▄▄▖ ▗▄▖ ▗▄▄▄▖▗▖  ▗▖ ▗▄▄▖▗▄▄▖▗▖  ▗▖▗▄▄▖▗▄▄▄▖▗▄▄▄▖ ▗▄▖ ▗▖  ▗▖
▐▌   ▐▌ ▐▌▐▌   ▐▛▚▖▐▌▐▌   ▐▌ ▐▌▝▚▞▘ ▐▌ ▐▌ █    █  ▐▌ ▐▌▐▛▚▖▐▌
▐▌▝▜▌▐▌ ▐▌▐▛▀▀▘▐▌ ▝▜▌▐▌   ▐▛▀▚▖ ▐▌  ▐▛▀▘  █    █  ▐▌ ▐▌▐▌ ▝▜▌ v%s
▝▚▄▞▘▝▚▄▞▘▐▙▄▄▖▐▌  ▐▌▝▚▄▄▖▐▌ ▐▌ ▐▌  ▐▌    █  ▗▄█▄▖▝▚▄▞▘▐▌  ▐▌
 Secure file encryption and decryption CLI tool built with Go
`, fp.config.Version)
	fmt.Println(logo)
}

func clearTerminal() error {
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "cls"}
	} else {
		cmd = "clear"
	}

	clearCmd := exec.Command(cmd, args...)
	clearCmd.Stdout = os.Stdout
	return clearCmd.Run()
}

func promptOperation() (Operation, error) {
	var operation string
	prompt := &survey.Select{
		Message: "Select operation:",
		Options: []string{string(Encrypt), string(Decrypt)},
		Default: string(Encrypt),
	}

	if err := survey.AskOne(prompt, &operation); err != nil {
		return "", fmt.Errorf("operation selection failed: %w", err)
	}

	return Operation(operation), nil
}

func (fp *FileProcessor) findEligibleFiles(operation Operation) ([]string, error) {
	var files []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		if shouldSkipFile(info, path) {
			return nil
		}

		if isFileEligible(path, operation, fp.config.EncExtension) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

func shouldSkipFile(info os.FileInfo, path string) bool {
	return info.IsDir() ||
		strings.HasPrefix(info.Name(), ".") ||
		strings.Contains(path, "vendor/") ||
		strings.Contains(path, "node_modules/")
}

func isFileEligible(path string, operation Operation, encExtension string) bool {
	isEncrypted := strings.HasSuffix(path, encExtension)
	return (operation == Encrypt && !isEncrypted) ||
		(operation == Decrypt && isEncrypted)
}

func selectFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", errors.New("no files available for selection")
	}

	var selectedFile string
	prompt := &survey.Select{
		Message: "Select file:",
		Options: files,
		Default: files[0],
	}

	if err := survey.AskOne(prompt, &selectedFile); err != nil {
		return "", fmt.Errorf("file selection failed: %w", err)
	}

	return selectedFile, nil
}

func (fp *FileProcessor) processFile(operation Operation, file string) error {
	switch operation {
	case Encrypt:
		return cmd.RunEncryption(file)
	case Decrypt:
		return cmd.RunDecryption(file)
	default:
		return fmt.Errorf("invalid operation: %s", operation)
	}
}
