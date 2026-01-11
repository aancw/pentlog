package cmd

import (
	"fmt"
	"pentlog/pkg/metadata"
	"pentlog/pkg/system"
	"pentlog/pkg/utils"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current tool and engagement status",
	Run: func(cmd *cobra.Command, args []string) {
		lines := []string{}

		ctx, err := metadata.Load()
		if err != nil {
			lines = append(lines, "Context: No active engagement found.")
		} else {
			lines = append(lines, "Context:    ACTIVE")
			if ctx.Type == "Exam/Lab" {
				lines = append(lines, fmt.Sprintf("Exam/Lab Name: %s", ctx.Client))
				lines = append(lines, fmt.Sprintf("Target:        %s", ctx.Engagement))
			} else {
				lines = append(lines, fmt.Sprintf("Client:     %s", ctx.Client))
				lines = append(lines, fmt.Sprintf("Engagement: %s", ctx.Engagement))
				lines = append(lines, fmt.Sprintf("Scope:      %s", ctx.Scope))
			}
			lines = append(lines, fmt.Sprintf("Operator:   %s", ctx.Operator))
			lines = append(lines, fmt.Sprintf("Phase:      %s", ctx.Phase))
		}

		lines = append(lines, "")

		if err := system.CheckDependencies(); err != nil {
			lines = append(lines, fmt.Sprintf("Dependencies: MISSING (%v)", err))
		} else {
			lines = append(lines, "Dependencies: OK (ttyrec, ttyplay)")
		}

		utils.PrintBox("Pentlog Status", lines)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
