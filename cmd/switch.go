package cmd

import (
	"fmt"
	"os"
	"pentlog/pkg/config"
	"pentlog/pkg/errors"
	"pentlog/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [phase]",
	Short: "Switch to a different pentest phase (Interactive)",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := config.Manager()
		ctx, err := mgr.LoadContext()
		if err != nil {
			errors.NoContext().Fatal()
		}

		newPhase := ""
		if len(args) > 0 {
			if args[0] == "-" {
				switchBack()
				return
			}
			newPhase = args[0]
		}

		if newPhase == "" {
			choice := utils.SelectItem("Select Action", []string{
				"Select from History",
				"Enter Manual/New",
			})

			if choice == 0 {
				if listSessions() {
					return
				}
				os.Exit(0)
			}
		}

		if newPhase == "" {
			if ctx.Type == "Exam/Lab" {
				fmt.Printf("Current Target: %s\n", ctx.Engagement)
				newTarget := utils.PromptString("New Target Host/IP", ctx.Engagement)
				if newTarget != "" {
					ctx.Engagement = newTarget
					ctx.Scope = newTarget
				}
			}

			fmt.Printf("Current Phase: %s\n", ctx.Phase)
			newPhase = utils.PromptString("New Phase", "")
		}

		if newPhase == "" {
			errors.NewError(errors.Generic, "Phase cannot be empty").Fatal()
		}

		ctx.Phase = newPhase
		ctx.Timestamp = time.Now().Format(time.RFC3339)

		if err := mgr.SaveContext(ctx); err != nil {
			errors.FromError(errors.Generic, "Error saving context", err).Fatal()
		}

		printSummary(*ctx)
	},
}

func switchBack() {
	mgr := config.Manager()
	history, err := mgr.LoadContextHistory()
	if err != nil {
		errors.FromError(errors.Generic, "Error loading history", err).Fatal()
	}

	if len(history) < 2 {
		fmt.Println("No previous session found.")
		os.Exit(1)
	}

	target := history[len(history)-2]
	target.Timestamp = time.Now().Format(time.RFC3339)

	if err := mgr.SaveContext(&target); err != nil {
		errors.FromError(errors.Generic, "Error saving context", err).Fatal()
	}

	fmt.Println("\nSwitched to previous session:")
	printSummary(target)
}

func listSessions() bool {
	mgr := config.Manager()
	history, err := mgr.LoadContextHistory()
	if err != nil {
		errors.FromError(errors.Generic, "Error loading history", err).Print()
		return false
	}

	if len(history) == 0 {
		fmt.Println("No session history found.")
		return false
	}

	type key struct {
		Type       string
		Client     string
		Engagement string
	}
	seen := make(map[key]bool)
	var candidates []config.ContextData

	for i := len(history) - 1; i >= 0; i-- {
		ctx := history[i]
		k := key{ctx.Type, ctx.Client, ctx.Engagement}
		if !seen[k] {
			seen[k] = true
			candidates = append(candidates, ctx)
		}
	}

	if len(candidates) == 0 {
		fmt.Println("No history available.")
		return false
	}

	var items []string
	for _, c := range candidates {
		label := fmt.Sprintf("[%s] %s - %s", c.Type, c.Client, c.Engagement)
		items = append(items, label)
	}

	idx := utils.SelectItem("Select Session", items)
	if idx == -1 {
		return false
	}

	selected := candidates[idx]
	selected.Timestamp = time.Now().Format(time.RFC3339)

	if err := mgr.SaveContext(&selected); err != nil {
		errors.FromError(errors.Generic, "Error saving context", err).Print()
		return false
	}

	fmt.Println("\nSwitched to session:")
	printSummary(selected)
	return true
}

func printSummary(ctx config.ContextData) {
	fmt.Printf("\nSwitched to phase: %s\n", ctx.Phase)

	summary := []string{
		"---------------------------------------------------",
	}

	if ctx.Type == "Exam/Lab" {
		summary = append(summary, fmt.Sprintf("Exam/Lab Name: %s", ctx.Client))
		summary = append(summary, fmt.Sprintf("Target:        %s", ctx.Engagement))
	} else {
		summary = append(summary, fmt.Sprintf("Client:     %s", ctx.Client))
		summary = append(summary, fmt.Sprintf("Engagement: %s", ctx.Engagement))
		summary = append(summary, fmt.Sprintf("Scope:      %s", ctx.Scope))
	}
	summary = append(summary, fmt.Sprintf("Operator:   %s", ctx.Operator))
	summary = append(summary, fmt.Sprintf("Phase:      %s", ctx.Phase))
	summary = append(summary, "---------------------------------------------------")
	utils.PrintCenteredBlock(summary)
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
