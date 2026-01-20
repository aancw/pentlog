package templates

import (
	"html/template"
	"pentlog/pkg/vulns"
)

// Shared Structs for Template Data

type SessionTemplateData struct {
	ID      int
	ModTime string
	Content template.HTML // HTML pre-rendered content
}

type PhaseTemplateData struct {
	Name     string
	Sessions []SessionTemplateData
}

type EngagementTemplateData struct {
	Name   string
	Phases []PhaseTemplateData
}

type ReportTemplateData struct {
	Client      string
	Findings    []vulns.Vuln
	AIAnalysis  template.HTML // HTML content
	CSS         template.CSS
	Engagements []EngagementTemplateData
}
