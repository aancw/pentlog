package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/config"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Clear the current active engagement context",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := config.Manager()
		path := mgr.GetPaths().ContextFile

		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Println("No active context to reset.")
			return
		}

		if err := os.Remove(path); err != nil {
			fmt.Printf("Error removing context file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Active context cleared. You can now start a new engagement.")
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

