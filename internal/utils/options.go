package utils

import (
	"bytes"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

func PromptOverwrite(outputFile string) (bool, error) {
	overwrite := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Output file %s already exists. Overwrite?", outputFile),
		Default: false,
	}
	err := survey.AskOne(prompt, &overwrite)
	return overwrite, err
}

func PromptPassword() (string, error) {
	password := ""
	confirmPassword := ""

	passwordPrompt := &survey.Password{
		Message: "Enter password:",
	}
	err := survey.AskOne(passwordPrompt, &password)
	if err != nil {
		return "", fmt.Errorf("failed to ask for password: %w", err)
	}

	confirmPasswordPrompt := &survey.Password{
		Message: "Confirm password:",
	}
	err = survey.AskOne(confirmPasswordPrompt, &confirmPassword)
	if err != nil {
		return "", fmt.Errorf("failed to ask for password confirmation: %w", err)
	}

	if !bytes.Equal([]byte(password), []byte(confirmPassword)) {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

func PromptDeleteOriginal(inputFile string) (bool, string, error) {
	deleteOriginal := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Delete original file %s?", inputFile),
		Default: false,
	}
	err := survey.AskOne(prompt, &deleteOriginal)
	if err != nil {
		return false, "", err
	}

	if !deleteOriginal {
		return false, "", nil
	}

	deleteType := ""
	deletePrompt := &survey.Select{
		Message: "Select delete type:",
		Options: []string{
			"Normal delete (faster, but recoverable)",
			"Secure delete (slower, but unrecoverable)",
		},
		Default: "Normal delete (faster, but recoverable)",
	}
	err = survey.AskOne(deletePrompt, &deleteType)
	return deleteOriginal, deleteType, err
}

func PromptDeleteEncrypted(inputFile string) (bool, string, error) {
	deleteEncrypted := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Delete encrypted file %s?", inputFile),
		Default: false,
	}
	err := survey.AskOne(prompt, &deleteEncrypted)
	if err != nil {
		return false, "", err
	}

	if !deleteEncrypted {
		return false, "", nil
	}

	deleteType := ""
	deletePrompt := &survey.Select{
		Message: "Select delete type:",
		Options: []string{
			"Normal delete (faster, but recoverable)",
			"Secure delete (slower, but unrecoverable)",
		},
		Default: "Normal delete (faster, but recoverable)",
	}
	err = survey.AskOne(deletePrompt, &deleteType)
	return deleteEncrypted, deleteType, err
}
