package cmd

import (
	"os"
	"pentlog/pkg/config"
	"pentlog/pkg/dashboard"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"sort"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	timelineClient     string
	timelineEngagement string
	timelineAll        bool
)

var dashboardTimelineCmd = &cobra.Command{
	Use:   "timeline",
	Short: "Engagement timeline view (sessions, phases, notes, vulns)",
	Run: func(cmd *cobra.Command, args []string) {
		filter := dashboard.TimelineFilter{
			Client:     timelineClient,
			Engagement: timelineEngagement,
			All:        timelineAll,
		}

		if shouldPromptForTimelineScope(filter) {
			applyTimelineScopePrompt(&filter)
		}

		if !filter.All && filter.Client == "" {
			mgr := config.Manager()
			ctx, err := mgr.LoadContext()
			if err == nil {
				filter.Client = ctx.Client
				filter.Engagement = ctx.Engagement
			} else {
				filter.All = true
			}
		}

		p := tea.NewProgram(dashboard.InitialTimelineModel(filter), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			errors.FromError(errors.Generic, "Error running timeline dashboard", err).Fatal()
		}
	},
}

func init() {
	dashboardTimelineCmd.Flags().StringVar(&timelineClient, "client", "", "Filter by client")
	dashboardTimelineCmd.Flags().StringVar(&timelineEngagement, "engagement", "", "Filter by engagement")
	dashboardTimelineCmd.Flags().BoolVar(&timelineAll, "all", false, "Show all engagements")
	dashboardCmd.AddCommand(dashboardTimelineCmd)
}

func shouldPromptForTimelineScope(filter dashboard.TimelineFilter) bool {
	if filter.All || filter.Client != "" || filter.Engagement != "" {
		return false
	}
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func applyTimelineScopePrompt(filter *dashboard.TimelineFilter) {
	sessions, err := logs.ListSessions()
	if err != nil || len(sessions) == 0 {
		return
	}

	var hasContext bool
	if ctx, err := config.Manager().LoadContext(); err == nil && ctx.Client != "" {
		hasContext = true
	}

	options := []string{}
	if hasContext {
		options = append(options, "Current context")
	}
	options = append(options, "All engagements", "Select client", "Select client + engagement")

	choice := utils.PromptSelect("Timeline scope", options)
	switch choice {
	case "All engagements":
		filter.All = true
		return
	case "Select client":
		client := promptClient(sessions)
		if client == "" {
			return
		}
		filter.Client = client
		filter.Engagement = ""
		return
	case "Select client + engagement":
		client := promptClient(sessions)
		if client == "" {
			return
		}
		eng := promptEngagement(sessions, client)
		if eng == "" {
			filter.Client = client
			filter.Engagement = ""
			return
		}
		filter.Client = client
		filter.Engagement = eng
		return
	default:
		return
	}
}

func promptClient(sessions []logs.Session) string {
	clientSet := make(map[string]struct{})
	for _, s := range sessions {
		if s.Metadata.Client != "" {
			clientSet[s.Metadata.Client] = struct{}{}
		}
	}
	if len(clientSet) == 0 {
		return ""
	}
	clients := make([]string, 0, len(clientSet))
	for c := range clientSet {
		clients = append(clients, c)
	}
	sort.Strings(clients)
	return utils.PromptSelect("Select client", clients)
}

func promptEngagement(sessions []logs.Session, client string) string {
	engSet := make(map[string]struct{})
	for _, s := range sessions {
		if s.Metadata.Client == client && s.Metadata.Engagement != "" {
			engSet[s.Metadata.Engagement] = struct{}{}
		}
	}
	if len(engSet) == 0 {
		return ""
	}
	engagements := make([]string, 0, len(engSet))
	for e := range engSet {
		engagements = append(engagements, e)
	}
	sort.Strings(engagements)
	engagements = append([]string{"All engagements for client"}, engagements...)
	choice := utils.PromptSelect("Select engagement", engagements)
	if choice == "All engagements for client" {
		return ""
	}
	return choice
}
