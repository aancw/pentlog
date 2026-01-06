package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"pentlog/pkg/logs"
	"pentlog/pkg/system"
	"pentlog/pkg/utils"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var replaySpeed float64

var replayCmd = &cobra.Command{
	Use:   "replay [id]",
	Short: "Replay a session (Interactive)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := system.CheckReplayDependencies(); err != nil {
			fmt.Printf("Warning: %v\n", err)
			fmt.Println("Replay might fail or be limited.")
		}

		var id int
		var err error

		if len(args) > 0 {
			id, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("Invalid session ID: %s\n", args[0])
				os.Exit(1)
			}
		} else {
			sessions, err := logs.ListSessions()
			if err != nil {
				fmt.Printf("Error listing sessions: %v\n", err)
				os.Exit(1)
			}
			if len(sessions) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			startIdx := 0
			if len(sessions) > 15 {
				startIdx = len(sessions) - 15
			}
			displaySessions := sessions[startIdx:]

			var items []string
			for _, s := range displaySessions {
				items = append(items, fmt.Sprintf("ID %d | %s | %s", s.ID, s.ModTime, s.DisplayPath))
			}

			fmt.Println("Recent Sessions:")
			idx := utils.SelectItem("Select Session to Replay:", items)
			if idx == -1 {
				fmt.Println("Selection canceled.")
				return
			}
			id = displaySessions[idx].ID
		}

		session, err := logs.GetSession(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if session.Path == "" {
			fmt.Println("Error: session file missing; cannot replay.")
			os.Exit(1)
		}

		fmt.Printf("Replaying session %d (%s) at %.1fx speed...\n", id, session.DisplayPath, replaySpeed)
		fmt.Println("Press Ctrl+C to stop early. Terminal will be restored automatically.")

		speed := fmt.Sprintf("%f", replaySpeed)
		c := exec.Command("ttyplay", "-s", speed, session.Path)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigCh)
		defer close(sigCh)
		go func() {
			for range sigCh {
				// Forward Ctrl+C to ttyplay so it exits cleanly.
				if c.Process != nil {
					_ = c.Process.Signal(os.Interrupt)
				}
			}
		}()
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		err = c.Run()

		// Reset terminal state
		// ttyplay usually handles this, but a safety reset is good
		if errReset := exec.Command("reset", "-I").Run(); errReset != nil {
			_ = exec.Command("stty", "sane").Run()
		}

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signal() == os.Interrupt {
					fmt.Println("Replay interrupted (Ctrl+C). Terminal refreshed.")
					return
				}
			}
			fmt.Printf("Error during replay: %v\n", err)
		}
	},
}

func init() {
	replayCmd.Flags().Float64VarP(&replaySpeed, "speed", "s", 1.0, "Replay speed (e.g. 2.0 for 2x, 0.5 for half speed)")
	rootCmd.AddCommand(replayCmd)
}
