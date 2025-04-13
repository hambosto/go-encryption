package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hambosto/go-encryption/internal/core"
)

type Terminal struct{}

func NewTerminal() *Terminal {
	return &Terminal{}
}

func (t *Terminal) Clear() error {
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

type FileFinder struct {
	skippedDirs  []string
	skippedFiles []string
}

func NewFileFinder() *FileFinder {
	return &FileFinder{
		skippedDirs:  []string{"vendor/", "node_modules/", ".git", ".github"},
		skippedFiles: []string{".go", "go.mod", "go.sum", ".nix", ".gitignore"},
	}
}

func (f *FileFinder) FindEligibleFiles(op core.OperationType) ([]string, error) {
	var files []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.isFileEligible(path, info, op) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (f *FileFinder) isFileEligible(path string, info os.FileInfo, op core.OperationType) bool {
	if info.IsDir() || strings.HasPrefix(info.Name(), ".") || f.shouldSkipPath(path) {
		return false
	}
	isEncrypted := strings.HasSuffix(path, ".enc")
	return (op == core.Encrypt && !isEncrypted) || (op == core.Decrypt && isEncrypted)
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
