package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/search"
	"pentlog/pkg/utils"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search command history across all sessions (supports Regex)",
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		if len(args) > 0 {
			query = args[0]
		}

		if query == "" {
			query = utils.PromptString("Search Query (Regex)", "")
		}

		if query == "" {
			fmt.Println("Error: Search query cannot be empty.")
			os.Exit(1)
		}

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
			if match.IsNote {
				fmt.Printf("[%d] %s [NOTE]: %s\n", match.Session.ID, match.Session.DisplayPath, match.Content)
			} else {
				fmt.Printf("[%d] %s:%d: %s\n", match.Session.ID, match.Session.DisplayPath, match.LineNum, match.Content)
			}
		}
		fmt.Printf("\nFound %d matches.\n", len(results))
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
