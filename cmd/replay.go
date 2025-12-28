package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"pentlog/pkg/logs"
	"strconv"

	"github.com/spf13/cobra"
)

var replayCmd = &cobra.Command{
	Use:   "replay <id>",
	Short: "Replay a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("ID must be a number")
			os.Exit(1)
		}

		path, err := logs.GetSessionPath(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Replaying session %d (%s)...\n", id, path)

		c := exec.Command("tlog-play", path)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		if err := c.Run(); err != nil {
			fmt.Printf("Error during replay: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(replayCmd)
}