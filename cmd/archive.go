package cmd

import (
	"bufio"
	"fmt"
	"os"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var (
	daysFlag       int
	deleteFlag     bool
	forceFlag      bool
	engagementFlag string
	phaseFlag      string
)

var archiveCmd = &cobra.Command{
	Use:   "archive [client]",
	Short: "Archive old sessions to save space",
	Long: `Archive moves session logs to a compressed archive file.
By default, original files are KEPT (copied to archive).
Use --delete to remove original files after archiving.

Examples:
  pentlog archive                 # Interactive mode
  pentlog archive acme            # Archive all 'acme' sessions (keep originals)
  pentlog archive acme -D         # Archive and DELETE originals
  pentlog archive acme -p recon   # Archive only 'recon' phase sessions
  pentlog archive acme -d 30      # Archive sessions older than 30 days`,
	Run: func(cmd *cobra.Command, args []string) {
		var clientName string
		var engagement, phase string
		var days int
		var deleteOrg bool

		if len(args) == 0 {
			sessions, err := logs.ListSessions()
			if err != nil {
				fmt.Printf("Error listing sessions: %v\n", err)
				os.Exit(1)
			}
			if len(sessions) == 0 {
				fmt.Println("No sessions found to archive.")
				return
			}

			clientMap := make(map[string]bool)
			var clients []string
			for _, s := range sessions {
				if s.Metadata.Client != "" && !clientMap[s.Metadata.Client] {
					clientMap[s.Metadata.Client] = true
					clients = append(clients, s.Metadata.Client)
				}
			}

			if len(clients) == 0 {
				fmt.Println("No clients found in sessions.")
				return
			}

			idx := utils.SelectItem("Select Client to Archive:", clients)
			if idx == -1 {
				return
			}
			clientName = clients[idx]

			engMap := make(map[string]bool)
			var engagements []string
			for _, s := range sessions {
				if s.Metadata.Client == clientName && s.Metadata.Engagement != "" && !engMap[s.Metadata.Engagement] {
					engMap[s.Metadata.Engagement] = true
					engagements = append(engagements, s.Metadata.Engagement)
				}
			}

			if len(engagements) > 0 {
				opts := append([]string{"All Engagements"}, engagements...)
				idx := utils.SelectItem("Filter by Engagement? (Select 'All' to skip):", opts)
				if idx > 0 {
					engagement = engagements[idx-1]
				}
			}

			phaseMap := make(map[string]bool)
			var phases []string
			for _, s := range sessions {
				if s.Metadata.Client != clientName {
					continue
				}
				if engagement != "" && s.Metadata.Engagement != engagement {
					continue
				}

				if s.Metadata.Phase != "" && !phaseMap[s.Metadata.Phase] {
					phaseMap[s.Metadata.Phase] = true
					phases = append(phases, s.Metadata.Phase)
				}
			}

			if len(phases) > 0 {
				opts := append([]string{"All Phases"}, phases...)
				idx := utils.SelectItem("Filter by Phase? (Select 'All' to skip):", opts)
				if idx > 0 {
					phase = phases[idx-1]
				}
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Archive sessions older than how many days? (default 0 for all): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if input != "" {
				if d, err := strconv.Atoi(input); err == nil {
					days = d
				} else {
					fmt.Println("Invalid number, using default (0).")
				}
			}

			fmt.Print("Delete original files after archiving? [y/N] (default: No): ")
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
				deleteOrg = true
			}
		} else {
			clientName = args[0]
			days = daysFlag
			deleteOrg = deleteFlag
			engagement = engagementFlag
			phase = phaseFlag
		}

		if !forceFlag {
			msg := fmt.Sprintf("About to archive sessions for client '%s'", clientName)
			if engagement != "" {
				msg += fmt.Sprintf(", engagement '%s'", engagement)
			}
			if phase != "" {
				msg += fmt.Sprintf(", phase '%s'", phase)
			}
			if days > 0 {
				msg += fmt.Sprintf(" older than %d days", days)
			} else {
				msg += " (ALL selected sessions)"
			}
			if deleteOrg {
				msg += ". Original files will be DELETED."
			} else {
				msg += ". Original files will be KEPT."
			}
			fmt.Println(msg)

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Continue? [y/N]: ")
			input, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(input)) != "y" {
				fmt.Println("Aborted.")
				return
			}
		}

		olderThan := time.Duration(days) * 24 * time.Hour
		count, err := logs.ArchiveSessions(clientName, engagement, phase, olderThan, deleteOrg)
		if err != nil {
			fmt.Printf("Error archiving sessions: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully archived %d sessions.\n", count)
	},
}

var archiveListCmd = &cobra.Command{
	Use:   "list",
	Short: "List existing archives",
	Run: func(cmd *cobra.Command, args []string) {
		items, err := logs.ListArchives()
		if err != nil {
			fmt.Printf("Error listing archives: %v\n", err)
			os.Exit(1)
		}

		if len(items) == 0 {
			fmt.Println("No archives found.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ARCHIVE (CLIENT/FILE)\tSIZE\tARCHIVED")
		for _, item := range items {
			size := utils.FormatBytes(item.Size)
			date := item.ModTime.Format("2006-01-02 15:04")
			fmt.Fprintf(w, "%s\t%s\t%s\n", item.DisplayPath, size, date)
		}
		w.Flush()
	},
}

func init() {
	archiveCmd.PersistentFlags().IntVarP(&daysFlag, "days", "d", 0, "Archive sessions older than N days")
	archiveCmd.PersistentFlags().BoolVarP(&deleteFlag, "delete", "D", false, "Delete original files after archiving")
	archiveCmd.PersistentFlags().BoolVarP(&forceFlag, "force", "y", false, "Skip confirmation")
	archiveCmd.PersistentFlags().StringVarP(&engagementFlag, "engagement", "e", "", "Filter by Engagement name")
	archiveCmd.PersistentFlags().StringVarP(&phaseFlag, "phase", "p", "", "Filter by Phase name")

	archiveCmd.AddCommand(archiveListCmd)
	rootCmd.AddCommand(archiveCmd)
}
