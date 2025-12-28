package cmd

import (
	"fmt"
	"os"
	"strings"
	"pentlog/pkg/config"
	"pentlog/pkg/metadata"
	"pentlog/pkg/system"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current tool and engagement status",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== Pentlog Status ===")
		
		ctx, err := metadata.Load()
		if err != nil {
			fmt.Println("Context: No active engagement found.")
		} else {
			fmt.Println("Context: ACTIVE")
			fmt.Printf("  Client:   %s\n", ctx.Client)
			fmt.Printf("  Operator: %s\n", ctx.Operator)
		}

		checkPam := func(name, path string) {
			content, err := os.ReadFile(path)
			status := "DISABLED"
			if err == nil && strings.Contains(string(content), "pam_tlog.so") {
				status = "ENABLED"
			}
			fmt.Printf("%s (%s): %s\n", name, path, status)
		}

		pamLocal, err := system.DetectLocalPamFile()
		if err != nil {
			fmt.Printf("PAM Local: ERROR (Not found)\n")
		} else {
			checkPam("PAM Local", pamLocal)
		}

		checkPam("PAM SSH", config.PamSSHD)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
