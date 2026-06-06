package cmd

import (
	"fmt"
	"pentlog/pkg/config"
	"pentlog/pkg/utils"
	"strings"
	"time"
)

const (
	contextDisplayTimeLayout   = "2006-01-02 15:04"
	shellContextStaleThreshold = 8 * time.Hour
)

func buildContextSummaryLines(ctx config.ContextData) []string {
	lines := []string{}

	if ctx.Type == "Exam/Lab" {
		lines = append(lines, fmt.Sprintf("Exam/Lab Name: %s", ctx.Client))
		lines = append(lines, fmt.Sprintf("Target:        %s", ctx.Engagement))
	} else {
		lines = append(lines, fmt.Sprintf("Client:     %s", ctx.Client))
		lines = append(lines, fmt.Sprintf("Engagement: %s", ctx.Engagement))
		lines = append(lines, fmt.Sprintf("Scope:      %s", ctx.Scope))
	}

	lines = append(lines, fmt.Sprintf("Operator:   %s", ctx.Operator))
	lines = append(lines, fmt.Sprintf("Phase:      %s", ctx.Phase))

	if strings.TrimSpace(ctx.Target) != "" {
		lines = append(lines, fmt.Sprintf("Target:     %s", ctx.Target))
	}
	if strings.TrimSpace(ctx.TargetIP) != "" {
		lines = append(lines, fmt.Sprintf("Target IP:  %s", ctx.TargetIP))
	}

	return lines
}

func loadRecentContextChanges(mgr *config.ConfigManager, limit int) ([]string, error) {
	history, err := mgr.LoadContextHistory()
	if err != nil {
		return nil, err
	}
	return summarizeRecentContextChangesAt(history, limit, time.Now()), nil
}

func summarizeRecentContextChangesAt(history []config.ContextData, limit int, now time.Time) []string {
	if limit <= 0 || len(history) == 0 {
		return nil
	}

	lines := make([]string, 0, limit)
	for i := len(history) - 1; i >= 0 && len(lines) < limit; i-- {
		curr := history[i]

		var prev *config.ContextData
		if i > 0 {
			prev = &history[i-1]
		}

		delta := describeContextDelta(prev, curr)
		if delta == "" {
			continue
		}

		lines = append(lines, fmt.Sprintf("%s: %s", formatRelativeTimeAt(curr.Timestamp, now), delta))
	}

	return lines
}

func describeContextDelta(prev *config.ContextData, curr config.ContextData) string {
	if prev == nil {
		label := fmt.Sprintf("%s / %s / %s", curr.Client, curr.Engagement, curr.Phase)
		if target := formatTargetDisplay(curr.Target, curr.TargetIP); target != "not set" {
			label = fmt.Sprintf("%s, target %s", label, target)
		}
		return fmt.Sprintf("Context created for %s", label)
	}

	var changes []string

	if !equalTrimmed(prev.Client, curr.Client) {
		changes = append(changes, fmt.Sprintf("client %s -> %s", formatValue(prev.Client), formatValue(curr.Client)))
	}
	if !equalTrimmed(prev.Engagement, curr.Engagement) {
		changes = append(changes, fmt.Sprintf("engagement %s -> %s", formatValue(prev.Engagement), formatValue(curr.Engagement)))
	}
	if !equalTrimmed(prev.Phase, curr.Phase) {
		changes = append(changes, fmt.Sprintf("phase %s -> %s", formatValue(prev.Phase), formatValue(curr.Phase)))
	}

	prevTarget := formatTargetDisplay(prev.Target, prev.TargetIP)
	currTarget := formatTargetDisplay(curr.Target, curr.TargetIP)
	if prevTarget != currTarget {
		switch {
		case prevTarget == "not set":
			changes = append(changes, fmt.Sprintf("target -> %s", currTarget))
		case currTarget == "not set":
			changes = append(changes, fmt.Sprintf("target cleared (was %s)", prevTarget))
		default:
			changes = append(changes, fmt.Sprintf("target %s -> %s", prevTarget, currTarget))
		}
	}

	return strings.Join(changes, "; ")
}

func describePendingContextMutation(saved, effective config.ContextData) string {
	var changes []string

	if !equalTrimmed(saved.Phase, effective.Phase) {
		changes = append(changes, fmt.Sprintf("phase %s -> %s", formatValue(saved.Phase), formatValue(effective.Phase)))
	}

	savedTarget := formatTargetDisplay(saved.Target, saved.TargetIP)
	effectiveTarget := formatTargetDisplay(effective.Target, effective.TargetIP)
	if savedTarget != effectiveTarget {
		switch {
		case savedTarget == "not set":
			changes = append(changes, fmt.Sprintf("target -> %s", effectiveTarget))
		case effectiveTarget == "not set":
			changes = append(changes, fmt.Sprintf("target cleared (was %s)", savedTarget))
		default:
			changes = append(changes, fmt.Sprintf("target %s -> %s", savedTarget, effectiveTarget))
		}
	}

	return strings.Join(changes, "; ")
}

func collectShellPreflightWarningsAt(saved, effective config.ContextData, targets []config.Target, now time.Time) []string {
	var warnings []string

	if age := formatContextAgeDetailAt(saved.Timestamp, now); strings.HasPrefix(age, "stale") {
		warnings = append(warnings, fmt.Sprintf("Saved context is %s.", strings.TrimPrefix(age, "stale ")))
	}

	if len(targets) > 1 && strings.TrimSpace(effective.Target) == "" {
		warnings = append(warnings, "Multiple targets are configured, but no active target is selected.")
	}

	if pending := describePendingContextMutation(saved, effective); pending != "" {
		warnings = append(warnings, fmt.Sprintf("Shell context differs from the saved context: %s.", pending))
	}

	return warnings
}

func formatContextAgeDetail(timestamp string) string {
	return formatContextAgeDetailAt(timestamp, time.Now())
}

func formatContextAgeDetailAt(timestamp string, now time.Time) string {
	if strings.TrimSpace(timestamp) == "" {
		return "unknown"
	}

	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "unknown"
	}

	age := now.Sub(ts)
	if age < 0 {
		age = 0
	}

	label := fmt.Sprintf("%s old (saved %s)", formatAgeDuration(age), ts.Local().Format(contextDisplayTimeLayout))
	if age > shellContextStaleThreshold {
		return "stale " + label
	}
	return label
}

func formatRelativeTimeAt(timestamp string, now time.Time) string {
	ts, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return "unknown time"
	}

	if sameDay(now, ts) {
		return utils.FormatRelativeTime(ts)
	}

	return ts.Local().Format(contextDisplayTimeLayout)
}

func formatAgeDuration(age time.Duration) string {
	switch {
	case age < time.Minute:
		return "less than 1 minute"
	case age < time.Hour:
		return fmt.Sprintf("%dm", int(age.Round(time.Minute).Minutes()))
	case age < 24*time.Hour:
		hours := int(age / time.Hour)
		minutes := int(age.Round(time.Minute)/time.Minute) % 60
		if minutes == 0 {
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		days := int(age / (24 * time.Hour))
		hours := int(age/time.Hour) % 24
		if hours == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}

func formatTargetDisplay(target, ip string) string {
	target = strings.TrimSpace(target)
	ip = strings.TrimSpace(ip)

	switch {
	case target != "" && ip != "":
		return fmt.Sprintf("%s (%s)", target, ip)
	case target != "":
		return target
	case ip != "":
		return ip
	default:
		return "not set"
	}
}

func formatValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "not set"
	}
	return value
}

func equalTrimmed(left, right string) bool {
	return strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right))
}

func sameDay(left, right time.Time) bool {
	ly, lm, ld := left.Date()
	ry, rm, rd := right.Date()
	return ly == ry && lm == rm && ld == rd
}
