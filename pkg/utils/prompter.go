package utils

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
)

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
