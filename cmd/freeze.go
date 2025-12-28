package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/logs"

	"github.com/spf13/cobra"
)

var freezeCmd = &cobra.Command{
	Use:   "freeze",
	Short: "Generate SHA256 hashes of all session logs",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Freezing logs (calculating SHA256)...")

		sessions, err := logs.ListSessions()
		if err != nil {
			fmt.Printf("Error listing sessions: %v\n", err)
			os.Exit(1)
		}

		hashesDir, err := config.GetHashesDir()
		if err != nil {
			fmt.Printf("Error getting hashes dir: %v\n", err)
			os.Exit(1)
		}

		if err := os.MkdirAll(hashesDir, 0700); err != nil {
			fmt.Printf("Error creating hashes dir: %v\n", err)
			os.Exit(1)
		}

		hashFilePath := filepath.Join(hashesDir, config.HashesFileName)
		f, err := os.Create(hashFilePath)
		if err != nil {
			fmt.Printf("Error creating hash file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		for _, s := range sessions {
			h := sha256.New()
			file, err := os.Open(s.Path)
			if err != nil {
				fmt.Printf("Warning: could not open %s: %v\n", s.Filename, err)
				continue
			}
			
			if _, err := io.Copy(h, file); err != nil {
				file.Close()
				fmt.Printf("Warning: could not hash %s: %v\n", s.Filename, err)
				continue
			}
			file.Close()

			sum := fmt.Sprintf("%x", h.Sum(nil))
			line := fmt.Sprintf("%s  %s\n", sum, s.Filename)
			
			if _, err := f.WriteString(line); err != nil {
				fmt.Printf("Error writing to hash file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("  %s: %s\n", s.Filename, sum)
		}

		fmt.Printf("\nHashes saved to: %s\n", hashFilePath)
	},
}

func init() {
	rootCmd.AddCommand(freezeCmd)
}
