package cmd

import (
	"fmt"
	"pentlog/pkg/config"
	"pentlog/pkg/deps"
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

		mgr := config.Manager()
		ctx, err := mgr.LoadContext()
		if err != nil {
			lines = append(lines, "Context: No active engagement found.")
		} else {
			lines = append(lines, "Context:    ACTIVE")
			lines = append(lines, buildContextSummaryLines(*ctx)...)
			lines = append(lines, fmt.Sprintf("Context Age: %s", formatContextAgeDetail(ctx.Timestamp)))

			recentChanges, historyErr := loadRecentContextChanges(mgr, 3)
			if historyErr == nil && len(recentChanges) > 0 {
				lines = append(lines, "")
				lines = append(lines, "Recent Changes:")
				for _, change := range recentChanges {
					lines = append(lines, "  - "+change)
				}
			}
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
