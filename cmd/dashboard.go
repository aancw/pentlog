package cmd

import (
	"pentlog/pkg/dashboard"
	"pentlog/pkg/errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show an interactive dashboard of your pentest activity",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(dashboard.InitialModel())
		if _, err := p.Run(); err != nil {
			errors.FromError(errors.Generic, "Error running dashboard", err).Fatal()
		}
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
