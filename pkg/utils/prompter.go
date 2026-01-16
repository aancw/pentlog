package utils

import (
	"fmt"
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
