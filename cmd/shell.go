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
		baseName := fmt.Sprintf("manual-%s-%s", utils.Slugify(ctx.Operator), timestamp)
		logFilePath := filepath.Join(sessionDir, baseName+".log")
		timingFilePath := filepath.Join(sessionDir, baseName+".timing")
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

		scriptPath, err := exec.LookPath("script")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: 'script' command not found.\n")
			os.Exit(1)
		}

		c := exec.Command(scriptPath, "--quiet", "--flush", "--append", "-t", timingFilePath, logFilePath)
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

		fmt.Printf("Starting RECORDED session for: %s [%s]\n", ctx.Client, ctx.Phase)
		fmt.Printf("Log file: %s\n", logFilePath)
		fmt.Printf("Timing file: %s\n", timingFilePath)
		fmt.Println("Type 'exit' or Ctrl+D to stop recording.")

		if err := c.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				os.Exit(exitError.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "Error running recorder: %v\n", err)
			os.Exit(1)
		}
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
