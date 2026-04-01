package cmd

import (
	"fmt"
	"pentlog/pkg/config"
	"pentlog/pkg/errors"
	"pentlog/pkg/utils"
	"time"

	"github.com/spf13/cobra"
)

var targetCmd = &cobra.Command{
	Use:   "target",
	Short: "Manage engagement targets",
	Long:  `Add, list, switch, and remove targets within the current engagement.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var targetAddCmd = &cobra.Command{
	Use:   "add [name] [ip]",
	Short: "Add a new target to the engagement",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := config.Manager()
		ctx, err := mgr.LoadContext()
		if err != nil {
			errors.NoContext().Fatal()
		}

		name := ""
		ip := ""

		if len(args) >= 1 {
			name = args[0]
		}
		if len(args) >= 2 {
			ip = args[1]
		}

		if name == "" {
			name = utils.PromptString("Target Name", "")
		}
		if ip == "" {
			ip = utils.PromptString("Target IP/Host", "")
		}

		if name == "" {
			errors.NewError(errors.InvalidInput, "Target name is required").Fatal()
		}

		targets, err := mgr.LoadTargets()
		if err != nil {
			errors.FromError(errors.Generic, "Failed to load targets", err).Fatal()
		}

		for _, t := range targets.Targets {
			if t.Name == name {
				errors.NewError(errors.InvalidInput, fmt.Sprintf("Target '%s' already exists", name)).
					AddSolution("Use a different name or remove the existing target first").
					AddSolution(fmt.Sprintf("$ pentlog target remove %s", name)).
					Fatal()
			}
		}

		targets.Targets = append(targets.Targets, config.Target{Name: name, IP: ip})

		if err := mgr.SaveTargets(targets); err != nil {
			errors.FromError(errors.Generic, "Failed to save targets", err).Fatal()
		}

		fmt.Printf("✓ Target '%s' added", name)
		if ip != "" {
			fmt.Printf(" (%s)", ip)
		}
		fmt.Println()

		switchNow := utils.PromptSelect("Switch to this target now?", []string{"Yes", "No"})
		if switchNow == "Yes" {
			ctx.Target = name
			ctx.TargetIP = ip
			ctx.Timestamp = time.Now().Format(time.RFC3339)
			if err := mgr.SaveContext(ctx); err != nil {
				errors.FromError(errors.Generic, "Failed to save context", err).Fatal()
			}
			fmt.Printf("✓ Switched to target: %s\n", name)
			printSummary(*ctx)
		}
	},
}

var targetListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all targets in the engagement",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := config.Manager()
		ctx, err := mgr.LoadContext()
		if err != nil {
			errors.NoContext().Fatal()
		}

		targets, err := mgr.LoadTargets()
		if err != nil {
			errors.FromError(errors.Generic, "Failed to load targets", err).Fatal()
		}

		if len(targets.Targets) == 0 {
			fmt.Println("No targets configured.")
			fmt.Println("Add one with: pentlog target add <name> <ip>")
			return
		}

		var lines []string
		for _, t := range targets.Targets {
			marker := "  "
			if t.Name == ctx.Target {
				marker = "▸ "
			}
			line := fmt.Sprintf("%s%-20s %s", marker, t.Name, t.IP)
			lines = append(lines, line)
		}

		utils.PrintBox("Targets", lines)
	},
}

var targetSwitchCmd = &cobra.Command{
	Use:   "switch [name]",
	Short: "Switch to a different target",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := config.Manager()
		ctx, err := mgr.LoadContext()
		if err != nil {
			errors.NoContext().Fatal()
		}

		targets, err := mgr.LoadTargets()
		if err != nil {
			errors.FromError(errors.Generic, "Failed to load targets", err).Fatal()
		}

		if len(targets.Targets) == 0 {
			fmt.Println("No targets configured.")
			fmt.Println("Add one with: pentlog target add <name> <ip>")
			return
		}

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		if name == "" {
			var items []string
			for _, t := range targets.Targets {
				label := t.Name
				if t.IP != "" {
					label = fmt.Sprintf("%s (%s)", t.Name, t.IP)
				}
				if t.Name == ctx.Target {
					label += " [current]"
				}
				items = append(items, label)
			}
			idx := utils.SelectItem("Select Target", items)
			if idx == -1 {
				return
			}
			name = targets.Targets[idx].Name
		}

		var found *config.Target
		for _, t := range targets.Targets {
			if t.Name == name {
				t := t
				found = &t
				break
			}
		}

		if found == nil {
			errors.NewError(errors.InvalidInput, fmt.Sprintf("Target '%s' not found", name)).
				AddSolution("$ pentlog target list    # See available targets").
				Fatal()
		}

		ctx.Target = found.Name
		ctx.TargetIP = found.IP
		ctx.Timestamp = time.Now().Format(time.RFC3339)

		if err := mgr.SaveContext(ctx); err != nil {
			errors.FromError(errors.Generic, "Failed to save context", err).Fatal()
		}

		fmt.Printf("✓ Switched to target: %s\n", found.Name)
		printSummary(*ctx)
	},
}

var targetRemoveCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm"},
	Short:   "Remove a target from the engagement",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := config.Manager()
		ctx, err := mgr.LoadContext()
		if err != nil {
			errors.NoContext().Fatal()
		}

		targets, err := mgr.LoadTargets()
		if err != nil {
			errors.FromError(errors.Generic, "Failed to load targets", err).Fatal()
		}

		if len(targets.Targets) == 0 {
			fmt.Println("No targets configured.")
			return
		}

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		if name == "" {
			var items []string
			for _, t := range targets.Targets {
				label := t.Name
				if t.IP != "" {
					label = fmt.Sprintf("%s (%s)", t.Name, t.IP)
				}
				items = append(items, label)
			}
			idx := utils.SelectItem("Select Target to Remove", items)
			if idx == -1 {
				return
			}
			name = targets.Targets[idx].Name
		}

		found := false
		var remaining []config.Target
		for _, t := range targets.Targets {
			if t.Name == name {
				found = true
				continue
			}
			remaining = append(remaining, t)
		}

		if !found {
			errors.NewError(errors.InvalidInput, fmt.Sprintf("Target '%s' not found", name)).Fatal()
		}

		targets.Targets = remaining
		if err := mgr.SaveTargets(targets); err != nil {
			errors.FromError(errors.Generic, "Failed to save targets", err).Fatal()
		}

		if ctx.Target == name {
			ctx.Target = ""
			ctx.TargetIP = ""
			ctx.Timestamp = time.Now().Format(time.RFC3339)
			if err := mgr.SaveContext(ctx); err != nil {
				errors.FromError(errors.Generic, "Failed to save context", err).Fatal()
			}
			fmt.Printf("✓ Target '%s' removed (was active, context cleared)\n", name)
		} else {
			fmt.Printf("✓ Target '%s' removed\n", name)
		}
	},
}

var targetClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the active target without removing it",
	Run: func(cmd *cobra.Command, args []string) {
		mgr := config.Manager()
		ctx, err := mgr.LoadContext()
		if err != nil {
			errors.NoContext().Fatal()
		}

		if ctx.Target == "" {
			fmt.Println("No active target set.")
			return
		}

		prev := ctx.Target
		ctx.Target = ""
		ctx.TargetIP = ""
		ctx.Timestamp = time.Now().Format(time.RFC3339)

		if err := mgr.SaveContext(ctx); err != nil {
			errors.FromError(errors.Generic, "Failed to save context", err).Fatal()
		}

		fmt.Printf("✓ Active target cleared (was: %s)\n", prev)
	},
}

func init() {
	targetCmd.AddCommand(targetAddCmd)
	targetCmd.AddCommand(targetListCmd)
	targetCmd.AddCommand(targetSwitchCmd)
	targetCmd.AddCommand(targetRemoveCmd)
	targetCmd.AddCommand(targetClearCmd)
	rootCmd.AddCommand(targetCmd)
}
