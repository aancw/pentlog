package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var outputFile string

var timelineCmd = &cobra.Command{
	Use:   "timeline [id]",
	Short: "Extract command timeline from a session",
	Long: `Analyze a terminal session recording and extract a chronological timeline
of commands executed and their outputs.

If no session ID is provided, an interactive session selector will be displayed.`,
	Run: func(cmd *cobra.Command, args []string) {
		var id int
		var err error

		if len(args) > 0 {
			id, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("Invalid session ID: %s\n", args[0])
				os.Exit(1)
			}
		} else {
			sessions, err := logs.ListSessions()
			if err != nil {
				fmt.Printf("Error listing sessions: %v\n", err)
				os.Exit(1)
			}
			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			startIdx := 0
			if len(sessions) > 15 {
				startIdx = len(sessions) - 15
			}
			displaySessions := sessions[startIdx:]

			var items []string
			for _, s := range displaySessions {
				items = append(items, fmt.Sprintf("ID %d | %s | %s", s.ID, s.ModTime, s.DisplayPath))
			}

			fmt.Println("Recent Sessions:")
			idx := utils.SelectItem("Select Session to Analyze:", items)
			if idx == -1 {
				fmt.Println("Selection canceled.")
				return
			}
			id = displaySessions[idx].ID
		}

		session, err := logs.GetSession(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if session.Path == "" {
			fmt.Println("Error: session file missing; cannot analyze.")
			os.Exit(1)
		}

		spin := utils.NewSpinner(fmt.Sprintf("Analyzing session %d...", id))
		spin.Start()

		timeline, err := logs.ParseTimeline(session.Path)
		spin.Stop()

		if err != nil {
			fmt.Printf("Error parsing timeline: %v\n", err)
			os.Exit(1)
		}

		if len(timeline.Commands) == 0 {
			fmt.Println("No commands found in this session")
			return
		}

		// If output file specified, export as JSON
		if outputFile != "" {
			jsonOutput, err := timeline.ToJSON()
			if err != nil {
				fmt.Printf("Error generating JSON: %v\n", err)
				os.Exit(1)
			}

			err = os.WriteFile(outputFile, []byte(jsonOutput), 0644)
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Timeline saved to %s\n", outputFile)
			return
		}

		// Interactive mode: keep looping until user exits with Ctrl+C
		for {
			options := []string{
				"View timeline (human-readable)",
				"Export as JSON",
			}

			choice := utils.SelectItem("What would you like to do?", options)
			if choice < 0 {
				return
			}

			if choice == 0 {
				displayTimeline(timeline)
			} else {
				fmt.Print("Enter output filename: ")
				var filename string
				fmt.Scanln(&filename)

				if filename == "" {
					filename = "timeline.json"
				}

				jsonOutput, err := timeline.ToJSON()
				if err != nil {
					fmt.Printf("Error generating JSON: %v\n", err)
					continue
				}

				err = os.WriteFile(filename, []byte(jsonOutput), 0644)
				if err != nil {
					fmt.Printf("Error writing file: %v\n", err)
					continue
				}

				fmt.Printf("\n✅ Timeline exported to %s\n\n", filename)
			}
		}
	},
}

func init() {
	timelineCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path for JSON export")
	rootCmd.AddCommand(timelineCmd)
}

func displayTimeline(timeline *logs.Timeline) {
	fmt.Printf("\n%s\n", strings.Repeat("=", 80))
	fmt.Printf("COMMAND TIMELINE (%d commands)\n", len(timeline.Commands))
	fmt.Printf("%s\n\n", strings.Repeat("=", 80))

	for i, cmd := range timeline.Commands {
		fmt.Printf("[%d] %s\n", i+1, cmd.Timestamp)
		fmt.Printf("Command: %s\n", cmd.Command)

		if cmd.Output != "" {
			fmt.Println("\nOutput:")
			fmt.Printf("%s\n", strings.Repeat("-", 80))
			fmt.Println(cmd.Output)
			fmt.Printf("%s\n", strings.Repeat("-", 80))
		} else {
			fmt.Println("Output: (none)")
		}

		fmt.Println()
	}

	fmt.Printf("%s\n", strings.Repeat("=", 80))
	fmt.Println("\nPress Enter to continue...")
	fmt.Scanln()
}
