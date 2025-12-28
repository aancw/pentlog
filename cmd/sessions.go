package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/logs"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List recorded sessions",
	Run: func(cmd *cobra.Command, args []string) {
		sessions, err := logs.ListSessions()
		if err != nil {
			fmt.Printf("Error listing sessions: %v\n", err)
			os.Exit(1)
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTIME\tSIZE\tFILENAME")
		for _, s := range sessions {
			fmt.Fprintf(w, "%d\t%s\t%d\t%s\n", s.ID, s.ModTime, s.Size, s.Filename)
		}
		w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(sessionsCmd)
}

