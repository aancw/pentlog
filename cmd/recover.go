package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"strconv"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const staleSessionTimeout = 5 * time.Minute

var recoverCmd = &cobra.Command{
	Use:   "recover",
	Short: "Recover and manage crashed or stale sessions",
	Long: `Recover crashed or stale sessions that were not properly closed.

This command helps salvage partial session data from sessions that were
terminated unexpectedly (SSH disconnect, OOM, SIGKILL, etc.).

It can:
- List all crashed/incomplete sessions
- Mark stale active sessions as crashed (no heartbeat for 5+ minutes)
- Recover specific sessions by marking them as completed
- Clean up orphaned sessions (database entries with missing files)
- Show session recovery details`,
	Run: func(cmd *cobra.Command, args []string) {
		listOnly, _ := cmd.Flags().GetBool("list")
		markStale, _ := cmd.Flags().GetBool("mark-stale")
		recoverID, _ := cmd.Flags().GetInt("recover")
		recoverAll, _ := cmd.Flags().GetBool("recover-all")
		cleanOrphans, _ := cmd.Flags().GetBool("clean-orphans")

		if markStale {
			count, err := logs.MarkStaleSessions(staleSessionTimeout)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marking stale sessions: %v\n", err)
				os.Exit(1)
			}
			if count > 0 {
				fmt.Printf("Marked %d stale session(s) as crashed\n", count)
			} else {
				fmt.Println("No stale sessions found")
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

		crashed, err := logs.GetCrashedSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting crashed sessions: %v\n", err)
			os.Exit(1)
		}

		active, err := logs.GetActiveSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting active sessions: %v\n", err)
			os.Exit(1)
		}

		orphaned, err := logs.GetOrphanedSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting orphaned sessions: %v\n", err)
			os.Exit(1)
		}

		if len(crashed) == 0 && len(active) == 0 && len(orphaned) == 0 {
			fmt.Println("✓ No crashed, orphaned, or stale sessions found")
			return
		}

		if listOnly {
			printSessionStatus(crashed, active, orphaned)
			return
		}

		runInteractiveRecover(crashed, active, orphaned)
	},
}

func printSessionStatus(crashed, active, orphaned []logs.Session) {
	if len(active) > 0 {
		fmt.Printf("\n⚠️  Active Sessions (%d) - may be orphaned if no shell is running:\n", len(active))
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for _, s := range active {
			lastSync := "unknown"
			if s.LastSyncAt != "" {
				if t, err := time.Parse(time.RFC3339, s.LastSyncAt); err == nil {
					lastSync = utils.FormatRelativeTime(t)
				}
			}
			fileSize := s.Size
			if info, err := os.Stat(s.Path); err == nil {
				fileSize = info.Size()
			}
			fmt.Printf("  [%d] %s/%s/%s\n", s.ID, s.Metadata.Client, s.Metadata.Engagement, s.Metadata.Phase)
			fmt.Printf("      File: %s (%s)\n", s.Filename, utils.FormatSize(fileSize))
			fmt.Printf("      Last heartbeat: %s\n", lastSync)
		}
	}

	if len(crashed) > 0 {
		fmt.Printf("\n❌ Crashed Sessions (%d):\n", len(crashed))
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for _, s := range crashed {
			lastSync := "unknown"
			if s.LastSyncAt != "" {
				if t, err := time.Parse(time.RFC3339, s.LastSyncAt); err == nil {
					lastSync = utils.FormatRelativeTime(t)
				}
			}
			fileSize := s.Size
			if info, err := os.Stat(s.Path); err == nil {
				fileSize = info.Size()
			}
			fmt.Printf("  [%d] %s/%s/%s\n", s.ID, s.Metadata.Client, s.Metadata.Engagement, s.Metadata.Phase)
			fmt.Printf("      File: %s (%s)\n", s.Filename, utils.FormatSize(fileSize))
			fmt.Printf("      Crashed: %s\n", lastSync)
		}
	}

	if len(orphaned) > 0 {
		fmt.Printf("\n🗑️  Orphaned Sessions (%d) - database entries with missing files:\n", len(orphaned))
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for _, s := range orphaned {
			fmt.Printf("  [%d] %s/%s/%s\n", s.ID, s.Metadata.Client, s.Metadata.Engagement, s.Metadata.Phase)
			fmt.Printf("      Missing: %s\n", s.Path)
		}
	}

	fmt.Println()
	fmt.Println("Run 'pentlog recover' without --list for interactive recovery")
	fmt.Println("Run 'pentlog recover --recover <ID>' to recover a specific session")
	fmt.Println("Run 'pentlog recover --mark-stale' to mark stale active sessions as crashed")
	fmt.Println("Run 'pentlog recover --clean-orphans' to remove orphaned database entries")
}

func runInteractiveRecover(crashed, active, orphaned []logs.Session) {
	printSessionStatus(crashed, active, orphaned)

	choices := []string{}
	if len(active) > 0 {
		choices = append(choices, "Mark stale active sessions as crashed")
	}
	if len(crashed) > 0 {
		choices = append(choices, "Recover a crashed session")
		choices = append(choices, "Recover all crashed sessions")
	}
	if len(orphaned) > 0 {
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
	case "Mark stale active sessions as crashed":
		count, err := logs.MarkStaleSessions(staleSessionTimeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Printf("Marked %d session(s) as crashed\n", count)

	case "Recover a crashed session":
		if len(crashed) == 0 {
			fmt.Println("No crashed sessions to recover")
			return
		}
		selectAndRecover(crashed)

	case "Recover all crashed sessions":
		if err := recoverAllCrashed(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

	case "Clean up orphaned sessions":
		if err := cleanOrphanedSessions(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

	case "Cancel":
		return
	}
}

func selectAndRecover(crashed []logs.Session) {
	items := make([]string, len(crashed))
	for i, s := range crashed {
		items[i] = fmt.Sprintf("[%d] %s/%s - %s (%s)",
			s.ID, s.Metadata.Client, s.Metadata.Phase, s.Filename, utils.FormatSize(s.Size))
	}

	prompt := promptui.Select{
		Label: "Select session to recover",
		Items: items,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return
	}

	session := crashed[idx]
	if err := logs.RecoverSession(session.ID); err != nil {
		fmt.Fprintf(os.Stderr, "Error recovering session: %v\n", err)
		return
	}

	fmt.Printf("✓ Session %d recovered successfully\n", session.ID)
	fmt.Printf("  File: %s\n", session.Path)
	fmt.Printf("  Size: %s\n", utils.FormatSize(session.Size))
}

func recoverSessionByID(id int) error {
	session, err := logs.GetSession(id)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
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
	recoverCmd.Flags().BoolP("list", "l", false, "List crashed/stale sessions without interactive mode")
	recoverCmd.Flags().Bool("mark-stale", false, "Mark active sessions with no heartbeat as crashed")
	recoverCmd.Flags().IntP("recover", "r", 0, "Recover a specific session by ID")
	recoverCmd.Flags().Bool("recover-all", false, "Recover all crashed sessions")
	recoverCmd.Flags().Bool("clean-orphans", false, "Remove database entries for sessions with missing files")
}

func parseSessionID(s string) int {
	id, _ := strconv.Atoi(s)
	return id
}
