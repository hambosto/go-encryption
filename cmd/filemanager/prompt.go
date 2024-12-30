package filemanager

import (
	"bytes"
	"fmt"

	"github.com/manifoldco/promptui"
)

type UserPrompt struct {
	fm *FileManager
}

func NewUserPrompt(fm *FileManager) *UserPrompt {
	return &UserPrompt{fm: fm}
}

func (up *UserPrompt) ConfirmOverwrite(path string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Output file %s already exists. Overwrite", path),
		IsConfirm: true,
		Default:   "n",
	}

	result, err := prompt.Run()
	if err == promptui.ErrAbort {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return result == "y" || result == "Y", nil
}

func (up *UserPrompt) GetPassword() (string, error) {
	passwordPrompt := promptui.Prompt{
		Label: "Enter password",
		Mask:  '*',
	}

	password, err := passwordPrompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get password: %w", err)
	}

	confirmPrompt := promptui.Prompt{
		Label: "Confirm password",
		Mask:  '*',
	}

	confirm, err := confirmPrompt.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get password confirmation: %w", err)
	}

	if !bytes.Equal([]byte(password), []byte(confirm)) {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

func (up *UserPrompt) ConfirmDelete(path string, prompt string) (bool, DeleteType, error) {
	confirmPrompt := promptui.Prompt{
		Label:     fmt.Sprintf("%s %s", prompt, path),
		IsConfirm: true,
		Default:   "n",
	}

	result, err := confirmPrompt.Run()
	if err == promptui.ErrAbort {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	if result != "y" && result != "Y" {
		return false, "", nil
	}

	typeSelect := promptui.Select{
		Label: "Select delete type",
		Items: []string{
			string(DeleteTypeNormal),
			string(DeleteTypeSecure),
		},
	}

	_, deleteType, err := typeSelect.Run()
	if err != nil {
		return false, "", err
	}

	return true, DeleteType(deleteType), nil
}
