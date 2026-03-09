package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	sessionsLimit  int
	sessionsOffset int
)

const defaultSessionsPageSize = 20

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage recorded sessions",
	Long:  `List or delete recorded sessions. Use 'sessions list' to view sessions or 'sessions delete <id>' to remove a session.`,
}

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recorded sessions",
	Run: func(cmd *cobra.Command, args []string) {
		var sessions []logs.Session
		var err error

		if sessionsLimit > 0 || sessionsOffset > 0 {
			sessions, err = logs.ListSessionsPaginated(sessionsLimit, sessionsOffset)
			if err != nil {
				errors.DatabaseErr("list sessions", err).Fatal()
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			printSessionsTable(sessions)

			if sessionsLimit > 0 {
				fmt.Printf("\nShowing %d session(s)", len(sessions))
				if sessionsOffset > 0 {
					fmt.Printf(" (offset: %d)", sessionsOffset)
				}
				fmt.Println()
			}

			return
		}

		offset := 0
		for {
			sessions, err = logs.ListSessionsPaginated(defaultSessionsPageSize+1, offset)
			if err != nil {
				errors.DatabaseErr("list sessions", err).Fatal()
			}

			if len(sessions) == 0 {
				if offset == 0 {
					fmt.Println("No sessions found.")
				}
				return
			}

			hasMore := len(sessions) > defaultSessionsPageSize
			if hasMore {
				sessions = sessions[:defaultSessionsPageSize]
			}

			printSessionsTable(sessions)

			if !hasMore {
				return
			}

			answer := utils.PromptString("Show more? (Y/n)", "Y")
			if !strings.EqualFold(strings.TrimSpace(answer), "y") {
				return
			}

			offset += defaultSessionsPageSize
			fmt.Println()
		}
	},
}

var sessionsDeleteCmd = &cobra.Command{
	Use:   "delete [session-id]",
	Short: "Delete a session by ID",
	Long:  `Delete a session and its associated files (.tty, .json, .notes.json) by session ID. Use 'sessions list' to find the ID.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var sessionID int
		var err error

		if len(args) == 0 {
			// Interactive mode - show list and prompt for selection
			sessions, err := logs.ListSessions()
			if err != nil {
				errors.DatabaseErr("list sessions", err).Fatal()
			}
			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			printSessionsTable(sessions)
			fmt.Println()

			prompt := promptui.Prompt{
				Label: "Enter session ID to delete",
				Validate: func(input string) error {
					id, err := strconv.Atoi(input)
					if err != nil {
						return fmt.Errorf("invalid session ID")
					}
					// Verify session exists
					for _, s := range sessions {
						if s.ID == id {
							return nil
						}
					}
					return fmt.Errorf("session ID not found")
				},
			}

			result, err := prompt.Run()
			if err != nil {
				fmt.Println("Cancelled.")
				return
			}

			sessionID, _ = strconv.Atoi(result)
		} else {
			sessionID, err = strconv.Atoi(args[0])
			if err != nil {
				errors.NewError(errors.InvalidInput, "Invalid session ID").WithDetails("Session ID must be a number").Fatal()
			}
		}

		// Get session details before deletion
		session, err := logs.GetSession(sessionID)
		if err != nil {
			errors.DatabaseErr("get session", err).Fatal()
		}

		// Confirm deletion
		confirmPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("Delete session %d (%s, %s)", session.ID, session.DisplayPath, utils.FormatBytes(session.Size)),
			IsConfirm: true,
		}

		_, err = confirmPrompt.Run()
		if err != nil {
			fmt.Println("Deletion cancelled.")
			return
		}

		// Delete associated files
		filesToDelete := []string{
			session.Path,
			session.Path + ".json",
			session.Path + ".notes.json",
		}

		deletedFiles := 0
		for _, file := range filesToDelete {
			if _, err := os.Stat(file); err == nil {
				if err := os.Remove(file); err == nil {
					deletedFiles++
				}
			}
		}

		// Delete from database
		if err := logs.DeleteSession(sessionID); err != nil {
			errors.DatabaseErr("delete session", err).Fatal()
		}

		fmt.Printf("✓ Deleted session %d (%d files removed)\n", sessionID, deletedFiles)
	},
}

func init() {
	sessionsListCmd.Flags().IntVarP(&sessionsLimit, "limit", "l", 0, "Maximum number of sessions to display")
	sessionsListCmd.Flags().IntVarP(&sessionsOffset, "offset", "o", 0, "Number of sessions to skip (for pagination)")

	rootCmd.AddCommand(sessionsCmd)
	sessionsCmd.AddCommand(sessionsListCmd)
	sessionsCmd.AddCommand(sessionsDeleteCmd)
}

func printSessionsTable(sessions []logs.Session) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTIME\tSIZE\tFILE")
	for _, s := range sessions {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", s.ID, s.ModTime, utils.FormatBytes(s.Size), s.DisplayPath)
	}
	w.Flush()
}
