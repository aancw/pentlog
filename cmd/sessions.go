package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	sessionsLimit  int
	sessionsOffset int
)

const defaultSessionsPageSize = 20

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
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

func init() {
	sessionsCmd.Flags().IntVarP(&sessionsLimit, "limit", "l", 0, "Maximum number of sessions to display")
	sessionsCmd.Flags().IntVarP(&sessionsOffset, "offset", "o", 0, "Number of sessions to skip (for pagination)")
	rootCmd.AddCommand(sessionsCmd)
}

func printSessionsTable(sessions []logs.Session) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTIME\tSIZE\tFILE")
	for _, s := range sessions {
		fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", s.ID, s.ModTime, s.Size, s.DisplayPath)
	}
	w.Flush()
}
