package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"pentlog/pkg/metadata"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start a new shell with the engagement context loaded",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, err := metadata.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading context: %v\n", err)
			os.Exit(1)
		}

		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}

		c := exec.Command(shell)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		newEnv := os.Environ()
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_CLIENT=%s", ctx.Client))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_ENGAGEMENT=%s", ctx.Engagement))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SCOPE=%s", ctx.Scope))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_OPERATOR=%s", ctx.Operator))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_PHASE=%s", ctx.Phase))
		
c.Env = newEnv

		fmt.Printf("Spawning %s with active context: %s [%s]\n", shell, ctx.Client, ctx.Phase)
		fmt.Println("Type 'exit' or Ctrl+D to leave this context.")

		if err := c.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				os.Exit(exitError.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "Error running shell: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}