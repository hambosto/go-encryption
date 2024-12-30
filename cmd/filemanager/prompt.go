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

func (up *UserPrompt) GetPassword() (string, error) {
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

func (up *UserPrompt) ConfirmDelete(path string, prompt string) (bool, DeleteType, error) {
	var result bool
	confirmPrompt := &survey.Confirm{
		Message: fmt.Sprintf("%s %s", prompt, path),
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
			string(DeleteTypeNormal),
			string(DeleteTypeSecure),
		},
	}

	err = survey.AskOne(typeSelect, &deleteType)
	if err != nil {
		return false, "", err
	}

	return true, DeleteType(deleteType), nil
}
