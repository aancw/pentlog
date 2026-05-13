package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"pentlog/pkg/errors"
	"pentlog/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

func getLastTimestampResume(ttyPath string) (uint32, uint32, error) {
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
		var header struct {
			Sec  uint32
			Usec uint32
			Len  uint32
		}
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

func writeTtyrecRecordResume(f *os.File, sec, usec uint32, data []byte) error {
	header := struct {
		Sec  uint32
		Usec uint32
		Len  uint32
	}{
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

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a paused recording session",
	Long: `Resume a previously paused recording session.

This command continues recording from where the session was paused,
maintaining a single continuous session for the entire engagement phase.

When resumed:
- A resume marker is written to the session log
- The pause marker file is removed
- Recording continues from the pause point

Use cases:
- Return from a break during OSCP exam
- Resume work after stepping away from sensitive environment
- Continue recording after temporary pause`,
	Run: func(cmd *cobra.Command, args []string) {
		sessionPath := os.Getenv("PENTLOG_SESSION_LOG_PATH")
		if sessionPath == "" {
			errors.NewError(errors.Generic, "not in an active pentlog session").
				AddReason("PENTLOG_SESSION_LOG_PATH environment variable is not set").
				AddSolution("Run 'pentlog shell' first to start a session").
				Fatal()
		}

		// Check if there's a pause marker
		pauseMarkerFile := sessionPath + ".pause_marker"
		if _, err := os.Stat(pauseMarkerFile); os.IsNotExist(err) {
			fmt.Println("⚠️  Session is not paused. No pause marker found.")
			return
		}

		// Read pause time for display
		pauseData, err := os.ReadFile(pauseMarkerFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not read pause marker file: %v\n", err)
		}

		// Get the last timestamp from the tty file
		lastSec, lastUsec, err := getLastTimestampResume(sessionPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not read last timestamp: %v\n", err)
			lastSec = uint32(time.Now().Unix())
		}

		// Write resume marker to session log file with proper ttyrec format
		resumeTime := time.Now()
		resumeMarker := "\r\n\r\n═══════════════════════════════════════════════════════════════\r\n"
		resumeMarker += fmt.Sprintf("                    SESSION RESUMED: %s\r\n", resumeTime.Format(time.RFC3339))
		resumeMarker += "═══════════════════════════════════════════════════════════════\r\n\r\n"

		// Append marker to the tty file with proper ttyrec header
		f, err := utils.OpenPrivateFile(sessionPath, os.O_APPEND|os.O_WRONLY)
		if err != nil {
			errors.FromError(errors.FileNotFound, "failed to write resume marker", err).
				AddReason("Could not open session log file for writing").
				Fatal()
		}
		defer f.Close()

		// Write with timestamp 2 seconds after last record
		markerSec := lastSec + 2
		if err := writeTtyrecRecordResume(f, markerSec, lastUsec, []byte(resumeMarker)); err != nil {
			errors.FromError(errors.Generic, "failed to write resume marker", err).
				AddReason("Could not write to session log file").
				Fatal()
		}

		// Remove pause marker file
		if err := os.Remove(pauseMarkerFile); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not remove pause marker file: %v\n", err)
		}

		fmt.Println()
		fmt.Println("▶️  Session resumed successfully!")
		fmt.Printf("   Time: %s\n", resumeTime.Format("2006-01-02 15:04:05"))
		if len(pauseData) > 0 {
			if pauseTime, err := time.Parse(time.RFC3339, string(pauseData)); err == nil {
				duration := resumeTime.Sub(pauseTime)
				fmt.Printf("   Paused for: %s\n", duration.Round(time.Second))
			}
		}
		fmt.Println()
		fmt.Println("   Recording is now active.")
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
