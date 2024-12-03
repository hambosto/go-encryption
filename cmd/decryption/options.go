package decryption

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
)

func promptOverwrite(outputFile string) (bool, error) {
	overwrite := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Output file %s already exists. Overwrite?", outputFile),
		Default: false,
	}
	err := survey.AskOne(prompt, &overwrite)
	return overwrite, err
}

func promptPassword() (string, error) {
	password := ""
	prompt := &survey.Password{
		Message: "Enter password:",
	}
	err := survey.AskOne(prompt, &password)
	return password, err
}

func promptDeleteEncrypted(inputFile string) (bool, string, error) {
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
