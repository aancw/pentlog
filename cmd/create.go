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
	createClient     string
	createEngagement string
	createScope      string
	createOperator   string
	createPhase      string
)

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Initialize a new engagement context (Interactive)",
	Aliases: []string{"init"},
	Run: func(cmd *cobra.Command, args []string) {
		if createClient == "" {
			createClient = utils.PromptString("Client Name", "")
		}
		if createEngagement == "" {
			createEngagement = utils.PromptString("Engagement", "")
		}
		if createScope == "" {
			createScope = utils.PromptString("Scope (CIDR/URL)", "")
		}
		if createOperator == "" {
			createOperator = utils.PromptString("Operator", os.Getenv("USER"))
		}
		if createPhase == "" {
			createPhase = utils.PromptString("Phase", "recon")
		}

		if createClient == "" || createEngagement == "" || createOperator == "" {
			fmt.Println("Error: Client, Engagement, and Operator are required.")
			os.Exit(1)
		}

		ctx := metadata.Context{
			Client:     createClient,
			Engagement: createEngagement,
			Scope:      createScope,
			Operator:   createOperator,
			Phase:      createPhase,
			Timestamp:  time.Now().Format(time.RFC3339),
		}

		if err := metadata.Save(ctx); err != nil {
			fmt.Printf("Error saving context: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\nContext initialized successfully!")

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
		fmt.Println("To start recording, run:")
		fmt.Println("  ./pentlog shell")
	},
}

func init() {
	createCmd.Flags().StringVarP(&createClient, "client", "c", "", "Client name")
	createCmd.Flags().StringVarP(&createEngagement, "engagement", "e", "", "Engagement type/name")
	createCmd.Flags().StringVarP(&createScope, "scope", "s", "", "Scope definition")
	createCmd.Flags().StringVarP(&createOperator, "operator", "o", "", "Operator name")
	createCmd.Flags().StringVarP(&createPhase, "phase", "p", "", "Pentest phase (recon, exploit, post, pivot, cleanup)")

	rootCmd.AddCommand(createCmd)
}
