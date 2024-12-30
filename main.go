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
	"github.com/common-nighthawk/go-figure"
	"github.com/hambosto/go-encryption/cmd"
)

type Operation string

const (
	Encrypt Operation = "Encrypt"
	Decrypt Operation = "Decrypt"
)

const encExtension = ".enc"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if err := clearTerminal(); err != nil {
		return fmt.Errorf("failed to clear terminal: %w", err)
	}

	displayLogo()

	operation, err := promptOperation()
	if err != nil {
		return fmt.Errorf("failed to get operation: %w", err)
	}

	files, err := findEligibleFiles(operation)
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

	return processFile(operation, selectedFile)
}

func displayLogo() {
	myFigure := figure.NewColorFigure("Go-Encryption", "rectangles", "green", true)
	myFigure.Print()
}

func clearTerminal() error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "cls"}
	default:
		cmd = "clear"
	}

	clearCmd := exec.Command(cmd, args...)
	clearCmd.Stdout = os.Stdout
	return clearCmd.Run()
}

func promptOperation() (Operation, error) {
	var operationStr string
	prompt := &survey.Select{
		Message: "Select Operation:",
		Options: []string{string(Encrypt), string(Decrypt)},
	}

	err := survey.AskOne(prompt, &operationStr)
	if err != nil {
		return "", fmt.Errorf("operation selection failed: %w", err)
	}

	return Operation(operationStr), nil
}

func findEligibleFiles(operation Operation) ([]string, error) {
	var files []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}

		if shouldSkipFile(info, path) {
			return nil
		}

		if isFileEligible(path, operation) {
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
		strings.Contains(path, "node_modules/") ||
		strings.Contains(path, ".git") ||
		strings.Contains(path, ".go") ||
		strings.Contains(path, "go.mod") ||
		strings.Contains(path, "go.sum")
}

func isFileEligible(path string, operation Operation) bool {
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
	}

	err := survey.AskOne(prompt, &selectedFile)
	if err != nil {
		return "", fmt.Errorf("file selection failed: %w", err)
	}

	return selectedFile, nil
}

func processFile(operation Operation, file string) error {
	switch operation {
	case Encrypt:
		return encryptFile(file)
	case Decrypt:
		return decryptFile(file)
	default:
		return fmt.Errorf("invalid operation: %s", operation)
	}
}

func encryptFile(file string) error {
	return cmd.RunEncryption(file)
}

func decryptFile(file string) error {
	return cmd.RunDecryption(file)
}
