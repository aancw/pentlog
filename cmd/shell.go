package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"pentlog/pkg/logs"
	"pentlog/pkg/metadata"
	"pentlog/pkg/system"
	"pentlog/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start a recorded shell with the engagement context loaded",
	Run: func(cmd *cobra.Command, args []string) {
		if os.Getenv("PENTLOG_SESSION_LOG_PATH") != "" {
			fmt.Fprintln(os.Stderr, "Error: You are already in a pentlog shell session.")
			os.Exit(1)
		}

		ctx, err := metadata.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading context: %v\n", err)
			os.Exit(1)
		}

		logDir, err := system.EnsureLogDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error preparing log dir: %v\n", err)
			os.Exit(1)
		}

		sessionDir := filepath.Join(
			logDir,
			utils.Slugify(ctx.Client),
			utils.Slugify(ctx.Engagement),
			utils.Slugify(ctx.Phase),
		)

		if err := os.MkdirAll(sessionDir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating session dir: %v\n", err)
			os.Exit(1)
		}

		timestamp := time.Now().Format("20060102-150405")
		baseName := fmt.Sprintf("session-%s-%s", utils.Slugify(ctx.Operator), timestamp)
		logFilePath := filepath.Join(sessionDir, baseName+".tty")
		metaFilePath := filepath.Join(sessionDir, baseName+".json")

		meta := logs.SessionMetadata{
			Client:     ctx.Client,
			Engagement: ctx.Engagement,
			Scope:      ctx.Scope,
			Operator:   ctx.Operator,
			Phase:      ctx.Phase,
			Timestamp:  time.Now().Format(time.RFC3339),
		}

		if err := writeMetadata(metaFilePath, meta); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing metadata: %v\n", err)
			os.Exit(1)
		}

		recorder := system.NewRecorder()
		c, err := recorder.BuildCommand("", logFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating recorder command: %v\n", err)
			os.Exit(1)
		}
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		shell := os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/sh"
		}
		baseShell := filepath.Base(shell)

		tempDir, err := os.MkdirTemp("", "pentlog-shell-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not create temp dir for shell config: %v\n", err)
		} else {
			defer os.RemoveAll(tempDir)
		}

		newEnv := os.Environ()

		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_CLIENT=%s", ctx.Client))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_ENGAGEMENT=%s", ctx.Engagement))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SCOPE=%s", ctx.Scope))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_OPERATOR=%s", ctx.Operator))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_PHASE=%s", ctx.Phase))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SESSION_DIR=%s", sessionDir))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SESSION_METADATA_PATH=%s", metaFilePath))
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SESSION_LOG_PATH=%s", logFilePath))

		promptSegment := fmt.Sprintf("(pentlog:%s/%s)", ctx.Client, ctx.Phase)

		if baseShell == "zsh" && tempDir != "" {
			zshrcPath := filepath.Join(tempDir, ".zshrc")

			userZshrc := filepath.Join(os.Getenv("HOME"), ".zshrc")
			zshContent := ""
			if _, err := os.Stat(userZshrc); err == nil {
				zshContent += fmt.Sprintf("source %s\n", userZshrc)
			}

			zshContent += "\n# Pentlog Prompt Injection\n"
			zshContent += "setopt TRANSIENT_RPROMPT\n"
			zshContent += fmt.Sprintf("RPROMPT=\"%%F{cyan}%s%%f $RPROMPT\"\n", promptSegment)

			if err := os.WriteFile(zshrcPath, []byte(zshContent), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not write .zshrc: %v\n", err)
			} else {
				newEnv = append(newEnv, fmt.Sprintf("ZDOTDIR=%s", tempDir))
			}
		} else {

			if baseShell == "bash" && tempDir != "" {
				bashrcPath := filepath.Join(tempDir, ".bashrc")
				userBashrc := filepath.Join(os.Getenv("HOME"), ".bashrc")
				bashContent := ""
				if _, err := os.Stat(userBashrc); err == nil {
					bashContent += fmt.Sprintf("source %s\n", userBashrc)
				}
				bashContent += fmt.Sprintf("\nPS1=\"\\[\\033[0;36m\\]%s\\[\\033[0m\\] $PS1\"\n", promptSegment)

				if err := os.WriteFile(bashrcPath, []byte(bashContent), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Could not write .bashrc: %v\n", err)
				} else {

					newEnv = append(newEnv, fmt.Sprintf("PS1=\\[\\033[0;36m\\]%s\\[\\033[0m\\] $PS1", promptSegment))
				}
			}
		}

		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Env = newEnv

		fmt.Println()
		fmt.Print(Banner)

		summary := []string{
			fmt.Sprintf("Client:     %s", ctx.Client),
			fmt.Sprintf("Engagement: %s", ctx.Engagement),
			fmt.Sprintf("Scope:      %s", ctx.Scope),
			fmt.Sprintf("Operator:   %s", ctx.Operator),
			fmt.Sprintf("Phase:      %s", ctx.Phase),
		}
		utils.PrintBox("Active Session", summary)

		fmt.Println()
		utils.PrintCenteredBlock([]string{"Type 'exit' or Ctrl+D to stop recording."})

		if err := c.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() != 0 {
					fmt.Println("\nLeaving pentlog shell session.")
					return
				}
			} else {
				fmt.Fprintf(os.Stderr, "Error running recorder: %v\n", err)
				return
			}
		}
		fmt.Println("\nLeaving pentlog shell session.")
	},
}

func init() {
	rootCmd.AddCommand(shellCmd)
}

func writeMetadata(path string, meta logs.SessionMetadata) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(meta)
}
