package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/deps"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	gifOutputFlag string
	gifSpeedFlag  float64
)

var gifCmd = &cobra.Command{
	Use:   "gif [id]",
	Short: "Convert session(s) to GIF using ttygif",
	Long: `Convert one or more sessions to an animated GIF.

Note: ttygif requires a graphical terminal environment (X11/Wayland).

Examples:
  pentlog gif 5              # Convert session ID 5 to GIF
  pentlog gif -s 5           # Convert at 5x speed (faster, lower CPU)
  pentlog gif                # Interactive mode: select session or merge
  pentlog gif -o demo.gif    # Specify output filename`,
	Run: func(cmd *cobra.Command, args []string) {
		dm := deps.NewManager()
		if ok, _ := dm.Check("ttygif"); !ok {
			install := utils.SelectItem("ttygif is required but not installed. Install it?", []string{"Yes", "No"})
			if install == 0 {
				if err := dm.Install("ttygif"); err != nil {
					fmt.Printf("Error installing ttygif: %v\n", err)
					return
				}
			} else {
				fmt.Println("GIF conversion requires ttygif. Aborting.")
				return
			}
		}

		if runtime.GOOS == "linux" {
			windowID := os.Getenv("WINDOWID")
			// Validate WINDOWID - either missing or set to 0 (both invalid)
			if windowID == "" || windowID == "0" {
				if ok, _ := dm.Check("xdotool"); !ok {
					fmt.Println("Error: xdotool is required to capture terminal window.")
					fmt.Println("Install it with: sudo apt-get install xdotool")
					return
				}
				out, err := exec.Command("xdotool", "getwindowfocus").Output()
				if err != nil {
					fmt.Println("Error: Could not get WINDOWID. Make sure you're in a graphical terminal.")
					fmt.Printf("xdotool error: %v\n", err)
					return
				}
				windowID = strings.TrimSpace(string(out))
				// Validate that WINDOWID is not 0 (which indicates no valid window)
				if windowID == "0" {
					fmt.Println("Error: Could not detect a valid terminal window.")
					fmt.Println("This command requires a graphical terminal environment (X11/Wayland).")
					fmt.Println("ttygif cannot run in headless/SSH-only environments.")
					return
				}
				os.Setenv("WINDOWID", windowID)
				fmt.Printf("Auto-detected WINDOWID: %s\n", windowID)
			}
		}

		if runtime.GOOS == "darwin" {
			fmt.Println("Note: On macOS, ttygif requires screen recording permissions.")
			fmt.Println("Go to: System Settings > Privacy & Security > Screen Recording")
			fmt.Println("Add your terminal app (Terminal.app, iTerm2, etc.)")
			fmt.Println("")
		}

		if len(args) > 0 {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("Invalid session ID: %s\n", args[0])
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

	reportsDir, err := config.GetReportsDir()
	if err != nil {
		fmt.Printf("Error getting reports directory: %v\n", err)
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

	// Save to ~/.pentlog/reports/ like HTML/Markdown exports
	outputPath := filepath.Join(reportsDir, outputName)
	runTtygif(session.Path, outputPath, id)
}

func runTtygif(inputPath, outputPath string, sessionID int) {
	// Create temp directory in ~/.pentlog/reports/tmp for ttygif intermediate files
	reportsDir, err := config.GetReportsDir()
	if err != nil {
		fmt.Printf("Error getting reports directory: %v\n", err)
		return
	}
	tmpDir := filepath.Join(reportsDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		fmt.Printf("Error creating temp directory: %v\n", err)
		return
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Change to temp directory for ttygif to output intermediate PNG files
	oldCwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}
	if err := os.Chdir(tmpDir); err != nil {
		fmt.Printf("Error changing to temp directory: %v\n", err)
		return
	}
	defer os.Chdir(oldCwd)

	speedStr := fmt.Sprintf("%.1f", gifSpeedFlag)
	if gifSpeedFlag > 1 {
		fmt.Printf("Converting at %.1fx speed (faster processing)...\n", gifSpeedFlag)
	}

	if sessionID > 0 {
		fmt.Printf("Converting session %d to GIF...\n", sessionID)
	} else {
		fmt.Printf("Converting to GIF...\n")
	}

	// Set TMPDIR environment variable to use ~/.pentlog/exports/tmp for ImageMagick
	cmd := exec.Command("ttygif", inputPath, "-s", speedStr)
	cmd.Env = append(os.Environ(), "TMPDIR="+tmpDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running ttygif: %v\n", err)
		fmt.Println("")
		fmt.Println("=== Troubleshooting Tips ===")
		fmt.Println("1. For medium/large sessions, try increasing speed to reduce file size:")
		fmt.Println("   ./pentlog gif -s 10  (10x speed = smaller GIF, faster conversion)")
		fmt.Println("")
		fmt.Println("2. Check ImageMagick policy restrictions:")
		fmt.Println("   cat /etc/ImageMagick-6/policy.xml | grep -A5 'gif'")
		fmt.Println("")
		fmt.Println("3. If policy restricts GIF, edit the policy file and remove GIF restrictions:")
		fmt.Println("   sudo nano /etc/ImageMagick-6/policy.xml")
		fmt.Println("")
		fmt.Println("4. Alternatively, increase memory limit for convert:")
		fmt.Println("   export MAGICK_MEMORY_LIMIT=2GB")
		fmt.Println("   ./pentlog gif -s 5")
		fmt.Println("")
		return
	}

	// Move GIF from temp dir to final output location
	gifTempPath := filepath.Join(tmpDir, "tty.gif")
	if err := os.Rename(gifTempPath, outputPath); err != nil {
		fmt.Printf("Conversion ran, but failed to move GIF to '%s': %v\n", outputPath, err)
		fmt.Printf("Check if 'tty.gif' was created in %s\n", tmpDir)
	} else {
		fmt.Printf("Success! GIF saved to %s\n", outputPath)
	}
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

	reportsDir, err := config.GetReportsDir()
	if err != nil {
		fmt.Printf("Error getting reports directory: %v\n", err)
		return
	}
	tempFile := filepath.Join(reportsDir, "tmp", "pentlog_merged.tty")
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

	// Save to ~/.pentlog/reports/ like HTML/Markdown exports
	outputPath := filepath.Join(reportsDir, outputName)
	runTtygif(tempFile, outputPath, 0)
}

func init() {
	gifCmd.Flags().StringVarP(&gifOutputFlag, "output", "o", "", "Output GIF filename")
	gifCmd.Flags().Float64VarP(&gifSpeedFlag, "speed", "s", 1.0, "Playback speed (higher = faster conversion, e.g., 5.0 for 5x)")
	rootCmd.AddCommand(gifCmd)
}
