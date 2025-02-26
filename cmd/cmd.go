package cmd

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
	"github.com/hambosto/go-encryption/cmd/filemanager"
	"github.com/hambosto/go-encryption/cmd/operations"
)

type Operation string

const (
	Encrypt      Operation = "Encrypt"
	Decrypt      Operation = "Decrypt"
	encExtension           = ".enc"
)

type FileProcessor struct {
	fileManager *filemanager.FileManager
	userPrompt  *filemanager.UserPrompt
	processor   *operations.Operations
}

func NewFileProcessor() *FileProcessor {
	fileManager := filemanager.NewFileManager(3)
	userPrompt := filemanager.NewUserPrompt(fileManager)
	return &FileProcessor{
		fileManager: fileManager,
		userPrompt:  userPrompt,
		processor:   operations.NewOperation(fileManager, userPrompt),
	}
}

func Execute() {
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

	op, err := promptOperation()
	if err != nil {
		return fmt.Errorf("failed to get operation: %w", err)
	}

	files, err := findEligibleFiles(op)
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

	processor := NewFileProcessor()
	return processor.process(selectedFile, op)
}

func displayLogo() {
	figure.NewColorFigure("Go-Encryption", "rectangles", "green", true).Print()
}

func promptOperation() (Operation, error) {
	var operationStr string
	prompt := &survey.Select{
		Message: "Select Operation:",
		Options: []string{string(Encrypt), string(Decrypt)},
	}
	if err := survey.AskOne(prompt, &operationStr); err != nil {
		return "", fmt.Errorf("operation selection failed: %w", err)
	}
	return Operation(operationStr), nil
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
	if err := survey.AskOne(prompt, &selectedFile); err != nil {
		return "", fmt.Errorf("file selection failed: %w", err)
	}
	return selectedFile, nil
}

func findEligibleFiles(op Operation) ([]string, error) {
	finder := newFileFinder()
	return finder.findEligibleFiles(op)
}

type FileFinder struct {
	skippedDirs  []string
	skippedFiles []string
}

func newFileFinder() *FileFinder {
	return &FileFinder{
		skippedDirs:  []string{"vendor/", "node_modules/", ".git", ".github"},
		skippedFiles: []string{".go", "go.mod", "go.sum", ".nix", ".gitignore"},
	}
}

func (f *FileFinder) findEligibleFiles(op Operation) ([]string, error) {
	var files []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", path, err)
		}
		if f.isFileEligible(path, info, op) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (f *FileFinder) isFileEligible(path string, info os.FileInfo, op Operation) bool {
	if info.IsDir() || strings.HasPrefix(info.Name(), ".") || f.shouldSkipPath(path) {
		return false
	}
	isEncrypted := strings.HasSuffix(path, encExtension)
	return (op == Encrypt && !isEncrypted) || (op == Decrypt && isEncrypted)
}

func (f *FileFinder) shouldSkipPath(path string) bool {
	for _, skip := range f.skippedDirs {
		if strings.Contains(path, skip) {
			return true
		}
	}
	for _, skip := range f.skippedFiles {
		if strings.Contains(path, skip) {
			return true
		}
	}
	return false
}

func clearTerminal() error {
	cmd, args := getClearCommand()
	clearCmd := exec.Command(cmd, args...)
	clearCmd.Stdout = os.Stdout
	return clearCmd.Run()
}

func getClearCommand() (string, []string) {
	switch runtime.GOOS {
	case "windows":
		return "cmd", []string{"/c", "cls"}
	default:
		return "clear", nil
	}
}

func (fp *FileProcessor) process(input string, op Operation) error {
	config := operations.OperationConfig{
		InputPath:  input,
		OutputPath: determineOutputPath(input, op),
		Operation:  mapOperation(op),
	}
	if err := fp.processor.Process(config); err != nil {
		return fmt.Errorf("%s failed: %w", strings.ToLower(string(op)), err)
	}
	return nil
}

func determineOutputPath(input string, op Operation) string {
	if op == Encrypt {
		return input + encExtension
	}
	return strings.TrimSuffix(input, encExtension)
}

func mapOperation(op Operation) operations.OperationType {
	if op == Encrypt {
		return operations.OperationEncrypt
	}
	return operations.OperationDecrypt
}
