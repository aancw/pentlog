package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var quickNoteCmd = &cobra.Command{
	Use:    "quicknote",
	Short:  "Quick note entry (designed for hotkey use)",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		sessionDir := os.Getenv("PENTLOG_SESSION_DIR")
		if sessionDir == "" {
			errors.NewError(errors.NoActiveContext, "Not in a pentlog session").Fatal()
		}

		tty, err := os.Open("/dev/tty")
		if err != nil {
			errors.NewError(errors.Generic, "Cannot open terminal").Fatal()
		}
		defer tty.Close()

		reader := bufio.NewReader(tty)
		fmt.Print("\033[1;36m📝 Quick note:\033[0m ")
		message, err := reader.ReadString('\n')
		if err != nil {
			os.Exit(1)
		}
		message = strings.TrimSpace(message)
		if message == "" {
			fmt.Println("\033[33m⚠ Empty note discarded\033[0m")
			return
		}

		timestamp := time.Now().Format("15:04:05")

		logPath := os.Getenv("PENTLOG_SESSION_LOG_PATH")
		var byteOffset int64 = -1
		if logPath != "" {
			if info, err := os.Stat(logPath); err == nil {
				byteOffset = info.Size()
			}
		}

		note := logs.SessionNote{
			Timestamp:  timestamp,
			Content:    message,
			ByteOffset: byteOffset,
		}

		metaPath := os.Getenv("PENTLOG_SESSION_METADATA_PATH")
		var notesPath string
		if metaPath != "" {
			notesPath = strings.TrimSuffix(metaPath, ".json") + ".notes.json"
		} else {
			notesPath = filepath.Join(sessionDir, "notes.json")
		}

		if err := logs.AppendNote(notesPath, note); err != nil {
			errors.FileErr(notesPath, err).Fatal()
		}

		fmt.Printf("\033[32m✓ Note saved\033[0m [%s]\n", timestamp)
	},
}

func init() {
	rootCmd.AddCommand(quickNoteCmd)
}
