package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"pentlog/pkg/config"
	"pentlog/pkg/vulns"

	"github.com/spf13/cobra"
)

var quickVulnCmd = &cobra.Command{
	Use:    "quickvuln",
	Short:  "Quick vulnerability entry (designed for hotkey use)",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		sessionDir := os.Getenv("PENTLOG_SESSION_DIR")
		if sessionDir == "" {
			fmt.Fprintln(os.Stderr, "Error: Not in a pentlog session.")
			os.Exit(1)
		}

		mgr := config.Manager()
		ctx, err := mgr.LoadContext()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Not in an active engagement.")
			os.Exit(1)
		}

		manager, err := vulns.NewManagerFromContext()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		tty, err := os.Open("/dev/tty")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Cannot open terminal.")
			os.Exit(1)
		}
		defer tty.Close()

		reader := bufio.NewReader(tty)

		fmt.Print("\033[1;31m🔓 Vuln title:\033[0m ")
		title, err := reader.ReadString('\n')
		if err != nil {
			os.Exit(1)
		}
		title = strings.TrimSpace(title)
		if title == "" {
			fmt.Println("\033[33m⚠ Empty vuln discarded\033[0m")
			return
		}

		fmt.Print("\033[1;33mSeverity (c/h/m/l/i):\033[0m ")
		sevInput, err := reader.ReadString('\n')
		if err != nil {
			os.Exit(1)
		}
		sevInput = strings.ToLower(strings.TrimSpace(sevInput))

		var severity vulns.Severity
		switch sevInput {
		case "c", "critical":
			severity = vulns.SeverityCritical
		case "h", "high":
			severity = vulns.SeverityHigh
		case "m", "medium", "med":
			severity = vulns.SeverityMedium
		case "l", "low":
			severity = vulns.SeverityLow
		case "i", "info", "":
			severity = vulns.SeverityInfo
		default:
			severity = vulns.SeverityInfo
		}

		fmt.Print("\033[1;37mDescription (optional):\033[0m ")
		desc, _ := reader.ReadString('\n')
		desc = strings.TrimSpace(desc)

		id, _ := manager.GenerateID(title)
		vuln := vulns.Vuln{
			ID:          id,
			Title:       title,
			Severity:    severity,
			Status:      vulns.StatusOpen,
			Description: desc,
			Phase:       ctx.Phase,
		}

		if err := manager.Save(vuln); err != nil {
			fmt.Fprintf(os.Stderr, "\033[31m✗ Error: %v\033[0m\n", err)
			os.Exit(1)
		}

		fmt.Printf("\033[32m✓ Vuln saved:\033[0m %s [%s] %s\n", id, severity, title)
	},
}

func init() {
	rootCmd.AddCommand(quickVulnCmd)
}
