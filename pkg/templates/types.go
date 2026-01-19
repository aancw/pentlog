package templates

import (
	"html/template"
	"pentlog/pkg/vulns"
)

// Shared Structs for Template Data

type SessionData struct {
	ID      int
	ModTime string
	Content template.HTML // HTML pre-rendered content
}

type PhaseData struct {
	Name     string
	Sessions []SessionData
}

type EngagementData struct {
	Name   string
	Phases []PhaseData
}

type ReportData struct {
	Client      string
	Findings    []vulns.Vuln
	AIAnalysis  template.HTML // HTML content
	CSS         template.CSS
	Engagements []EngagementData
}
