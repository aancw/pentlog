package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v80/github"
	"github.com/spf13/cobra"
)

const (
	RepoOwner = "aancw"
	RepoName  = "pentlog"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update pentlog to the latest version",
	Long:  `Downloads and installs the latest version of pentlog from GitHub Releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var client *github.Client

		token := os.Getenv("GH_TOKEN")
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}

		if token != "" {
			fmt.Println("Using provided GitHub token for authentication...")
			client = github.NewClient(nil).WithAuthToken(token)
		} else {
			fmt.Println("No GitHub token found. Attempting unauthenticated access (public repo)...")
			client = github.NewClient(nil)
		}

		fmt.Println("Checking for updates in upstream server...")
		release, _, err := client.Repositories.GetLatestRelease(ctx, RepoOwner, RepoName)
		if err != nil {
			fmt.Printf("Error fetching latest release: %v\n", err)
			return
		}

		tagName := release.GetTagName()
		fmt.Printf("Latest version: %s\n", tagName)
		fmt.Printf("Current version: %s\n", Version)

		if tagName == Version {
			fmt.Println("You are already using the latest version.")
			return
		}

		assetNamePrefix := fmt.Sprintf("pentlog-%s-%s", runtime.GOOS, runtime.GOARCH)
		var targetAsset *github.ReleaseAsset

		for _, asset := range release.Assets {
			if strings.HasPrefix(asset.GetName(), assetNamePrefix) {
				targetAsset = asset
				break
			}
		}

		if targetAsset == nil {
			fmt.Printf("No compatible asset found for %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return
		}

		fmt.Printf("OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("Downloading update...")

		var rc io.ReadCloser
		if token != "" {
			rc, _, err = client.Repositories.DownloadReleaseAsset(ctx, RepoOwner, RepoName, targetAsset.GetID(), http.DefaultClient)

		} else {
			resp, err := http.Get(targetAsset.GetBrowserDownloadURL())
			if err != nil {
				fmt.Printf("Error downloading asset: %v\n", err)
				return
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error downloading asset: status %s\n", resp.Status)
				return
			}
			rc = resp.Body
		}

		if err != nil {
			fmt.Printf("Error downloading asset: %v\n", err)
			return
		}

		if rc != nil {
			defer rc.Close()
		}

		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error finding executable path: %v\n", err)
			return
		}
		exeDir := filepath.Dir(exePath)

		tmpFile, err := os.CreateTemp(exeDir, "pentlog-update-*")
		if err != nil {
			fmt.Printf("Error creating temp file in %s: %v\n", exeDir, err)
			return
		}
		defer os.Remove(tmpFile.Name())

		if rc != nil {
			_, err = io.Copy(tmpFile, rc)
			if err != nil {
				fmt.Printf("Error writing to temp file: %v\n", err)
				tmpFile.Close()
				return
			}
		}
		tmpFile.Close()

		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			fmt.Printf("Error making binary executable: %v\n", err)
			return
		}

		if err := os.Rename(tmpFile.Name(), exePath); err != nil {
			fmt.Printf("Error replacing binary: %v\n", err)
			return
		}

		greenArrow := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("->")
		fmt.Printf("Successfully updated: %s %s %s\n", Version, greenArrow, tagName)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
