package cmd

import (
	"fmt"
	"os"
	"time"
	"pentlog/pkg/metadata"

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
		fmt.Println("---------------------------------------------------")
		fmt.Printf("Client:     %s\n", ctx.Client)
		fmt.Printf("Engagement: %s\n", ctx.Engagement)
		fmt.Printf("Scope:      %s\n", ctx.Scope)
		fmt.Printf("Operator:   %s\n", ctx.Operator)
		fmt.Printf("Phase:      %s\n", ctx.Phase)
		fmt.Println("---------------------------------------------------")
		fmt.Println("To ensure metadata is recorded in current session logs, run:")
		fmt.Printf("export PENTLOG_CLIENT='%s'\n", ctx.Client)
		fmt.Printf("export PENTLOG_ENGAGEMENT='%s'\n", ctx.Engagement)
		fmt.Printf("export PENTLOG_SCOPE='%s'\n", ctx.Scope)
		fmt.Printf("export PENTLOG_OPERATOR='%s'\n", ctx.Operator)
		fmt.Printf("export PENTLOG_PHASE='%s'\n", ctx.Phase)
	},
}

func init() {
	startCmd.Flags().StringVar(&startClient, "client", "", "Client name")
	startCmd.Flags().StringVar(&startEngagement, "engagement", "", "Engagement type/name")
	startCmd.Flags().StringVar(&startScope, "scope", "", "Scope definition")
	startCmd.Flags().StringVar(&startOperator, "operator", "", "Operator name")
	startCmd.Flags().StringVar(&startPhase, "phase", "generic", "Pentest phase (recon, exploit, post, pivot, cleanup)")

	startCmd.MarkFlagRequired("client")
	startCmd.MarkFlagRequired("engagement")
	startCmd.MarkFlagRequired("operator")
	
	rootCmd.AddCommand(startCmd)
}