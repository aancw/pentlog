package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/deps"
	"pentlog/pkg/system"
	"strings"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Verify dependencies and prepare local logging",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(Banner)
		fmt.Println("Starting pentlog setup...")

		fmt.Print("Checking dependencies...\n")
		depManager := deps.NewManager()
		allGood := true
		for _, dep := range depManager.Dependencies {
			ok, path := depManager.Check(dep.Name)
			if ok {
				fmt.Printf("  - %s: OK (%s)\n", dep.Name, path)
			} else {
				fmt.Printf("  - %s: MISSING\n", dep.Name)
				allGood = false

				if promptYesNo(fmt.Sprintf("Do you want to try installing %s automatically?", dep.Name)) {
					if err := depManager.Install(dep.Name); err != nil {
						fmt.Printf("    Installation failed: %v\n", err)
						fmt.Println("    Please install it manually.")
					} else {
						fmt.Println("    Installation successful.")
						if ok, path := depManager.Check(dep.Name); ok {
							fmt.Printf("  - %s: OK (%s)\n", dep.Name, path)
						}
					}
				} else {
					fmt.Println("    Skipping installation. Some features may not work.")
				}
			}
		}

		if !allGood {
			fmt.Println("\nWarning: Some dependencies are missing. 'setup' will continue, but ensure you install them for full functionality.")
		}

		fmt.Print("Preparing log directory... ")
		logDir, err := system.EnsureLogDir()
		if err != nil {
			fmt.Printf("FAIL\n%v\n", err)
			os.Exit(1)
		}
		fmt.Printf("OK (%s)\n", logDir)

		fmt.Print("Downloading latest templates... ")
		mgr := config.Manager()
		templatesDir := mgr.GetPaths().TemplatesDir
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			fmt.Printf("FAIL (mkdir)\n%v\n", err)
		} else {
			baseURL := "https://raw.githubusercontent.com/aancw/pentlog/main/assets/templates/"
			files := []string{"report.html", "report.css"}
			success := true

			for _, file := range files {
				destPath := filepath.Join(templatesDir, file)
				url := baseURL + file
				if err := downloadFile(url, destPath); err != nil {
					fmt.Printf("\n    FAIL (download %s): %v", file, err)
					success = false
				}
			}
			if success {
				fmt.Printf("OK (%s)\n", templatesDir)
			} else {
				fmt.Println("\n    (Some templates failed to download. Check your internet connection.)")
			}
		}

		fmt.Println("Setup complete. Run 'pentlog create' to initialize a new session.")
	},
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func promptYesNo(question string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", question)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
