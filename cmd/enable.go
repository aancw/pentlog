package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Legacy command (no longer required)",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("PAM-based recording has been removed. Run 'pentlog shell' after connecting to start a recorded session.")
	},
}

func init() {
	rootCmd.AddCommand(enableCmd)
}
