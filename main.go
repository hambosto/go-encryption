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

const (
	encExtension = ".enc"
)

type Operation string

const (
	Encrypt Operation = "Encrypt"
	Decrypt Operation = "Decrypt"
)

func clearTerminal() error {
	cmd := exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	return cmd.Run()
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

func listEligibleFiles(operation Operation) ([]string, error) {
	var files []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		switch operation {
		case Encrypt:
			if !strings.HasSuffix(path, encExtension) {
				files = append(files, path)
			}
		case Decrypt:
			if strings.HasSuffix(path, encExtension) {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

func promptFileSelection(files []string) (string, error) {
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

func handleFileOperation(operation Operation, file string) error {
	switch operation {
	case Encrypt:
		return cmd.RunEncryption(file)
	case Decrypt:
		return cmd.RunDecryption(file)
	default:
		return errors.New("invalid operation")
	}
}

func main() {
	if err := clearTerminal(); err != nil {
		fmt.Fprintf(os.Stderr, "Error clearing terminal: %v\n", err)
		return
	}

	operation, err := promptOperation()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting operation: %v\n", err)
		return
	}

	files, err := listEligibleFiles(operation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing files: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No eligible files found.")
		return
	}

	selectedFile, err := promptFileSelection(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error selecting file: %v\n", err)
		return
	}

	if err := handleFileOperation(operation, selectedFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error processing file: %v\n", err)
	}
}
