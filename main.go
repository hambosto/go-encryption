package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hambosto/go-encryption/cmd/decryption"
	"github.com/hambosto/go-encryption/cmd/encryption"
)

const (
	encExtension = ".enc"
	appName      = "go-encryption"
)

type Operation string

const (
	OperationEncrypt Operation = "Encrypt"
	OperationDecrypt Operation = "Decrypt"
)

func clearTerminal() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	case "linux":
		cmd = exec.Command("clear")
	default:
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func getOperation() (Operation, error) {
	var operation string
	promptOperation := &survey.Select{
		Message: "Select operation:",
		Options: []string{
			string(OperationEncrypt),
			string(OperationDecrypt),
		},
		Default: string(OperationEncrypt),
	}
	err := survey.AskOne(promptOperation, &operation)
	if err != nil {
		return "", fmt.Errorf("failed to ask for operation: %w", err)
	}

	return Operation(operation), nil
}

func selectFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files found")
	}

	var selectedFile string
	promptSelectedFile := &survey.Select{
		Message: "Select file:",
		Options: files,
		Default: files[0],
	}

	if err := survey.AskOne(promptSelectedFile, &selectedFile); err != nil {
		return "", fmt.Errorf("failed to ask for selected file: %w", err)
	}

	return selectedFile, nil
}

func listFiles(operation Operation) ([]string, error) {
	var files []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access path %s: %v", path, err) // Handle path access errors
		}

		// Skip directories and hidden files
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Filter files based on the selected operation
		switch operation {
		case OperationEncrypt:
			if !strings.HasSuffix(path, encExtension) {
				files = append(files, path) // Add file for encryption if it doesn't have the .enc extension
			}
		case OperationDecrypt:
			if strings.HasSuffix(path, encExtension) {
				files = append(files, path) // Add file for decryption if it has the .enc extension
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return files, nil // Return the filtered list of files
}

func processFile(operation Operation, filename string) error {
	switch operation {
	case OperationEncrypt:
		return encryption.RunEncryption(filename)
	case OperationDecrypt:
		return decryption.RunDecryption(filename)
	}
	return nil
}

func main() {
	if err := clearTerminal(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clear terminal: %v\n", err)
		os.Exit(1)
	}

	operation, err := getOperation()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get operation: %v\n", err)
		os.Exit(1)
	}

	files, err := listFiles(operation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No files found")
		os.Exit(0)
	}

	selectedFile, err := selectFile(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to select file: %v\n", err)
		os.Exit(1)
	}

	if err := processFile(operation, selectedFile); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to process file: %v\n", err)
		os.Exit(1)
	}
}
