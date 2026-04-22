package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"pentlog/pkg/api"
	"pentlog/pkg/api/handlers"
	"pentlog/pkg/utils"

	"github.com/spf13/cobra"
)

var (
	webPort int
	webOpen bool
)

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start web dashboard server",
	Long: `Start the PentLog web dashboard server to view sessions, reports,
and manage your penetration testing evidence through a web interface.

The web dashboard provides:
  - Dashboard with statistics and activity
  - Session browser with search and filters
  - Session replay with timeline
  - Vulnerability management
  - Report generation and AI analysis
  - Archive management

Examples:
  pentlog web                    # Start on default port 8080
  pentlog web --port 3000        # Start on port 3000
  pentlog web --open             # Open in browser after starting`,
	Run: func(cmd *cobra.Command, args []string) {
		handlers.Version = Version

		if distDir, err := rebuildWebAssets(); err != nil {
			fmt.Printf("Warning: failed to rebuild web assets: %v\n", err)
		} else if distDir != "" {
			api.SetStaticDir(distDir)
		}

		server := api.NewServer(webPort)

		if webOpen {
			url := fmt.Sprintf("http://localhost:%d", webPort)
			go func() {
				time.Sleep(750 * time.Millisecond)
				if err := utils.OpenURL(url); err != nil {
					fmt.Printf("Warning: failed to open browser: %v\n", err)
				}
			}()
		}

		if err := server.Start(); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	},
}

func init() {
	webCmd.Flags().IntVarP(&webPort, "port", "p", 8080, "Port to listen on")
	webCmd.Flags().BoolVarP(&webOpen, "open", "o", false, "Open in browser after starting")
	rootCmd.AddCommand(webCmd)
}

func rebuildWebAssets() (string, error) {
	frontendDir, err := findFrontendDir()
	if err != nil {
		return "", err
	}

	fmt.Println("Rebuilding web assets...")
	buildCmd := exec.Command("npm", "run", "build")
	buildCmd.Dir = frontendDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return "", err
	}

	return filepath.Clean(filepath.Join(frontendDir, "..", "dist")), nil
}

func findFrontendDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(cwd, "pkg", "web", "frontend", "package.json")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Dir(candidate), nil
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", fmt.Errorf("frontend source not found from %s", cwd)
		}
		cwd = parent
	}
}
