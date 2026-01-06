package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"pentlog/pkg/metadata"
	"pentlog/pkg/vulns"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

const (
	markerTuiStart = "\x1b]99;PENTLOG_TUI_START\x07"
	markerTuiEnd   = "\x1b]99;PENTLOG_TUI_END\x07"
)

var (
	vulnCmd = &cobra.Command{
		Use:   "vuln",
		Short: "Manage findings and vulnerabilities",
		Long:  `Track findings, vulnerabilities, and misconfigurations found during the engagement.`,
	}

	vulnAddCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a new vuln",
		Run:   runVulnAdd,
	}

	vulnListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all vulns",
		Run:   runVulnList,
	}
)

func init() {
	rootCmd.AddCommand(vulnCmd)
	vulnCmd.AddCommand(vulnAddCmd)
	vulnCmd.AddCommand(vulnListCmd)
}

func runVulnAdd(cmd *cobra.Command, args []string) {
	ctx, err := metadata.Load()
	if err != nil {
		fmt.Println("Error: Not in an active engagement. Run 'pentlog create' or 'pentlog switch' first.")
		return
	}

	manager, err := vulns.NewManagerFromContext()
	if err != nil {
		fmt.Printf("Error initializing vulns manager: %v\n", err)
		return
	}

	var (
		title       string
		severity    string
		description string
	)

	severityOptions := []huh.Option[string]{
		huh.NewOption("Critical", string(vulns.SeverityCritical)),
		huh.NewOption("High", string(vulns.SeverityHigh)),
		huh.NewOption("Medium", string(vulns.SeverityMedium)),
		huh.NewOption("Low", string(vulns.SeverityLow)),
		huh.NewOption("Info", string(vulns.SeverityInfo)),
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Title").
				Placeholder("Input the title...").
				Value(&title),

			huh.NewSelect[string]().
				Title("Severity").
				Options(severityOptions...).
				Value(&severity),

			huh.NewText().
				Title("Description").
				Placeholder("Describe the issue...").
				Value(&description),
		),
	).WithTheme(huh.ThemeDracula())

	// Mark the start of TUI interaction for log cleaning
	fmt.Print(markerTuiStart)
	os.Stdout.Sync()
	err = form.Run()
	// Mark the end of TUI interaction
	fmt.Print(markerTuiEnd)
	os.Stdout.Sync()

	if err != nil {
		fmt.Println("Cancelled.")
		return
	}

	id, _ := manager.GenerateID(title)
	vuln := vulns.Vuln{
		ID:          id,
		Title:       title,
		Severity:    vulns.Severity(severity),
		Status:      vulns.StatusOpen,
		Description: description,
		Phase:       ctx.Phase,
	}

	if err := manager.Save(vuln); err != nil {
		fmt.Printf("Error saving vuln: %v\n", err)
		return
	}

	fmt.Printf("\n✅ Vuln saved: %s [%s] %s\n", id, severity, title)
}

func runVulnList(cmd *cobra.Command, args []string) {
	manager, err := vulns.NewManagerFromContext()
	if err != nil {
		fmt.Println("Error: Not in an active engagement.")
		return
	}

	list, err := manager.List()
	if err != nil {
		fmt.Printf("Error listing vulns: %v\n", err)
		return
	}

	if len(list) == 0 {
		fmt.Println("No vulns recorded yet.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tSEVERITY\tTITLE\tSTATUS")

	for _, v := range list {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			v.ID,
			v.Severity,
			v.Title,
			v.Status,
		)
	}
	w.Flush()
}
