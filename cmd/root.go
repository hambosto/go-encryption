package cmd

import (
	"fmt"

	"github.com/hambosto/go-encryption/internal/core"
	"github.com/hambosto/go-encryption/internal/ui"
)

func Execute() error {
	terminal := ui.NewTerminal()
	if err := terminal.Clear(); err != nil {
		return fmt.Errorf("failed to clear terminal: %w", err)
	}

	prompt := ui.NewPrompt()
	fileManager := core.NewFileManager(3)
	processor := core.NewProcessor(fileManager, prompt)

	opType, err := prompt.GetOperation()
	if err != nil {
		return fmt.Errorf("failed to get operation: %w", err)
	}

	fileFinder := ui.NewFileFinder()
	files, err := fileFinder.FindEligibleFiles(opType)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No eligible files found.")
		return nil
	}

	selectedFile, err := prompt.SelectFile(files)
	if err != nil {
		return fmt.Errorf("failed to select file: %w", err)
	}

	return processor.ProcessFile(selectedFile, opType)
}
