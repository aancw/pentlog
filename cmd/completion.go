package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"pentlog/pkg/utils"
	"strings"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh]",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:
  $ source <(pentlog completion bash)

  # To load completions for each session, execute once:
  $ pentlog completion bash > /etc/bash_completion.d/pentlog

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ pentlog completion zsh > "${fpath[1]}/_pentlog"

  # You will need to start a new shell for this setup to take effect.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh"},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			runInteractiveCompletion(cmd)
			return
		}

		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		default:
			fmt.Println("Invalid shell specified. Supported: bash, zsh")
		}
	},
}

func runInteractiveCompletion(cmd *cobra.Command) {
	fmt.Println("Interactive Completion Setup")
	fmt.Println("----------------------------")

	options := []string{"zsh", "bash"}
	idx := utils.SelectItem("Select your shell", options)
	if idx == -1 {
		fmt.Println("Selection aborted")
		return
	}
	selectedShell := options[idx]

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		return
	}

	targetFile := filepath.Join(home, fmt.Sprintf(".pentlog.%s", selectedShell))
	rcFile := filepath.Join(home, fmt.Sprintf(".%src", selectedShell))

	confirmInstall := utils.SelectItem(fmt.Sprintf("Install completion script to %s?", targetFile), []string{"Yes", "No"})
	if confirmInstall == 0 {
		var buf bytes.Buffer
		switch selectedShell {
		case "bash":
			cmd.Root().GenBashCompletion(&buf)
		case "zsh":
			cmd.Root().GenZshCompletion(&buf)
		}

		err := os.WriteFile(targetFile, buf.Bytes(), 0644)
		if err != nil {
			fmt.Printf("Error writing completion file: %v\n", err)
			return
		}
		fmt.Printf("✓ Completion script saved to %s\n", targetFile)
	} else {
		fmt.Println("Skipping script installation.")
		return
	}

	sourceLine := fmt.Sprintf("source %s", targetFile)

	content, err := os.ReadFile(rcFile)
	exists := false
	if err == nil {
		if strings.Contains(string(content), sourceLine) {
			exists = true
			fmt.Printf("✓ Configuration already exists in %s\n", rcFile)
		}
	}

	if !exists {
		confirmRc := utils.SelectItem(fmt.Sprintf("Add '%s' to %s?", sourceLine, rcFile), []string{"Yes", "No"})
		if confirmRc == 0 {
			f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Error opening rc file: %v\n", err)
				return
			}
			defer f.Close()

			if _, err := f.WriteString(fmt.Sprintf("\n# pentlog completion\n%s\n", sourceLine)); err != nil {
				fmt.Printf("Error appending to rc file: %v\n", err)
				return
			}
			fmt.Printf("✓ Added source line to %s\n", rcFile)
		} else {
			fmt.Println("Skipping rc file update.")
		}
	}

	fmt.Println("\nSetup complete!")
	fmt.Printf("Please run: source %s\n", rcFile)
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
