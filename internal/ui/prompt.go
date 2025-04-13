package ui

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/hambosto/go-encryption/internal/core"
)

type Prompt struct{}

func NewPrompt() *Prompt {
	return &Prompt{}
}

func (p *Prompt) ConfirmOverwrite(path string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Output file %s already exists. Overwrite?", path),
		Default: false,
	}

	err := survey.AskOne(prompt, &result)
	if err != nil {
		return false, err
	}

	return result, nil
}

func (p *Prompt) GetPassword() (string, error) {
	var password string
	passwordPrompt := &survey.Password{
		Message: "Enter password:",
	}

	err := survey.AskOne(passwordPrompt, &password)
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	var confirm string
	confirmPrompt := &survey.Password{
		Message: "Confirm password:",
	}

	err = survey.AskOne(confirmPrompt, &confirm)
	if err != nil {
		return "", fmt.Errorf("failed to get password confirmation: %w", err)
	}

	if !bytes.Equal([]byte(password), []byte(confirm)) {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

func (p *Prompt) ConfirmDelete(path string, promptMsg string) (bool, core.DeleteType, error) {
	var result bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("%s %s", promptMsg, path),
		Default: false,
	}

	err := survey.AskOne(confirmPrompt, &result)
	if err != nil {
		return false, "", err
	}

	if !result {
		return false, "", nil
	}

	var deleteType string
	typeSelect := &survey.Select{
		Message: "Select delete type",
		Options: []string{
			string(core.DeleteTypeNormal),
			string(core.DeleteTypeSecure),
		},
	}

	err = survey.AskOne(typeSelect, &deleteType)
	if err != nil {
		return false, "", err
	}

	return true, core.DeleteType(deleteType), nil
}

func (p *Prompt) GetOperation() (core.OperationType, error) {
	var operationStr string
	prompt := &survey.Select{
		Message: "Select Operation:",
		Options: []string{string(core.Encrypt), string(core.Decrypt)},
	}
	if err := survey.AskOne(prompt, &operationStr); err != nil {
		return "", fmt.Errorf("operation selection failed: %w", err)
	}
	return core.OperationType(operationStr), nil
}

func (p *Prompt) SelectFile(files []string) (string, error) {
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
