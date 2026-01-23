package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/logs"
	"pentlog/pkg/recorder"
	"pentlog/pkg/utils"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	gifOutputFlag string
	gifSpeedFlag  float64
	gifColsFlag   int
	gifRowsFlag   int
)

var gifCmd = &cobra.Command{
	Use:   "gif [id|file.tty]",
	Short: "Convert session(s) to animated GIF",
	Long: `Convert one or more sessions to an animated GIF using native Go rendering.

Examples:
  pentlog gif 5                    # Convert session ID 5 to GIF
  pentlog gif session.tty          # Convert a .tty file directly
  pentlog gif -s 5                 # Convert at 5x speed (faster playback)
  pentlog gif                      # Interactive mode: select session or merge
  pentlog gif -o demo.gif          # Specify output filename
  pentlog gif --cols 120 --rows 40 # Custom terminal size`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			arg := args[0]
			if _, err := os.Stat(arg); err == nil && filepath.Ext(arg) == ".tty" {
				convertTTYFile(arg)
				return
			}
			id, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Printf("Invalid session ID or file not found: %s\n", arg)
				os.Exit(1)
			}
			convertSingleSession(id)
			return
		}

		sessions, err := logs.ListSessions()
		if err != nil {
			fmt.Printf("Error listing sessions: %v\n", err)
			os.Exit(1)
		}
		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
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

		idx := utils.SelectItem("Select Client:", clients)
		if idx == -1 {
			return
		}
		selectedClient := clients[idx]

		engMap := make(map[string]bool)
		var engagements []string
		for _, s := range sessions {
			if s.Metadata.Client == selectedClient && s.Metadata.Engagement != "" && !engMap[s.Metadata.Engagement] {
				engMap[s.Metadata.Engagement] = true
				engagements = append(engagements, s.Metadata.Engagement)
			}
		}

		var selectedEngagement string
		if len(engagements) > 0 {
			opts := append([]string{"All Engagements"}, engagements...)
			idx := utils.SelectItem("Select Engagement:", opts)
			if idx == -1 {
				return
			}
			if idx > 0 {
				selectedEngagement = engagements[idx-1]
			}
		}

		modeIdx := utils.SelectItem("Select Mode:", []string{"Select Single Session", "Merge All Sessions in Scope"})
		if modeIdx == -1 {
			return
		}

		var filteredSessions []logs.Session
		for _, s := range sessions {
			if s.Metadata.Client != selectedClient {
				continue
			}
			if selectedEngagement != "" && s.Metadata.Engagement != selectedEngagement {
				continue
			}
			filteredSessions = append(filteredSessions, s)
		}

		if len(filteredSessions) == 0 {
			fmt.Println("No sessions found matching the criteria.")
			return
		}

		if modeIdx == 0 {
			var items []string
			for _, s := range filteredSessions {
				items = append(items, fmt.Sprintf("ID %d | %s | %s", s.ID, s.ModTime, s.DisplayPath))
			}

			idx := utils.SelectItem("Select Session to Convert:", items)
			if idx == -1 {
				return
			}
			convertSingleSession(filteredSessions[idx].ID)
		} else {
			convertMergedSessions(filteredSessions, selectedClient, selectedEngagement)
		}
	},
}

func convertTTYFile(inputPath string) {
	// Prompt for resolution
	resOptions := []string{"720p (1280x720)", "1080p (1920x1080)"}
	resIdx := utils.SelectItem("Select Resolution:", resOptions)
	if resIdx == -1 {
		return
	}
	resolution := "720p"
	if resIdx == 1 {
		resolution = "1080p"
	}

	outputName := gifOutputFlag
	if outputName == "" {
		base := filepath.Base(inputPath)
		defaultName := base[:len(base)-len(filepath.Ext(base))] + ".gif"
		outputName = utils.PromptString("Enter output filename", defaultName)
		if outputName == "" {
			outputName = defaultName
		}
	}

	if filepath.Ext(outputName) != ".gif" {
		outputName += ".gif"
	}

	outputPath := outputName
	if !filepath.IsAbs(outputPath) {
		mgr := config.Manager()
		reportsDir := mgr.GetPaths().ReportsDir
		os.MkdirAll(reportsDir, 0700)
		outputPath = filepath.Join(reportsDir, outputName)
	}

	renderGIF(inputPath, outputPath, 0, resolution)
}

func convertSingleSession(id int) {
	session, err := logs.GetSession(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if session.Path == "" {
		fmt.Println("Error: session file missing; cannot convert.")
		os.Exit(1)
	}

	// Prompt for resolution
	resOptions := []string{"720p (1280x720)", "1080p (1920x1080)"}
	resIdx := utils.SelectItem("Select Resolution:", resOptions)
	if resIdx == -1 {
		return
	}
	resolution := "720p"
	if resIdx == 1 {
		resolution = "1080p"
	}

	mgr := config.Manager()
	reportsDir := mgr.GetPaths().ReportsDir

	if err := os.MkdirAll(reportsDir, 0700); err != nil {
		fmt.Printf("Error creating reports directory: %v\n", err)
		os.Exit(1)
	}

	outputName := gifOutputFlag
	if outputName == "" {
		defaultName := fmt.Sprintf("session_%d.gif", id)
		outputName = utils.PromptString("Enter output filename", defaultName)
		if outputName == "" {
			outputName = defaultName
		}
	}

	if filepath.Ext(outputName) != ".gif" {
		outputName += ".gif"
	}

	outputPath := filepath.Join(reportsDir, outputName)
	renderGIF(session.Path, outputPath, id, resolution)
}

func renderGIF(inputPath, outputPath string, sessionID int, resolution string) {
	if sessionID > 0 {
		fmt.Printf("Converting session %d to GIF...\n", sessionID)
	} else {
		fmt.Printf("Converting to GIF...\n")
	}

	cfg := recorder.DefaultConfig()
	cfg.Speed = gifSpeedFlag
	cfg.Resolution = resolution

	// Set dimensions based on resolution
	if resolution == "1080p" {
		cfg.Cols = 274
		cfg.Rows = 83
	} else { // 720p
		cfg.Cols = 183
		cfg.Rows = 55
	}

	// Allow manual override via flags
	if gifColsFlag > 0 {
		cfg.Cols = gifColsFlag
	}
	if gifRowsFlag > 0 {
		cfg.Rows = gifRowsFlag
	}

	if err := recorder.RenderToGIF(inputPath, outputPath, cfg); err != nil {
		fmt.Printf("Error rendering GIF: %v\n", err)
		return
	}

	fmt.Printf("Success! GIF saved to %s\n", outputPath)
}

func convertMergedSessions(sessions []logs.Session, client, engagement string) {
	var ttyFiles []string
	for _, s := range sessions {
		if s.Path != "" {
			ttyFiles = append(ttyFiles, s.Path)
		}
	}

	if len(ttyFiles) == 0 {
		fmt.Println("No valid session files found to merge.")
		return
	}

	fmt.Printf("Merging %d sessions...\n", len(ttyFiles))

	// Prompt for resolution
	resOptions := []string{"720p (1280x720)", "1080p (1920x1080)"}
	resIdx := utils.SelectItem("Select Resolution:", resOptions)
	if resIdx == -1 {
		return
	}
	resolution := "720p"
	if resIdx == 1 {
		resolution = "1080p"
	}

	mgr := config.Manager()
	reportsDir := mgr.GetPaths().ReportsDir

	if err := os.MkdirAll(reportsDir, 0700); err != nil {
		fmt.Printf("Error creating reports directory: %v\n", err)
		return
	}

	tempFile := filepath.Join(os.TempDir(), "pentlog_merged.tty")
	if err := logs.MergeTTYFiles(ttyFiles, tempFile); err != nil {
		fmt.Printf("Error merging TTY files: %v\n", err)
		return
	}
	defer os.Remove(tempFile)

	outputName := gifOutputFlag
	if outputName == "" {
		defaultName := utils.Slugify(client)
		if engagement != "" {
			defaultName += "_" + utils.Slugify(engagement)
		}
		defaultName += "_merged.gif"
		outputName = utils.PromptString("Enter output filename", defaultName)
		if outputName == "" {
			outputName = defaultName
		}
	}

	if filepath.Ext(outputName) != ".gif" {
		outputName += ".gif"
	}

	outputPath := filepath.Join(reportsDir, outputName)
	renderGIF(tempFile, outputPath, 0, resolution)
}

func init() {
	gifCmd.Flags().StringVarP(&gifOutputFlag, "output", "o", "", "Output GIF filename")
	gifCmd.Flags().Float64VarP(&gifSpeedFlag, "speed", "s", 1.0, "Playback speed (higher = faster GIF playback)")
	gifCmd.Flags().IntVar(&gifColsFlag, "cols", 160, "Terminal width in columns")
	gifCmd.Flags().IntVar(&gifRowsFlag, "rows", 45, "Terminal height in rows")
	rootCmd.AddCommand(gifCmd)
}
