package cmd

import (
	"os"
	"os/exec"
	"pentlog/pkg/dashboard"
	"pentlog/pkg/errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show an interactive dashboard of your pentest activity",
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(dashboard.InitialModel(), tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			errors.FromError(errors.Generic, "Error running dashboard", err).Fatal()
		}

		model, ok := finalModel.(dashboard.Model)
		if !ok {
			return
		}

		action := model.LaunchArgs()
		if len(action) == 0 {
			return
		}

		subCmd := exec.Command(os.Args[0], action...)
		subCmd.Stdin = os.Stdin
		subCmd.Stdout = os.Stdout
		subCmd.Stderr = os.Stderr
		subCmd.Env = os.Environ()

		if err := subCmd.Run(); err != nil {
			errors.FromError(errors.Generic, "Error running dashboard action", err).Fatal()
		}
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
