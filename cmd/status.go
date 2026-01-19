package cmd

import (
	"fmt"
	"pentlog/pkg/deps"
	"pentlog/pkg/metadata"
	"pentlog/pkg/utils"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current tool and engagement status",
	Run: func(cmd *cobra.Command, args []string) {
		lines := []string{}

		depManager := deps.NewManager()
		depLines := []string{}
		allDepsOK := true

		for _, dep := range depManager.Dependencies {
			ok, path := depManager.Check(dep.Name)
			if !ok {
				allDepsOK = false
				depLines = append(depLines, fmt.Sprintf("%-10s : MISSING", dep.Name))
			} else {
				depLines = append(depLines, fmt.Sprintf("%-10s : OK (%s)", dep.Name, path))
			}
		}

		if showDepsOnly {
			utils.PrintBox("Dependency Check", depLines)
			return
		}

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

		if allDepsOK {
			lines = append(lines, "Dependencies: OK")
		} else {
			lines = append(lines, "Dependencies: ISSUES FOUND (use --dependencies for details)")
		}

		utils.PrintBox("Pentlog Status", lines)
	},
}

var showDepsOnly bool

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVar(&showDepsOnly, "dependencies", false, "Show detailed dependency health")
}
