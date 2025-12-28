package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/metadata"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Output environment variables for the current context",
	Long:  `Outputs export commands for the currently active engagement context.
Usage: eval $(pentlog env)`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx, err := metadata.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading context: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("export PENTLOG_CLIENT='%s'\n", ctx.Client)
		fmt.Printf("export PENTLOG_ENGAGEMENT='%s'\n", ctx.Engagement)
		fmt.Printf("export PENTLOG_SCOPE='%s'\n", ctx.Scope)
		fmt.Printf("export PENTLOG_OPERATOR='%s'\n", ctx.Operator)
		fmt.Printf("export PENTLOG_PHASE='%s'\n", ctx.Phase)
	},
}

func init() {
	rootCmd.AddCommand(envCmd)
}