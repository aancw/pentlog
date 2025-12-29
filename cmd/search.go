package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/search"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search command history across all sessions (supports Regex)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]

		fmt.Printf("Searching for %q...\n", query)
		results, err := search.Search(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No matches found.")
			return
		}

		for _, match := range results {
			fmt.Printf("[%d] %s:%d: %s\n", match.Session.ID, match.Session.DisplayPath, match.LineNum, match.Content)
		}
		fmt.Printf("\nFound %d matches.\n", len(results))
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
