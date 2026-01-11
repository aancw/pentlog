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
			if ctx.Type == "Exam/Lab" {
				fmt.Printf("Current Target: %s\n", ctx.Engagement)
				newTarget := utils.PromptString("New Target Host/IP", ctx.Engagement)
				if newTarget != "" {
					ctx.Engagement = newTarget
					// Also update scope to match target
					ctx.Scope = newTarget
				}
			}

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
		}

		if ctx.Type == "Exam/Lab" {
			summary = append(summary, fmt.Sprintf("Exam/Lab Name: %s", ctx.Client))
			summary = append(summary, fmt.Sprintf("Target:        %s", ctx.Engagement))
		} else {
			summary = append(summary, fmt.Sprintf("Client:     %s", ctx.Client))
			summary = append(summary, fmt.Sprintf("Engagement: %s", ctx.Engagement))
			summary = append(summary, fmt.Sprintf("Scope:      %s", ctx.Scope))
		}
		summary = append(summary, fmt.Sprintf("Operator:   %s", ctx.Operator))
		summary = append(summary, fmt.Sprintf("Phase:      %s", ctx.Phase))
		summary = append(summary, "---------------------------------------------------")
		utils.PrintCenteredBlock(summary)
	},
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
