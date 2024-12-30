package filemanager

import (
	"bytes"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

type UserPrompt struct {
	fm *FileManager
}

func NewUserPrompt(fm *FileManager) *UserPrompt {
	return &UserPrompt{fm: fm}
}

func (up *UserPrompt) ConfirmOverwrite(path string) (bool, error) {
	var overwrite bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Output file %s already exists. Overwrite?", path),
		Default: false,
	}
	err := survey.AskOne(prompt, &overwrite)
	return overwrite, err
}

func (up *UserPrompt) GetPassword() (string, error) {
	var password, confirm string

	if err := survey.AskOne(&survey.Password{
		Message: "Enter password:",
	}, &password); err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	if err := survey.AskOne(&survey.Password{
		Message: "Confirm password:",
	}, &confirm); err != nil {
		return "", fmt.Errorf("failed to get password confirmation: %w", err)
	}

	if !bytes.Equal([]byte(password), []byte(confirm)) {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

func (up *UserPrompt) ConfirmDelete(path string, prompt string) (bool, DeleteType, error) {
	var shouldDelete bool
	if err := survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("%s %s?", prompt, path),
		Default: false,
	}, &shouldDelete); err != nil {
		return false, "", err
	}

	if !shouldDelete {
		return false, "", nil
	}

	var deleteType string
	if err := survey.AskOne(&survey.Select{
		Message: "Select delete type:",
		Options: []string{
			string(DeleteTypeNormal),
			string(DeleteTypeSecure),
		},
		Default: string(DeleteTypeNormal),
	}, &deleteType); err != nil {
		return false, "", err
	}

	return true, DeleteType(deleteType), nil
}
