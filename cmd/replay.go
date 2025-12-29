package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"pentlog/pkg/logs"
	"runtime"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var replaySpeed float64

var replayCmd = &cobra.Command{
	Use:   "replay <id>",
	Short: "Replay a session",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if runtime.GOOS == "darwin" {
			fmt.Println("Warning: Session replay is not available on macOS.")
			os.Exit(1)
		}

		if _, err := exec.LookPath("scriptreplay"); err != nil {
			fmt.Println("Error: 'scriptreplay' command not found.")
			os.Exit(1)
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("ID must be a number")
			os.Exit(1)
		}

		session, err := logs.GetSession(id)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if session.TimingPath == "" {
			fmt.Println("Error: timing file missing for this session; cannot replay.")
			os.Exit(1)
		}

		fmt.Printf("Replaying session %d (%s) at %.1fx speed...\n", id, session.DisplayPath, replaySpeed)
		fmt.Println("Press Ctrl+C to stop early. Terminal will be restored automatically.")

		// Use --divisor for speed control (divisor = speed, e.g. 2.0x speed = divisor 2)
		divisor := fmt.Sprintf("%f", replaySpeed)
		c := exec.Command("scriptreplay", "--divisor", divisor, session.TimingPath, session.Path)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigCh)
		defer close(sigCh)
		go func() {
			for range sigCh {
				// Forward Ctrl+C to scriptreplay so it exits cleanly.
				if c.Process != nil {
					_ = c.Process.Signal(os.Interrupt)
				}
			}
		}()
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		c.Stdin = os.Stdin

		err = c.Run()

		// Reset terminal state using reset -I to avoid losing scrollback; fall back to stty sane.
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
