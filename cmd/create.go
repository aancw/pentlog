package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/config"
	"pentlog/pkg/errors"
	"pentlog/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

var (
	createType       string
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
		if createType == "" {
			createType = utils.PromptSelect("Context Type", []string{"Client", "Exam/Lab", "Log Only"})
		}

		if createType == "Log Only" {
			// Log Only Mode - Minimal Interaction
			if createClient == "" {
				createClient = utils.PromptString("Project Name", "QuickLog")
			}
			if createEngagement == "" {
				createEngagement = "Session"
			}
			if createScope == "" {
				createScope = "N/A"
			}
			if createOperator == "" {
				createOperator = os.Getenv("USER")
			}
			if createPhase == "" {
				createPhase = "N/A"
			}
		} else if createType == "Exam/Lab" || createType == "Exam" || createType == "Lab" {
			// Normalize to Exam/Lab
			createType = "Exam/Lab"

			if createClient == "" {
				createClient = utils.PromptString("Exam/Lab Name", "")
			}
			if createEngagement == "" {
				createEngagement = utils.PromptString("Target Host/IP", "")
			}
			// Scope is optional or same as Target in Exam, but let's keep it consistent or auto-fill
			if createScope == "" {
				createScope = createEngagement
			}

			if createOperator == "" {
				createOperator = utils.PromptString("Operator", os.Getenv("USER"))
			}
			if createPhase == "" {
				createPhase = utils.PromptString("Phase", "recon")
			}
		} else {
			// Client Mode (Default)
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
		}

		if createClient == "" || createEngagement == "" || createOperator == "" {
			errors.NewError(errors.InvalidContext, "Required fields missing").
				AddReason("Client name not provided").
				AddReason("Engagement name not provided").
				AddReason("Operator name not provided").
				AddSolution("Re-run pentlog create with all required information").
				Fatal()
		}

		ctx := &config.ContextData{
			Client:     createClient,
			Engagement: createEngagement,
			Scope:      createScope,
			Operator:   createOperator,
			Phase:      createPhase,
			Timestamp:  time.Now().Format(time.RFC3339),
			Type:       createType,
		}

		mgr := config.Manager()
		if err := mgr.SaveContext(ctx); err != nil {
			errors.FromError(errors.Generic, "Failed to save engagement context", err).Fatal()
		}

		fmt.Println("\nContext initialized successfully!")

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
		fmt.Println("To start recording, run:")
		fmt.Println(" pentlog shell")
	},
}

func init() {
	createCmd.Flags().StringVarP(&createType, "type", "t", "", "Context type (Client/Exam/Lab)")
	createCmd.Flags().StringVarP(&createClient, "client", "c", "", "Client name / Exam/Lab Name")
	createCmd.Flags().StringVarP(&createEngagement, "engagement", "e", "", "Engagement type/name")
	createCmd.Flags().StringVarP(&createScope, "scope", "s", "", "Scope definition")
	createCmd.Flags().StringVarP(&createOperator, "operator", "o", "", "Operator name")
	createCmd.Flags().StringVarP(&createPhase, "phase", "p", "", "Pentest phase (recon, exploit, post, pivot, cleanup)")

	rootCmd.AddCommand(createCmd)
}
