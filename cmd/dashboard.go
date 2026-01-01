package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/dashboard"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show an interactive dashboard of your pentest activity",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(dashboard.InitialModel())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running dashboard: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
