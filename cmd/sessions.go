package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	sessionsLimit  int
	sessionsOffset int
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List recorded sessions",
	Run: func(cmd *cobra.Command, args []string) {
		var sessions []logs.Session
		var err error

		if sessionsLimit > 0 || sessionsOffset > 0 {
			sessions, err = logs.ListSessionsPaginated(sessionsLimit, sessionsOffset)
		} else {
			sessions, err = logs.ListSessions()
		}

		if err != nil {
			errors.DatabaseErr("list sessions", err).Fatal()
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTIME\tSIZE\tFILE")
		for _, s := range sessions {
			fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", s.ID, s.ModTime, s.Size, s.DisplayPath)
		}
		w.Flush()

		if sessionsLimit > 0 {
			fmt.Printf("\nShowing %d session(s)", len(sessions))
			if sessionsOffset > 0 {
				fmt.Printf(" (offset: %d)", sessionsOffset)
			}
			fmt.Println()
		}
	},
}

func init() {
	sessionsCmd.Flags().IntVarP(&sessionsLimit, "limit", "l", 0, "Maximum number of sessions to display")
	sessionsCmd.Flags().IntVarP(&sessionsOffset, "offset", "o", 0, "Number of sessions to skip (for pagination)")
	rootCmd.AddCommand(sessionsCmd)
}
