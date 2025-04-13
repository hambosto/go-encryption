package ui

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/hambosto/go-encryption/internal/core"
)

type Prompt struct{}

func NewPrompt() *Prompt {
	return &Prompt{}
}

func (p *Prompt) ConfirmOverwrite(path string) (bool, error) {
	var result bool
	err := huh.NewConfirm().
		Title(fmt.Sprintf("Output file %s already exists. Overwrite?", path)).
		Value(&result).
		Run()
	if err != nil {
		return false, err
	}
	return result, nil
}

func (p *Prompt) GetPassword() (string, error) {
	var password string
	var confirm string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter password:").
				EchoMode(huh.EchoModePassword).
				Value(&password),
			huh.NewInput().
				Title("Confirm password:").
				EchoMode(huh.EchoModePassword).
				Value(&confirm),
		),
	)

	err := form.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	if !bytes.Equal([]byte(password), []byte(confirm)) {
		return "", fmt.Errorf("passwords do not match")
	}
	return password, nil
}

func (p *Prompt) ConfirmDelete(path string, promptMsg string) (bool, core.DeleteType, error) {
	var result bool
	err := huh.NewConfirm().
		Title(fmt.Sprintf("%s %s", promptMsg, path)).
		Value(&result).
		Run()
	if err != nil {
		return false, "", err
	}

	if !result {
		return false, "", nil
	}

	var deleteType string
	deleteOptions := []string{
		string(core.DeleteTypeNormal),
		string(core.DeleteTypeSecure),
	}

	err = huh.NewSelect[string]().
		Title("Select delete type").
		Options(
			huh.NewOptions(deleteOptions...)...,
		).
		Value(&deleteType).
		Run()
	if err != nil {
		return false, "", err
	}

	return true, core.DeleteType(deleteType), nil
}

func (p *Prompt) GetOperation() (core.OperationType, error) {
	var operationType string
	operationOptions := []string{
		string(core.Encrypt),
		string(core.Decrypt),
	}

	err := huh.NewSelect[string]().
		Title("Select Operation:").
		Options(
			huh.NewOptions(operationOptions...)...,
		).
		Value(&operationType).
		Run()
	if err != nil {
		return "", fmt.Errorf("operation selection failed: %w", err)
	}

	return core.OperationType(operationType), nil
}

func (p *Prompt) SelectFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", errors.New("no files available for selection")
	}

	var selectedFile string

	err := huh.NewSelect[string]().
		Title("Select file:").
		Options(
			huh.NewOptions(files...)...,
		).
		Value(&selectedFile).
		Run()
	if err != nil {
		return "", fmt.Errorf("file selection failed: %w", err)
	}

	return selectedFile, nil
}
