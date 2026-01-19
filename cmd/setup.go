package cmd

import (
	"bufio"
	"fmt"
	"os"
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

		fmt.Println("Setup complete. Run 'pentlog create' to initialize a new session.")
	},
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
