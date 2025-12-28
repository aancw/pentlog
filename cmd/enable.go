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
				changed, err := system.EnablePamTlog(pamFile)
				if err != nil {
					fmt.Printf("FAIL: %v\n", err)
				} else {
					if changed {
						fmt.Println("OK")
					} else {
						fmt.Println("Already enabled")
					}
				}
			}
		}

		if enableSSH {
			fmt.Println("WARNING: Enabling PAM for SSH will restart the SSH service.")
			fmt.Print("Enabling PAM for SSH sessions (sshd)... ")
			changed, err := system.EnablePamTlog(config.PamSSHD)
			if err != nil {
				fmt.Printf("FAIL: %v\n", err)
			} else {
				if changed {
					fmt.Print("OK. Restarting SSH... ")
					if err := system.RestartSSHService(); err != nil {
						fmt.Printf("FAIL (%v)\n", err)
						fmt.Println("Please restart ssh service manually.")
					} else {
						fmt.Println("DONE")
					}
				} else {
					fmt.Println("Already enabled")
				}
			}
		}
	},
}

func init() {
	enableCmd.Flags().BoolVar(&enableLocal, "local", false, "Enable PAM tlog for local sessions")
	enableCmd.Flags().BoolVar(&enableSSH, "ssh", false, "Enable PAM tlog for SSH sessions")
	rootCmd.AddCommand(enableCmd)
}