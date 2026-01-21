package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/utils"
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

var updateTemplateFlag bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update pentlog to the latest version",
	Long:  `Downloads and installs the latest version of pentlog from GitHub Releases.`,
	Run: func(cmd *cobra.Command, args []string) {
		if updateTemplateFlag {
			updateTemplates()
			return
		}

		ctx := context.Background()
		var client *github.Client

		client = github.NewClient(nil)

		fmt.Println("Checking for updates in upstream server...")
		release, _, err := client.Repositories.GetLatestRelease(ctx, RepoOwner, RepoName)
		if err != nil {
			fmt.Printf("Error fetching latest release: %v\n", err)
			return
		}

		tagName := release.GetTagName()
		fmt.Printf("Latest version: %s\n", tagName)
		fmt.Printf("Current version: %s\n", Version)

		prompt := utils.PromptString("Do you want to read changelog? (y/N)", "no")
		if strings.ToLower(prompt) == "y" || strings.ToLower(prompt) == "yes" {
			fmt.Println("\n--- Changelog ---")
			fmt.Println(release.GetBody())
			fmt.Println("-----------------")
		}

		if tagName == Version {
			fmt.Println("You are already using the latest version.")
			return
		}

		targetAsset := findCompatibleAsset(release.Assets, runtime.GOOS, runtime.GOARCH)

		if targetAsset == nil {
			fmt.Printf("No compatible asset found for %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return
		}

		fmt.Printf("OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Println("Downloading update...")

		resp, err := http.Get(targetAsset.GetBrowserDownloadURL())
		if err != nil {
			fmt.Printf("Error downloading asset: %v\n", err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error downloading asset: status %s\n", resp.Status)
			return
		}
		rc := resp.Body
		defer rc.Close()

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

		counter := &WriteCounter{Total: uint64(resp.ContentLength)}

		if rc != nil {
			_, err = io.Copy(tmpFile, io.TeeReader(rc, counter))
			if err != nil {
				fmt.Printf("\nError writing to temp file: %v\n", err)
				tmpFile.Close()
				return
			}
		}
		fmt.Println()
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

func updateTemplates() {
	fmt.Println("Updating report templates from repository...")

	templatesDir, err := config.GetTemplatesDir()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		fmt.Printf("Error creating templates directory: %v\n", err)
		os.Exit(1)
	}

	baseURL := "https://raw.githubusercontent.com/aancw/pentlog/main/assets/templates/"
	files := []string{"report.html", "report.css"}

	for _, file := range files {
		destPath := filepath.Join(templatesDir, file)
		url := baseURL + file

		fmt.Printf("  Downloading %s... ", file)
		if err := downloadTemplateFile(url, destPath); err != nil {
			fmt.Printf("FAIL (%v)\n", err)
		} else {
			fmt.Println("OK")
		}
	}

	fmt.Println("Templates updated successfully.")
}

func downloadTemplateFile(url, dest string) error {
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
	updateCmd.Flags().BoolVarP(&updateTemplateFlag, "template", "t", false, "Update report templates from repository")
	rootCmd.AddCommand(updateCmd)
}

type WriteCounter struct {
	Total   uint64
	Current uint64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Current += uint64(n)
	wc.PrintProgress()
	return n, nil
}

func (wc *WriteCounter) PrintProgress() {
	if wc.Total == 0 {
		fmt.Printf("\rDownloading... %s", utils.FormatBytes(int64(wc.Current)))
	} else {
		percent := float64(wc.Current) * 100 / float64(wc.Total)
		fmt.Printf("\rDownloading... %.2f%% (%s / %s)", percent, utils.FormatBytes(int64(wc.Current)), utils.FormatBytes(int64(wc.Total)))
	}
}

func findCompatibleAsset(assets []*github.ReleaseAsset, goos, goarch string) *github.ReleaseAsset {
	targetName := fmt.Sprintf("pentlog-%s-%s", goos, goarch)

	for _, asset := range assets {
		if asset.GetName() == targetName {
			return asset
		}
	}
	return nil
}
