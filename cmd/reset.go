package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/config"
	"pentlog/pkg/errors"

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
			errors.FileErr(path, err).Fatal()
		}

		fmt.Println("Active context cleared. You can now start a new engagement.")
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

