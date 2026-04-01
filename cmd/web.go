package cmd

import (
	"fmt"
	"pentlog/pkg/api"
	"pentlog/pkg/api/handlers"
	"pentlog/pkg/web"

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

		api.SetStaticFS(web.StaticFS)

		server := api.NewServer(webPort)

		if webOpen {
			url := fmt.Sprintf("http://localhost:%d", webPort)
			fmt.Printf("Opening %s in browser...\n", url)
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
