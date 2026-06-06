package utils

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/manifoldco/promptui"
)

func PromptSelect(label string, options []string) string {
	var selected string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(label).
				Options(huh.NewOptions(options...)...).
				Value(&selected),
		),
	)

	err := form.Run()
	if err != nil {
		return ""
	}
	return selected
}

func PromptSelectWithDefault(label string, options []string, defaultValue string) string {
	var selected string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(label).
				Options(huh.NewOptions(options...)...).
				Value(&selected).
				Value(&selected),
		),
	)
	selected = defaultValue

	err := form.Run()
	if err != nil {
		return ""
	}
	return selected
}

func PromptString(label string, defaultValue string) string {
	displayLabel := label
	if defaultValue != "" {
		displayLabel = fmt.Sprintf("%s [%s]", label, defaultValue)
	}

	prompt := promptui.Prompt{
		Label:     displayLabel,
		AllowEdit: true,
	}

	result, err := prompt.Run()
	if err != nil {
		return ""
	}

	result = strings.TrimSpace(result)
	if result == "" {
		return defaultValue
	}
	return result
}

func PromptPassword(label string) string {
	prompt := promptui.Prompt{
		Label: label,
		Mask:  '*',
	}

	result, err := prompt.Run()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(result)
}

func ReadSecretFromStdin(stdin io.Reader) (string, error) {
	if file, ok := stdin.(*os.File); ok {
		info, err := file.Stat()
		if err != nil {
			return "", fmt.Errorf("inspect stdin: %w", err)
		}
		if info.Mode()&os.ModeCharDevice != 0 {
			return "", errors.New("stdin is a terminal; use the password prompt or pipe a secret into --password-stdin")
		}
	}

	return ReadSecretFromReader(stdin)
}

func ReadSecretFromReader(reader io.Reader) (string, error) {
	if reader == nil {
		return "", errors.New("stdin reader is not available")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	secret := strings.TrimRight(string(data), "\r\n")
	if secret == "" {
		return "", errors.New("no secret received on stdin")
	}

	return secret, nil
}

func SelectItem(label string, items []string) int {
	prompt := promptui.Select{
		Label: label,
		Items: items,
		Size:  10,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return -1
	}

	return idx
}
