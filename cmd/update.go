package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

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

		fmt.Printf("Checking for updates in %s/%s...\n", RepoOwner, RepoName)
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

		fmt.Printf("Downloading %s...\n", targetAsset.GetName())

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
		defer rc.Close()

		tmpFile, err := os.CreateTemp("", "pentlog-update-*")
		if err != nil {
			fmt.Printf("Error creating temp file: %v\n", err)
			return
		}
		defer os.Remove(tmpFile.Name())

		_, err = io.Copy(tmpFile, rc)
		if err != nil {
			fmt.Printf("Error writing to temp file: %v\n", err)
			tmpFile.Close()
			return
		}
		tmpFile.Close()

		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			fmt.Printf("Error making binary executable: %v\n", err)
			return
		}

		exePath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error finding executable path: %v\n", err)
			return
		}

		if err := os.Rename(tmpFile.Name(), exePath); err != nil {
			fmt.Printf("Error replacing binary: %v\n", err)
			return
		}

		fmt.Printf("Successfully updated to %s\n", tagName)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
