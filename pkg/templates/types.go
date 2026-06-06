package templates

import (
	"html/template"
	"pentlog/pkg/vulns"
)

type ReportSummaryTemplateData struct {
	SessionCount    int
	EngagementCount int
	PhaseCount      int
	FindingCount    int
	EvidenceCount   int
	CommandCount    int
	ArchiveRefCount int
	FirstObserved   string
	LastObserved    string
	CriticalCount   int
	HighCount       int
	MediumCount     int
	LowCount        int
	InfoCount       int
	OpenCount       int
	VerifiedCount   int
	ClosedCount     int
	Targets         []string
	Tags            []string
	Operators       []string
}

type PhaseMetadataTemplateData struct {
	Name         string
	SessionCount int
	Targets      []string
	Tags         []string
}

type EngagementMetadataTemplateData struct {
	Name         string
	DateRange    string
	SessionCount int
	Operators    []string
	Targets      []string
	Tags         []string
	Phases       []PhaseMetadataTemplateData
}

type EvidenceSnippetTemplateData struct {
	Title      string
	Source     string
	Summary    string
	Snippet    string
	Timestamp  string
	Engagement string
	Phase      string
	Target     string
	Tags       []string
	SessionID  int
	FindingID  string
}

type CommandTemplateData struct {
	Timestamp string
	Command   string
	Output    string
}

type CommandAppendixTemplateData struct {
	SessionID  int
	Label      string
	Timestamp  string
	Engagement string
	Phase      string
	Target     string
	Tags       []string
	GIFPath    string
	Commands   []CommandTemplateData
}

type IntegrityReferenceTemplateData struct {
	SessionID      int
	State          string
	Engagement     string
	Phase          string
	Target         string
	TranscriptPath string
	ArchivePath    string
	ManifestSHA256 string
	ArchivedAt     string
}

type TranscriptAppendixTemplateData struct {
	SessionID  int
	Label      string
	Timestamp  string
	Engagement string
	Phase      string
	Target     string
	Tags       []string
	GIFPath    string
	Content    template.HTML
}

type ReportTemplateData struct {
	Client                string
	GeneratedAt           string
	Summary               ReportSummaryTemplateData
	Findings              []vulns.Vuln
	AIAnalysis            template.HTML
	CSS                   template.CSS
	EngagementMetadata    []EngagementMetadataTemplateData
	EvidenceSnippets      []EvidenceSnippetTemplateData
	CommandAppendix       []CommandAppendixTemplateData
	IntegrityReferences   []IntegrityReferenceTemplateData
	HasTranscriptAppendix bool
	TranscriptAppendix    []TranscriptAppendixTemplateData
}
