package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

type ttyrecHeader struct {
	Sec  uint32
	Usec uint32
	Len  uint32
}

func getLastTimestamp(ttyPath string) (uint32, uint32, error) {
	data, err := os.ReadFile(ttyPath)
	if err != nil {
		return 0, 0, err
	}

	if len(data) == 0 {
		return 0, 0, nil
	}

	var lastSec, lastUsec uint32
	reader := bytes.NewReader(data)

	for reader.Len() >= 12 {
		var header ttyrecHeader
		if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
			break
		}
		if reader.Len() < int(header.Len) {
			break
		}
		lastSec = header.Sec
		lastUsec = header.Usec
		reader.Seek(int64(header.Len), 1)
	}

	return lastSec, lastUsec, nil
}

func writeTtyrecRecord(f *os.File, sec, usec uint32, data []byte) error {
	header := ttyrecHeader{
		Sec:  sec,
		Usec: usec,
		Len:  uint32(len(data)),
	}
	if err := binary.Write(f, binary.LittleEndian, header); err != nil {
		return err
	}
	_, err := f.Write(data)
	return err
}

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause the current recording session",
	Long: `Pause the current recording session without creating a new session.

This allows operators to take breaks (e.g., during OSCP exams or when stepping
away from sensitive environments) while maintaining a single continuous session.

When paused:
- A pause marker is written to the session log
- The session state is updated to "paused"
- Recording stops until 'pentlog resume' is called

Use cases:
- OSCP exam: Take a break without creating multiple disjointed sessions
- Client engagement: Pause before entering sensitive areas
- Clean evidence trails with single continuous session per phase`,
	Run: func(cmd *cobra.Command, args []string) {
		sessionPath := os.Getenv("PENTLOG_SESSION_LOG_PATH")
		if sessionPath == "" {
			errors.NewError(errors.Generic, "not in an active pentlog session").
				AddReason("PENTLOG_SESSION_LOG_PATH environment variable is not set").
				AddSolution("Run 'pentlog shell' first to start a session").
				Fatal()
		}

		// Check if already paused
		pauseMarkerFile := sessionPath + ".pause_marker"
		if _, err := os.Stat(pauseMarkerFile); err == nil {
			fmt.Println("⚠️  Session is already paused. Run 'pentlog resume' to continue.")
			return
		}

		// Get the last timestamp from the tty file
		lastSec, lastUsec, err := getLastTimestamp(sessionPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not read last timestamp: %v\n", err)
			lastSec = uint32(time.Now().Unix())
		}

		// Write pause marker to session log file with proper ttyrec format
		pauseTime := time.Now()
		pauseMarker := "\r\n\r\n═══════════════════════════════════════════════════════════════\r\n"
		pauseMarker += fmt.Sprintf("                    SESSION PAUSED: %s\r\n", pauseTime.Format(time.RFC3339))
		pauseMarker += "═══════════════════════════════════════════════════════════════\r\n\r\n"

		// Append marker to the tty file with proper ttyrec header
		f, err := utils.OpenPrivateFile(sessionPath, os.O_APPEND|os.O_WRONLY)
		if err != nil {
			errors.FromError(errors.FileNotFound, "failed to write pause marker", err).
				AddReason("Could not open session log file for writing").
				Fatal()
		}
		defer f.Close()

		// Write with timestamp 2 seconds after last record
		markerSec := lastSec + 2
		if err := writeTtyrecRecord(f, markerSec, lastUsec, []byte(pauseMarker)); err != nil {
			errors.FromError(errors.Generic, "failed to write pause marker", err).
				AddReason("Could not write to session log file").
				Fatal()
		}

		// Create pause marker file with timestamp
		if err := utils.WritePrivateFile(pauseMarkerFile, []byte(pauseTime.Format(time.RFC3339))); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not create pause marker file: %v\n", err)
		}
		if sessionIDValue := os.Getenv("PENTLOG_SESSION_ID"); sessionIDValue != "" {
			if sessionID, err := strconv.ParseInt(sessionIDValue, 10, 64); err == nil {
				if err := logs.PauseSession(sessionID); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Could not update session state to paused: %v\n", err)
				}
			}
		}

		fmt.Println()
		fmt.Println("⏸️  Session paused successfully!")
		fmt.Printf("   Time: %s\n", pauseTime.Format("2006-01-02 15:04:05"))
		fmt.Println()
		fmt.Println("   Recording is paused. The shell remains active.")
		fmt.Println("   Run 'pentlog resume' to continue recording.")
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
}
