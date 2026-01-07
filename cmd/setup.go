package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/system"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Verify dependencies and prepare local logging",
	Run: func(cmd *cobra.Command, args []string) {
		const banner = `
 ███████████                       █████    █████                        
░░███░░░░░███                     ░░███    ░░███                         
 ░███    ░███  ██████  ████████   ███████   ░███         ██████   ███████
 ░██████████  ███░░███░░███░░███ ░░░███░    ░███        ███░░███ ███░░███
 ░███░░░░░░  ░███████  ░███ ░███   ░███     ░███       ░███ ░███░███ ░███
 ░███        ░███░░░   ░███ ░███   ░███ ███ ░███      █░███ ░███░███ ░███
 █████       ░░██████  ████ █████  ░░█████  ███████████░░██████ ░░███████
░░░░░         ░░░░░░  ░░░░ ░░░░░    ░░░░░  ░░░░░░░░░░░  ░░░░░░   ░░░░░███
                                                                 ███ ░███
                                                                ░░██████ 
                                                                 ░░░░░░  
                 PentLog — Evidence-First Pentest Logging Tool
                                        created by Petruknisme
`
		fmt.Print(banner)
		fmt.Println("Starting pentlog setup...")

		fmt.Print("Checking dependencies... ")
		if err := system.CheckDependencies(); err != nil {
			fmt.Printf("FAIL\n%v\n", err)
			os.Exit(1)
		}
		fmt.Println("OK")

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
