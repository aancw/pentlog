package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Review and repair session lifecycle state",
	Long: `Review and repair session lifecycle state for recordings that were
interrupted, paused, archived, or left with stale heartbeats.

PentLog now separates:
- likely-live sessions (recent heartbeat or verified recorder PID)
- review-needed sessions (heartbeat is stale, but crash cannot be proved safely)
- stale sessions (recorder proved dead and safe to mark crashed)
- crashed sessions ready to recover
- orphaned database entries with missing evidence files`,
	Run: func(cmd *cobra.Command, args []string) {
		timeout, err := recoverTimeout(cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing timeout: %v\n", err)
			os.Exit(1)
		}

		listOnly, _ := cmd.Flags().GetBool("list")
		markStale, _ := cmd.Flags().GetBool("mark-stale")
		recoverID, _ := cmd.Flags().GetInt("recover")
		forceCrashID, _ := cmd.Flags().GetInt("force-crash")
		recoverAll, _ := cmd.Flags().GetBool("recover-all")
		cleanOrphans, _ := cmd.Flags().GetBool("clean-orphans")

		if markStale {
			count, err := logs.MarkStaleSessions(timeout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marking stale sessions: %v\n", err)
				os.Exit(1)
			}
			if count > 0 {
				fmt.Printf("Marked %d definitely stale session(s) as crashed\n", count)
			} else {
				fmt.Println("No definitely stale sessions found")
			}
			return
		}

		if forceCrashID > 0 {
			if err := forceCrashSessionByID(forceCrashID); err != nil {
				fmt.Fprintf(os.Stderr, "Error forcing crash state: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if recoverID > 0 {
			if err := recoverSessionByID(recoverID); err != nil {
				fmt.Fprintf(os.Stderr, "Error recovering session: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if recoverAll {
			if err := recoverAllCrashed(); err != nil {
				fmt.Fprintf(os.Stderr, "Error recovering sessions: %v\n", err)
				os.Exit(1)
			}
			return
		}

		if cleanOrphans {
			if err := cleanOrphanedSessions(); err != nil {
				fmt.Fprintf(os.Stderr, "Error cleaning orphaned sessions: %v\n", err)
				os.Exit(1)
			}
			return
		}

		overview, err := logs.GetRecoveryOverview(timeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading recovery overview: %v\n", err)
			os.Exit(1)
		}

		if len(overview.Active) == 0 && len(overview.Paused) == 0 && len(overview.ReviewNeeded) == 0 &&
			len(overview.Stale) == 0 && len(overview.Crashed) == 0 && len(overview.Orphaned) == 0 {
			fmt.Println("✓ No lifecycle issues found")
			return
		}

		if listOnly {
			printRecoveryOverview(overview)
			return
		}

		runInteractiveRecover(overview)
	},
}

func recoverTimeout(cmd *cobra.Command) (time.Duration, error) {
	value, _ := cmd.Flags().GetString("timeout")
	value = strings.TrimSpace(value)
	if value == "" {
		return getConfiguredStaleTimeout(), nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}
	if duration <= 0 {
		return 0, fmt.Errorf("timeout must be greater than zero")
	}
	return duration, nil
}

func printRecoveryOverview(overview logs.RecoveryOverview) {
	fmt.Printf("Configured stale timeout: %s\n", overview.Timeout)

	printRecoverySection("Likely-Live Active Sessions", overview.Active)
	printRecoverySection("Paused Sessions", overview.Paused)
	printRecoverySection("Review-Needed Sessions", overview.ReviewNeeded)
	printRecoverySection("Definitely Stale Sessions", overview.Stale)
	printRecoverySection("Crashed Sessions", overview.Crashed)

	if len(overview.Orphaned) > 0 {
		fmt.Printf("\n🗑️  Orphaned Sessions (%d) - non-archived database entries with missing files:\n", len(overview.Orphaned))
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for _, s := range overview.Orphaned {
			fmt.Printf("  [%d] %s/%s/%s\n", s.ID, s.Metadata.Client, s.Metadata.Engagement, s.Metadata.Phase)
			fmt.Printf("      Missing: %s\n", s.Path)
		}
	}

	fmt.Println()
	fmt.Println("Run 'pentlog recover' without --list for interactive actions")
	fmt.Println("Run 'pentlog recover --mark-stale' to crash only definitely stale sessions")
	fmt.Println("Run 'pentlog recover --force-crash <ID>' to manually crash a review-needed active/paused session")
	fmt.Println("Run 'pentlog recover --recover <ID>' to recover a crashed session")
	fmt.Println("Run 'pentlog recover --clean-orphans' to remove orphaned database rows")
}

func printRecoverySection(title string, sessions []logs.RecoveryCandidate) {
	if len(sessions) == 0 {
		return
	}

	fmt.Printf("\n%s (%d):\n", title, len(sessions))
	fmt.Println("─────────────────────────────────────────────────────────────────")
	for _, candidate := range sessions {
		s := candidate.Session
		fileSize := s.Size
		if info, err := os.Stat(s.Path); err == nil {
			fileSize = info.Size()
		}

		fmt.Printf("  [%d] %s/%s/%s\n", s.ID, s.Metadata.Client, s.Metadata.Engagement, s.Metadata.Phase)
		fmt.Printf("      State: %s | File: %s (%s)\n", s.State, s.Filename, utils.FormatSize(fileSize))
		if candidate.LastSeenAge != "" {
			fmt.Printf("      Last heartbeat: %s ago\n", candidate.LastSeenAge)
		} else if s.LastSyncAt != "" {
			fmt.Printf("      Last heartbeat: %s\n", s.LastSyncAt)
		}
		if s.RecorderPID > 0 {
			hostLabel := s.Hostname
			if hostLabel == "" {
				hostLabel = "unknown-host"
			}
			fmt.Printf("      Recorder PID: %d on %s\n", s.RecorderPID, hostLabel)
		}
		if s.ResumeCount > 0 {
			fmt.Printf("      Resume count: %d\n", s.ResumeCount)
		}
		if s.StartedAt != "" {
			fmt.Printf("      Started: %s\n", s.StartedAt)
		}
		if s.EndedAt != "" {
			fmt.Printf("      Ended: %s\n", s.EndedAt)
		}
		fmt.Printf("      Reason: %s\n", candidate.Reason)
	}
}

func runInteractiveRecover(overview logs.RecoveryOverview) {
	printRecoveryOverview(overview)

	choices := []string{}
	if len(overview.Stale) > 0 {
		choices = append(choices, "Mark definitely stale sessions as crashed")
	}
	if len(overview.ReviewNeeded) > 0 {
		choices = append(choices, "Force-mark a review-needed session as crashed")
	}
	if len(overview.Crashed) > 0 {
		choices = append(choices, "Recover a crashed session")
		choices = append(choices, "Recover all crashed sessions")
	}
	if len(overview.Orphaned) > 0 {
		choices = append(choices, "Clean up orphaned sessions")
	}
	choices = append(choices, "Cancel")

	prompt := promptui.Select{
		Label: "What would you like to do?",
		Items: choices,
	}

	_, result, err := prompt.Run()
	if err != nil {
		return
	}

	switch result {
	case "Mark definitely stale sessions as crashed":
		count, err := logs.MarkStaleSessions(overview.Timeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Printf("Marked %d session(s) as crashed\n", count)
	case "Force-mark a review-needed session as crashed":
		selectAndForceCrash(overview.ReviewNeeded)
	case "Recover a crashed session":
		selectAndRecover(overview.Crashed)
	case "Recover all crashed sessions":
		if err := recoverAllCrashed(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	case "Clean up orphaned sessions":
		if err := cleanOrphanedSessions(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

func selectAndRecover(crashed []logs.RecoveryCandidate) {
	items := make([]string, len(crashed))
	for i, candidate := range crashed {
		s := candidate.Session
		items[i] = fmt.Sprintf("[%d] %s/%s - %s (%s)", s.ID, s.Metadata.Client, s.Metadata.Phase, s.Filename, utils.FormatSize(s.Size))
	}

	prompt := promptui.Select{
		Label: "Select crashed session to recover",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return
	}

	session := crashed[idx].Session
	if err := logs.RecoverSession(session.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error recovering session: %v\n", err)
		return
	}

	fmt.Printf("✓ Session %d recovered successfully\n", session.ID)
	fmt.Printf("  File: %s\n", session.Path)
	fmt.Printf("  Size: %s\n", utils.FormatSize(session.Size))
}

func selectAndForceCrash(reviewNeeded []logs.RecoveryCandidate) {
	items := make([]string, len(reviewNeeded))
	for i, candidate := range reviewNeeded {
		s := candidate.Session
		items[i] = fmt.Sprintf("[%d] %s/%s - %s (%s)", s.ID, s.Metadata.Client, s.Metadata.Phase, s.Filename, candidate.Reason)
	}

	prompt := promptui.Select{
		Label: "Select review-needed session to force-mark crashed",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return
	}

	if err := forceCrashSessionByID(reviewNeeded[idx].Session.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error forcing crash state: %v\n", err)
	}
}

func recoverSessionByID(id int) error {
	session, err := logs.GetSession(id)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}
	if session.State != logs.SessionStateCrashed {
		return fmt.Errorf("session %d is %q, not crashed", id, session.State)
	}

	fileSize := session.Size
	if info, err := os.Stat(session.Path); err == nil {
		fileSize = info.Size()
	} else if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: Session file does not exist: %s\n", session.Path)
	}

	if err := logs.RecoverSession(id); err != nil {
		return err
	}

	fmt.Printf("✓ Session %d recovered successfully\n", id)
	fmt.Printf("  Client:     %s\n", session.Metadata.Client)
	fmt.Printf("  Engagement: %s\n", session.Metadata.Engagement)
	fmt.Printf("  Phase:      %s\n", session.Metadata.Phase)
	fmt.Printf("  File:       %s\n", session.Path)
	fmt.Printf("  Size:       %s\n", utils.FormatSize(fileSize))

	return nil
}

func forceCrashSessionByID(id int) error {
	session, err := logs.GetSession(id)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}
	if session.State != logs.SessionStateActive && session.State != logs.SessionStatePaused {
		return fmt.Errorf("session %d is %q, not active or paused", id, session.State)
	}

	if err := logs.UpdateSessionState(int64(id), logs.SessionStateCrashed); err != nil {
		return err
	}

	fmt.Printf("✓ Session %d marked as crashed\n", id)
	fmt.Printf("  Client:     %s\n", session.Metadata.Client)
	fmt.Printf("  Engagement: %s\n", session.Metadata.Engagement)
	fmt.Printf("  Phase:      %s\n", session.Metadata.Phase)
	return nil
}

func recoverAllCrashed() error {
	crashed, err := logs.GetCrashedSessions()
	if err != nil {
		return err
	}

	if len(crashed) == 0 {
		fmt.Println("No crashed sessions to recover")
		return nil
	}

	recovered := 0
	for _, s := range crashed {
		if err := logs.RecoverSession(s.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to recover session %d: %v\n", s.ID, err)
			continue
		}
		recovered++
	}

	fmt.Printf("✓ Recovered %d/%d crashed session(s)\n", recovered, len(crashed))
	return nil
}

func cleanOrphanedSessions() error {
	orphaned, err := logs.GetOrphanedSessions()
	if err != nil {
		return err
	}

	if len(orphaned) == 0 {
		fmt.Println("No orphaned sessions to clean up")
		return nil
	}

	fmt.Printf("Found %d orphaned session(s) with missing files:\n", len(orphaned))
	for _, s := range orphaned {
		fmt.Printf("  [%d] %s - %s\n", s.ID, s.Metadata.Client, s.Path)
	}

	prompt := promptui.Prompt{
		Label:     "Remove these database entries",
		IsConfirm: true,
	}

	_, err = prompt.Run()
	if err != nil {
		fmt.Println("Cancelled")
		return nil
	}

	deleted := 0
	for _, s := range orphaned {
		if err := logs.DeleteSession(s.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to delete session %d: %v\n", s.ID, err)
			continue
		}
		deleted++
	}

	fmt.Printf("✓ Removed %d/%d orphaned session(s) from database\n", deleted, len(orphaned))
	return nil
}

func init() {
	rootCmd.AddCommand(recoverCmd)
	recoverCmd.Flags().BoolP("list", "l", false, "List lifecycle state without interactive mode")
	recoverCmd.Flags().Bool("mark-stale", false, "Mark only definitely stale live sessions as crashed")
	recoverCmd.Flags().String("timeout", "", "Override stale-session timeout for this command (for example: 15m, 1h)")
	recoverCmd.Flags().IntP("recover", "r", 0, "Recover a specific crashed session by ID")
	recoverCmd.Flags().Int("force-crash", 0, "Force-mark a review-needed active or paused session as crashed")
	recoverCmd.Flags().Bool("recover-all", false, "Recover all crashed sessions")
	recoverCmd.Flags().Bool("clean-orphans", false, "Remove database entries for non-archived sessions with missing files")
}

func parseSessionID(s string) int {
	id, _ := strconv.Atoi(s)
	return id
}
