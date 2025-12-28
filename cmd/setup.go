package cmd

import (
	"fmt"
	"os"
	"runtime"
	"pentlog/pkg/system"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install dependencies and configure PAM",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting pentlog setup...")

		if runtime.GOOS != "linux" {
			fmt.Println("Warning: This tool is designed for Linux systems.")
			fmt.Println("Continuing in simulation/unsafe mode for development...")
		} else {
			if !system.IsRoot() {
				fmt.Println("Error: setup must be run as root.")
				os.Exit(1)
			}
		}

		fmt.Print("Checking dependencies... ")
		if err := system.CheckDependencies(); err != nil {
			fmt.Printf("FAIL\n%v\n", err)
			os.Exit(1)
		}
		fmt.Println("OK")

		fmt.Print("Configuring tlog... ")
		if err := system.EnsureTlogConfig(); err != nil {
			fmt.Printf("FAIL\n%v\n", err)
			os.Exit(1)
		}
		fmt.Println("OK")

		fmt.Print("Detecting local PAM configuration... ")
		pamFile, err := system.DetectLocalPamFile()
		if err != nil {
			fmt.Printf("FAIL\n%v\n", err)
			if runtime.GOOS == "linux" {
				os.Exit(1)
			}
			pamFile = "/etc/pam.d/common-session"
		} else {
			fmt.Printf("found %s\n", pamFile)
		}

		fmt.Printf("Enabling PENTLOG MANAGED BLOCK for local sessions (%s)... ", pamFile)
		changed, err := system.EnablePamTlog(pamFile)
		if err != nil {
			fmt.Printf("FAIL\n%v\n", err)
            if runtime.GOOS == "linux" {
			    os.Exit(1)
            }
		} else {
			if changed {
				fmt.Println("OK")
			} else {
				fmt.Println("Already enabled")
			}
        }

		fmt.Println("Setup complete. tlog is ready.")
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
