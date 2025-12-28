package cmd

import (
	"fmt"
	"os"
	"runtime"
	"pentlog/pkg/config"
	"pentlog/pkg/system"

	"github.com/spf13/cobra"
)

var (
	enableLocal bool
	enableSSH   bool
)

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable recording for Local or SSH sessions",
	Run: func(cmd *cobra.Command, args []string) {
		if !enableLocal && !enableSSH {
			cmd.Help()
			return
		}

		if runtime.GOOS == "linux" && !system.IsRoot() {
			fmt.Println("Error: enable must be run as root.")
			os.Exit(1)
		}

		if enableLocal {
			fmt.Print("Enabling PAM for local sessions... ")
			pamFile, err := system.DetectLocalPamFile()
			if err != nil {
				fmt.Printf("FAIL: %v\n", err)
			} else {
				fmt.Printf("(%s) ", pamFile)
				if err := system.EnablePamTlog(pamFile); err != nil {
					fmt.Printf("FAIL: %v\n", err)
				} else {
					fmt.Println("OK")
				}
			}
		}

		if enableSSH {
			fmt.Print("Enabling PAM for SSH sessions (sshd)... ")
			if err := system.EnablePamTlog(config.PamSSHD); err != nil {
				fmt.Printf("FAIL: %v\n", err)
			} else {
				fmt.Println("OK")
			}
		}
	},
}

func init() {
	enableCmd.Flags().BoolVar(&enableLocal, "local", false, "Enable PAM tlog for local sessions")
	enableCmd.Flags().BoolVar(&enableSSH, "ssh", false, "Enable PAM tlog for SSH sessions")
	rootCmd.AddCommand(enableCmd)
}