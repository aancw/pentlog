package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/metadata"
	"pentlog/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

var (
	startClient     string
	startEngagement string
	startScope      string
	startOperator   string
	startPhase      string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new engagement context",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := metadata.Context{
			Client:     startClient,
			Engagement: startEngagement,
			Scope:      startScope,
			Operator:   startOperator,
			Phase:      startPhase,
			Timestamp:  time.Now().Format(time.RFC3339),
		}

		if err := metadata.Save(ctx); err != nil {
			fmt.Printf("Error saving context: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Context saved and history updated.")

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
		fmt.Println("To enter this context immediately, run:")
		fmt.Println("  ./pentlog shell")
		fmt.Println("\nOr for scripting:")
		fmt.Println("  eval $(./pentlog env)")
	},
}

func init() {
	startCmd.Flags().StringVarP(&startClient, "client", "c", "", "Client name")
	startCmd.Flags().StringVarP(&startEngagement, "engagement", "e", "", "Engagement type/name")
	startCmd.Flags().StringVarP(&startScope, "scope", "s", "", "Scope definition")
	startCmd.Flags().StringVarP(&startOperator, "operator", "o", "", "Operator name")
	startCmd.Flags().StringVarP(&startPhase, "phase", "p", "generic", "Pentest phase (recon, exploit, post, pivot, cleanup)")

	startCmd.MarkFlagRequired("client")
	startCmd.MarkFlagRequired("engagement")
	startCmd.MarkFlagRequired("operator")

	rootCmd.AddCommand(startCmd)
}
