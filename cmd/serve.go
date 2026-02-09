package cmd

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/errors"
	"pentlog/pkg/utils"
	"strings"

	"github.com/spf13/cobra"
)

var servePort int
var serveOpen bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a local HTTP server to view generated reports",
	Long: `Start a local HTTP server to serve the reports directory.
This allows viewing HTML reports with embedded GIF players without CORS issues.

Examples:
  pentlog serve              # Interactive report selection
  pentlog serve --port 8080  # Use specific port
  pentlog serve --open       # Auto-open in browser`,
	Run: func(cmd *cobra.Command, args []string) {
		mgr := config.Manager()
		reportsDir := mgr.GetPaths().ReportsDir

		if _, err := os.Stat(reportsDir); os.IsNotExist(err) {
			errors.DirErr(reportsDir, fmt.Errorf("reports directory not found")).Print()
			fmt.Println("Run 'pentlog export' to generate reports first.")
			return
		}

		clients, err := os.ReadDir(reportsDir)
		if err != nil {
			errors.DirErr(reportsDir, err).Print()
			return
		}

		var reports []string
		var reportPaths []string

		for _, client := range clients {
			if !client.IsDir() {
				continue
			}
			clientDir := filepath.Join(reportsDir, client.Name())
			files, err := os.ReadDir(clientDir)
			if err != nil {
				continue
			}
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".html") {
					displayName := fmt.Sprintf("%s / %s", client.Name(), f.Name())
					reports = append(reports, displayName)
					reportPaths = append(reportPaths, filepath.Join(client.Name(), f.Name()))
				}
			}
		}

		if len(reports) == 0 {
			fmt.Println("No HTML reports found.")
			fmt.Println("Run 'pentlog export' and save as HTML to generate reports.")
			return
		}

		reports = append([]string{"[Browse All Reports]"}, reports...)
		reportPaths = append([]string{""}, reportPaths...)

		idx := utils.SelectItem("Select Report to View", reports)
		if idx == -1 {
			return
		}

		selectedPath := reportPaths[idx]

		port := servePort
		if port == 0 {
			listener, err := net.Listen("tcp", ":0")
			if err != nil {
				errors.FromError(errors.Generic, "Failed to find available port", err).Print()
				return
			}
			port = listener.Addr().(*net.TCPAddr).Port
			listener.Close()
		}

		fs := http.FileServer(http.Dir(reportsDir))
		http.Handle("/", fs)

		var url string
		if selectedPath == "" {
			url = fmt.Sprintf("http://localhost:%d/", port)
		} else {
			url = fmt.Sprintf("http://localhost:%d/%s", port, selectedPath)
		}

		fmt.Printf("\n📁 Serving reports from: %s\n", reportsDir)
		fmt.Printf("🌐 Server running at: http://localhost:%d/\n", port)
		fmt.Printf("📄 Opening: %s\n", url)
		fmt.Println("\nPress Ctrl+C to stop the server.")

		go func() {
			if err := utils.OpenURL(url); err != nil {
				fmt.Printf("Could not open browser automatically. Please visit: %s\n", url)
			}
		}()

		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
			errors.FromError(errors.Generic, "Server error", err).Print()
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 0, "Port to serve on (default: random available port)")
	serveCmd.Flags().BoolVarP(&serveOpen, "open", "o", true, "Open browser automatically")
}
