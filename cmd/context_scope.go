package cmd

import (
	"fmt"
	"pentlog/pkg/config"
	"pentlog/pkg/logs"
	"strings"
)

type sessionScope struct {
	Client     string
	Engagement string
	Phase      string
}

func filterSessionsByScope(sessions []logs.Session, scope sessionScope) []logs.Session {
	var filtered []logs.Session
	for _, session := range sessions {
		if scope.Client != "" && session.Metadata.Client != scope.Client {
			continue
		}
		if scope.Engagement != "" && session.Metadata.Engagement != scope.Engagement {
			continue
		}
		if scope.Phase != "" && !strings.EqualFold(strings.TrimSpace(session.Metadata.Phase), strings.TrimSpace(scope.Phase)) {
			continue
		}
		filtered = append(filtered, session)
	}

	return filtered
}

func hasSessionsForScope(sessions []logs.Session, scope sessionScope) bool {
	return len(filterSessionsByScope(sessions, scope)) > 0
}

func currentEngagementScope(ctx *config.ContextData) sessionScope {
	if ctx == nil {
		return sessionScope{}
	}

	return sessionScope{
		Client:     ctx.Client,
		Engagement: ctx.Engagement,
	}
}

func currentPhaseScope(ctx *config.ContextData) sessionScope {
	if ctx == nil {
		return sessionScope{}
	}

	return sessionScope{
		Client:     ctx.Client,
		Engagement: ctx.Engagement,
		Phase:      ctx.Phase,
	}
}

func currentClientScope(ctx *config.ContextData) sessionScope {
	if ctx == nil {
		return sessionScope{}
	}

	return sessionScope{Client: ctx.Client}
}

func formatScopeLabel(ctx *config.ContextData, includePhase bool) string {
	if ctx == nil {
		return ""
	}

	if includePhase {
		return fmt.Sprintf("Current phase (%s / %s / %s)", ctx.Client, ctx.Engagement, ctx.Phase)
	}

	return fmt.Sprintf("Current engagement (%s / %s)", ctx.Client, ctx.Engagement)
}
