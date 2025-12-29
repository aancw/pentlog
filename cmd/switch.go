package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/metadata"
	"pentlog/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [phase]",
	Short: "Switch to a different pentest phase (Interactive)",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, err := metadata.Load()
		if err != nil {
			fmt.Println("Error: No active engagement found. Run 'pentlog create' first.")
			os.Exit(1)
		}

		newPhase := ""
		if len(args) > 0 {
			newPhase = args[0]
		}

		if newPhase == "" {
			fmt.Printf("Current Phase: %s\n", ctx.Phase)
			newPhase = utils.PromptString("New Phase", "")
		}

		if newPhase == "" {
			fmt.Println("Error: Phase cannot be empty.")
			os.Exit(1)
		}

		ctx.Phase = newPhase
		ctx.Timestamp = time.Now().Format(time.RFC3339)

		if err := metadata.Save(*ctx); err != nil {
			fmt.Printf("Error saving context: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nSwitched to phase: %s\n", ctx.Phase)

		summary := []string{
			"---------------------------------------------------",
			fmt.Sprintf("Client:     %s", ctx.Client),
			fmt.Sprintf("Engagement: %s", ctx.Engagement),
			fmt.Sprintf("Scope:      %s", ctx.Scope),
			fmt.Sprintf("Operator:   %s", ctx.Operator),
			fmt.Sprintf("Phase:      %s", ctx.Phase),
			"---------------------------------------------------",
		}
		utils.PrintCenteredBlock(summary)
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
