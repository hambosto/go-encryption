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
	}
	err := survey.AskOne(prompt, &result)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (p *Prompt) GetPassword() (string, error) {
	var password string
	var confirm string

	questions := []*survey.Question{
		{
			Name: "password",
			Prompt: &survey.Password{
				Message: "Enter password:",
			},
		},
		{
			Name: "confirm",
			Prompt: &survey.Password{
				Message: "Confirm password:",
			},
		},
	}

	answers := struct {
		Password string
		Confirm  string
	}{}

	err := survey.Ask(questions, &answers)
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	password = answers.Password
	confirm = answers.Confirm

	if !bytes.Equal([]byte(password), []byte(confirm)) {
		return "", fmt.Errorf("passwords do not match")
	}
	return password, nil
}

func (p *Prompt) ConfirmDelete(path string, promptMsg string) (bool, core.DeleteType, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("%s %s", promptMsg, path),
	}
	err := survey.AskOne(prompt, &result)
	if err != nil {
		return false, "", err
	}
	if !result {
		return false, "", nil
	}

	deleteOptions := []string{
		string(core.DeleteTypeNormal),
		string(core.DeleteTypeSecure),
	}
	var deleteType string
	deletePrompt := &survey.Select{
		Message: "Select delete type",
		Options: deleteOptions,
	}
	err = survey.AskOne(deletePrompt, &deleteType)
	if err != nil {
		return false, "", err
	}

	return true, core.DeleteType(deleteType), nil
}

func (p *Prompt) GetOperation() (core.OperationType, error) {
	operationOptions := []string{
		string(core.Encrypt),
		string(core.Decrypt),
	}
	var operationType string
	prompt := &survey.Select{
		Message: "Select Operation:",
		Options: operationOptions,
	}
	err := survey.AskOne(prompt, &operationType)
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
