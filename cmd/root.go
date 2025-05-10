package cmd

import (
	"fmt"
	"os"

	"github.com/hambosto/go-encryption/internal/core"
	"github.com/hambosto/go-encryption/internal/ui"
)

func Execute() {
	terminal := ui.NewTerminal()
	if err := terminal.Clear(); err != nil {
		fmt.Printf("Error: failed to clear terminal: %v\n", err)
		os.Exit(1)
	}

	prompt := ui.NewPrompt()
	fileManager := core.NewFileManager(3)
	processor := core.NewProcessor(fileManager, prompt)

	operation, err := prompt.GetOperation()
	if err != nil {
		fmt.Printf("Error: failed to get operation: %v\n", err)
		os.Exit(1)
	}

	fileFinder := ui.NewFileFinder()
	files, err := fileFinder.FindEligibleFiles(operation)
	if err != nil {
		fmt.Printf("Error: failed to list files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No eligible files found.")
		os.Exit(1)
	}

	selectedFile, err := prompt.SelectFile(files)
	if err != nil {
		fmt.Printf("Error: failed to select file: %v\n", err)
		os.Exit(1)
	}

	if err := processor.ProcessFile(selectedFile, operation); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
