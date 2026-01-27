package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	importPasswordFlag   string
	importClientFlag     string
	importEngagementFlag string
	importPhaseFlag      string
	importOverwriteFlag  bool
)

var importCmd = &cobra.Command{
	Use:   "import <archive.zip>",
	Short: "Import sessions from an archive",
	Long: `Import sessions from a ZIP archive.

Supports two types of archives:
1. Pentlog archives (created by 'pentlog archive') - auto-detected by folder structure
2. Generic archives containing .tty and .json files

For pentlog archives, the original folder structure is preserved.
For generic archives, you must specify the target client/engagement/phase.

Examples:
  pentlog import backup.zip                          # Import pentlog archive
  pentlog import backup.zip -P mypassword            # Import encrypted archive
  pentlog import sessions.zip -c acme -e webapp      # Import generic archive
  pentlog import external.zip -c client -p recon     # Specify target location`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		archivePath := args[0]

		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			errors.ArchiveErr(archivePath, err).Fatal()
		}

		if !strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
			errors.NewError(errors.Generic, "Only ZIP archives are supported").Fatal()
		}

		_, needsPassword, err := logs.ListArchiveContents(archivePath, "")
		if err != nil && !needsPassword {
			errors.ArchiveErr(archivePath, err).Fatal()
		}

		password := importPasswordFlag
		if needsPassword && password == "" {
			password = utils.PromptPassword("Enter archive password: ")
			if password == "" {
				errors.ArchivePasswordErr(archivePath).Fatal()
			}
		}

		if needsPassword {
			if err := logs.CheckArchivePassword(archivePath, password); err != nil {
				errors.ArchivePasswordErr(archivePath).Fatal()
			}
		}

		archiveType, err := logs.DetectArchiveType(archivePath, password)
		if err != nil && !strings.Contains(err.Error(), "password protected") {
			errors.ArchiveErr(archivePath, err).Fatal()
		}

		opts := logs.ImportOptions{
			Password:          password,
			TargetClient:      importClientFlag,
			TargetEngagement:  importEngagementFlag,
			TargetPhase:       importPhaseFlag,
			OverwriteExisting: importOverwriteFlag,
		}

		var result *logs.ImportResult

		if archiveType == logs.ArchiveTypePentlog {
			fmt.Println("📦 Detected Pentlog archive structure")

			files, _, _ := logs.ListArchiveContents(archivePath, password)
			ttyCount := 0
			for _, f := range files {
				if strings.HasSuffix(f, ".tty") {
					ttyCount++
				}
			}
			fmt.Printf("   Found %d session(s) to import\n", ttyCount)

			if !forceFlag {
				confirm := utils.PromptString("Continue with import? [Y/n]", "Y")
				if strings.ToLower(confirm) == "n" || strings.ToLower(confirm) == "no" {
					fmt.Println("Import cancelled.")
					return
				}
			}

			result, err = logs.ImportFromPentlogArchive(archivePath, opts)
		} else {
			fmt.Println("📁 Detected generic archive (no Pentlog structure)")

			if opts.TargetClient == "" {
				sessions, _ := logs.ListSessions()
				clientMap := make(map[string]bool)
				var clients []string
				for _, s := range sessions {
					if s.Metadata.Client != "" && !clientMap[s.Metadata.Client] {
						clientMap[s.Metadata.Client] = true
						clients = append(clients, s.Metadata.Client)
					}
				}

				if len(clients) > 0 {
					options := append([]string{"[Create New Client]"}, clients...)
					idx := utils.SelectItem("Select target client:", options)
					if idx == -1 {
						fmt.Println("Import cancelled.")
						return
					}
					if idx == 0 {
						opts.TargetClient = utils.PromptString("Enter new client name:", "")
					} else {
						opts.TargetClient = clients[idx-1]
					}
				} else {
					opts.TargetClient = utils.PromptString("Enter client name:", "")
				}

				if opts.TargetClient == "" {
					errors.NewError(errors.Generic, "Client name is required for generic archives").Fatal()
				}
			}

			if opts.TargetEngagement == "" {
				opts.TargetEngagement = utils.PromptString("Enter engagement name:", "imported")
			}

			if opts.TargetPhase == "" {
				opts.TargetPhase = utils.PromptString("Enter phase name:", "imported")
			}

			fmt.Printf("\n📍 Import destination:\n")
			fmt.Printf("   Client: %s\n", opts.TargetClient)
			fmt.Printf("   Engagement: %s\n", opts.TargetEngagement)
			fmt.Printf("   Phase: %s\n", opts.TargetPhase)

			// List contents for preview
			files, _, _ := logs.ListArchiveContents(archivePath, password)
			ttyCount := 0
			for _, f := range files {
				if strings.HasSuffix(f, ".tty") {
					ttyCount++
				}
			}
			fmt.Printf("   Files to import: %d session(s)\n\n", ttyCount)

			if !forceFlag {
				confirm := utils.PromptString("Continue with import? [Y/n]", "Y")
				if strings.ToLower(confirm) == "n" || strings.ToLower(confirm) == "no" {
					fmt.Println("Import cancelled.")
					return
				}
			}

			result, err = logs.ImportFromGenericArchive(archivePath, opts)
		}

		if err != nil {
			errors.FromError(errors.Generic, "Error during import", err).Fatal()
		}

		fmt.Println()
		fmt.Println("═══════════════════════════════════════")
		fmt.Println("              Import Summary           ")
		fmt.Println("═══════════════════════════════════════")
		fmt.Printf("  Total files processed: %d\n", result.TotalFiles)
		fmt.Printf("  Successfully imported: %d\n", result.ImportedFiles)
		fmt.Printf("  Skipped/Errors: %d\n", result.SkippedFiles)
		fmt.Println("═══════════════════════════════════════")

		if len(result.ImportedSessions) > 0 {
			fmt.Println("\n📋 Imported Sessions:")
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tCLIENT\tENGAGEMENT\tPHASE\tTIMESTAMP")
			for _, s := range result.ImportedSessions {
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
					s.ID,
					s.Metadata.Client,
					s.Metadata.Engagement,
					s.Metadata.Phase,
					s.ModTime,
				)
			}
			w.Flush()
		}

		if len(result.Errors) > 0 {
			fmt.Println("\n⚠️  Errors:")
			for _, e := range result.Errors {
				fmt.Printf("  - %s\n", e)
			}
		}

		if result.ImportedFiles > 0 {
			fmt.Printf("\n✅ Import complete! Run 'pentlog sessions' to view imported sessions.\n")
		}
	},
}

var importListCmd = &cobra.Command{
	Use:   "list <archive.zip>",
	Short: "Preview archive contents without importing",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		archivePath := args[0]

		if _, err := os.Stat(archivePath); os.IsNotExist(err) {
			errors.ArchiveErr(archivePath, err).Fatal()
		}

		password := importPasswordFlag
		files, needsPassword, err := logs.ListArchiveContents(archivePath, "")

		if needsPassword && password == "" {
			password = utils.PromptPassword("Enter archive password: ")
			if err := logs.CheckArchivePassword(archivePath, password); err != nil {
				errors.ArchivePasswordErr(archivePath).Fatal()
			}
			files, _, err = logs.ListArchiveContents(archivePath, password)
		}

		if err != nil {
			errors.ArchiveErr(archivePath, err).Fatal()
		}

		archiveType, _ := logs.DetectArchiveType(archivePath, password)
		if archiveType == logs.ArchiveTypePentlog {
			fmt.Println("📦 Archive Type: Pentlog structured archive")
		} else {
			fmt.Println("📁 Archive Type: Generic archive")
		}

		fmt.Printf("📄 Total files: %d\n\n", len(files))

		var sessions, metadata, notes, reports, other []string
		for _, f := range files {
			switch {
			case strings.HasSuffix(f, ".tty"):
				sessions = append(sessions, f)
			case strings.HasSuffix(f, ".notes.json"):
				notes = append(notes, f)
			case strings.HasSuffix(f, ".json"):
				metadata = append(metadata, f)
			case strings.HasPrefix(f, "reports/"):
				reports = append(reports, f)
			default:
				other = append(other, f)
			}
		}

		if len(sessions) > 0 {
			fmt.Printf("🎬 Sessions (%d):\n", len(sessions))
			for _, f := range sessions {
				fmt.Printf("   %s\n", f)
			}
		}

		if len(reports) > 0 {
			fmt.Printf("\n📊 Reports (%d):\n", len(reports))
			for _, f := range reports {
				fmt.Printf("   %s\n", f)
			}
		}

		if len(metadata) > 0 && cmd.Flags().Changed("verbose") {
			fmt.Printf("\n📋 Metadata files (%d):\n", len(metadata))
			for _, f := range metadata {
				fmt.Printf("   %s\n", f)
			}
		}

		if archiveType == logs.ArchiveTypePentlog {
			clients := make(map[string]bool)
			engagements := make(map[string]bool)
			phases := make(map[string]bool)

			for _, f := range sessions {
				parts := strings.Split(strings.TrimPrefix(f, "logs/"), string(filepath.Separator))
				if len(parts) >= 1 {
					clients[parts[0]] = true
				}
				if len(parts) >= 2 {
					engagements[parts[1]] = true
				}
				if len(parts) >= 3 {
					phases[parts[2]] = true
				}
			}

			fmt.Println("\n📊 Structure Summary:")
			fmt.Printf("   Clients: %v\n", mapKeys(clients))
			fmt.Printf("   Engagements: %v\n", mapKeys(engagements))
			fmt.Printf("   Phases: %v\n", mapKeys(phases))
		}
	},
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func init() {
	importCmd.PersistentFlags().StringVarP(&importPasswordFlag, "password", "P", "", "Password for encrypted archive")
	importCmd.PersistentFlags().StringVarP(&importClientFlag, "client", "c", "", "Target client (for generic archives)")
	importCmd.PersistentFlags().StringVarP(&importEngagementFlag, "engagement", "e", "", "Target engagement (for generic archives)")
	importCmd.PersistentFlags().StringVarP(&importPhaseFlag, "phase", "p", "", "Target phase (for generic archives)")
	importCmd.PersistentFlags().BoolVar(&importOverwriteFlag, "overwrite", false, "Overwrite existing files")
	importCmd.PersistentFlags().BoolVarP(&forceFlag, "force", "y", false, "Skip confirmation prompts")

	importListCmd.Flags().Bool("verbose", false, "Show all files including metadata")

	importCmd.AddCommand(importListCmd)
	rootCmd.AddCommand(importCmd)
}
