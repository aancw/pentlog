package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var noteCmd = &cobra.Command{
	Use:   "note",
	Short: "Manage session notes",
	Long:  `Add or list notes for the current running session. This command relies on environment variables set by 'pentlog shell'.`,
}

var noteAddCmd = &cobra.Command{
	Use:   "add [message]",
	Short: "Add a timestamped note to the current session",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		sessionDir := os.Getenv("PENTLOG_SESSION_DIR")
		if sessionDir == "" {
			errors.NewError(errors.NoActiveContext, "Not currently in an active pentlog session").WithDetails("PENTLOG_SESSION_DIR not set").Fatal()
		}

		message := strings.Join(args, " ")
		timestamp := time.Now().Format("15:04:05")

		logPath := os.Getenv("PENTLOG_SESSION_LOG_PATH")
		var byteOffset int64 = -1
		if logPath != "" {
			info, err := os.Stat(logPath)
			if err == nil {
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

		if byteOffset != -1 {
			fmt.Printf("Note added: [%s] (offset: %d) %s\n", timestamp, byteOffset, message)
		} else {
			fmt.Printf("Note added: [%s] %s\n", timestamp, message)

		}
	},
}

var noteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes in the current session (Interactive)",
	Run: func(cmd *cobra.Command, args []string) {
		sessionDir := os.Getenv("PENTLOG_SESSION_DIR")

		if sessionDir != "" {
			metaPath := os.Getenv("PENTLOG_SESSION_METADATA_PATH")
			logPath := os.Getenv("PENTLOG_SESSION_LOG_PATH")
			var notesPath string
			if metaPath != "" {
				notesPath = strings.TrimSuffix(metaPath, ".json") + ".notes.json"
			} else {
				notesPath = filepath.Join(sessionDir, "notes.json")
			}
			runInteractiveNoteList(notesPath, logPath, false)
			return
		}

		for {
			sessions, err := logs.ListSessions()
			if err != nil {
				errors.DatabaseErr("list sessions", err).Fatal()
			}
			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			templates := &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "\U000025B6 [{{ .ID | cyan }}] {{ .Metadata.Client }} / {{ .Metadata.Phase }} ({{ .ModTime }})",
				Inactive: "  [{{ .ID | faint }}] {{ .Metadata.Client }} / {{ .Metadata.Phase }} ({{ .ModTime }})",
				Selected: "\U000025B6 Selected Session: {{ .ID | cyan }}",
			}

			prompt := promptui.Select{
				Label:     "Select Session to Review Notes (Ctrl+C to Exit)",
				Items:     sessions,
				Templates: templates,
				Size:      10,
			}

			i, _, err := prompt.Run()
			if err != nil {
				if err == promptui.ErrInterrupt {
					os.Exit(0)
				}
				return
			}

			selectedSession := sessions[i]
			var notesPath string
			if selectedSession.NotesPath != "" {
				notesPath = selectedSession.NotesPath
			} else {
				ext := filepath.Ext(selectedSession.Path)
				if ext == ".tty" {
					notesPath = strings.TrimSuffix(selectedSession.Path, ".tty") + ".notes.json"
				} else {
					info, err := os.Stat(selectedSession.Path)
					if err == nil && !info.IsDir() {
						notesPath = strings.TrimSuffix(selectedSession.Path, ext) + ".notes.json"
					} else {
						notesPath = filepath.Join(selectedSession.Path, "notes.json")
					}
				}
			}
			runInteractiveNoteList(notesPath, selectedSession.Path, true)
		}
	},
}

func runInteractiveNoteList(notesPath, logPath string, allowBack bool) {
	for {
		if notesPath == "" {
			fmt.Println("No notes available for this session.")
			if allowBack {
				waitForEnter("Press Enter to go back...")
				return
			}
			os.Exit(0)
		}

		notes, err := logs.ReadNotes(notesPath)
		if err != nil {
			errors.FileErr(notesPath, err).Print()
			if allowBack {
				waitForEnter("Press Enter to go back...")
				return
			}
			os.Exit(1)
		}

		if len(notes) == 0 {
			fmt.Println("No notes found for this session.")
			if allowBack {
				waitForEnter("Press Enter to go back...")
				return
			}
			return
		}

		info, _ := os.Stdout.Stat()
		isInteractive := (info.Mode() & os.ModeCharDevice) != 0
		if !isInteractive {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TIMESTAMP\tOFFSET\tNOTE")
			for _, n := range notes {
				offsetStr := "N/A"
				if n.ByteOffset != -1 {
					offsetStr = fmt.Sprintf("%d", n.ByteOffset)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", n.Timestamp, offsetStr, n.Content)
			}
			w.Flush()
			return
		}

		type item struct {
			Label    string
			IsBack   bool
			Original logs.SessionNote
		}

		var items []item
		for _, n := range notes {
			items = append(items, item{
				Label:    fmt.Sprintf("[%s] %s", n.Timestamp, n.Content),
				Original: n,
			})
		}

		if allowBack {
			items = append(items, item{Label: "Start Over (Back)", IsBack: true})
		}

		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "\U000025B6 {{ .Label | cyan }}",
			Inactive: "  {{ .Label }}",
			Selected: "\U000025B6 {{ .Label | cyan }}",
		}

		label := "Select a note to preview context"
		if allowBack {
			label += " (Ctrl+C to Exit)"
		}

		prompt := promptui.Select{
			Label:     label,
			Items:     items,
			Templates: templates,
			Size:      10,
		}

		i, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				os.Exit(0)
			}
			return
		}

		selection := items[i]
		if selection.IsBack {
			return
		}

		previewLogContext(selection.Original, logPath)

		waitForEnter("Press Enter to continue...")
	}
}

func waitForEnter(msg string) {
	fmt.Println(msg)
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

type CountingReader struct {
	io.Reader
	N int64
}

func (r *CountingReader) Read(p []byte) (n int, err error) {
	n, err = r.Reader.Read(p)
	r.N += int64(n)
	return n, err
}

func previewLogContext(note logs.SessionNote, logPath string) {
	if logPath == "" {
		fmt.Println("Log path not available for preview.")
		return
	}

	f, err := os.Open(logPath)
	if err != nil {
		fmt.Printf("Error opening log: %v\n", err)
		return
	}
	defer f.Close()

	limitReader := io.LimitReader(f, note.ByteOffset)

	var textReader io.Reader = limitReader
	if strings.HasSuffix(logPath, ".tty") {
		textReader = logs.NewTtyReader(limitReader)
	}

	fmt.Printf("\n--- Context Preview (Offset %d) ---\n", note.ByteOffset)

	cleaner := utils.NewCleanReader(textReader)

	scanner := bufio.NewScanner(cleaner)

	maxLines := 15
	ringBuffer := make([]string, 0, maxLines)

	for scanner.Scan() {

		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(ringBuffer) < maxLines {
			ringBuffer = append(ringBuffer, line)
		} else {
			// Shift left
			ringBuffer = append(ringBuffer[1:], line)
		}
	}

	for _, line := range ringBuffer {
		fmt.Println(line)
	}

	fmt.Println("-----------------------------------")
}

func init() {
	rootCmd.AddCommand(noteCmd)
	noteCmd.AddCommand(noteAddCmd)
	noteCmd.AddCommand(noteListCmd)
}
