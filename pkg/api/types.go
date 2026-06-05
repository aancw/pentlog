package api

import "time"

type SessionResponse struct {
	ID          int         `json:"id"`
	Filename    string      `json:"filename"`
	Path        string      `json:"path"`
	DisplayPath string      `json:"display_path"`
	Size        int64       `json:"size"`
	SizeHuman   string      `json:"size_human"`
	ModTime     string      `json:"mod_time"`
	State       string      `json:"state"`
	Duration    string      `json:"duration"`
	Metadata    SessionMeta `json:"metadata"`
	Tags        []string    `json:"tags"`
	NotesCount  int         `json:"notes_count"`
	HasGIF      bool        `json:"has_gif"`
}

type SessionMeta struct {
	Client     string `json:"client"`
	Engagement string `json:"engagement"`
	Scope      string `json:"scope"`
	Operator   string `json:"operator"`
	Phase      string `json:"phase"`
	Target     string `json:"target"`
	TargetIP   string `json:"target_ip"`
}

type SessionsListResponse struct {
	Sessions []SessionResponse `json:"sessions"`
	Total    int               `json:"total"`
	Page     int               `json:"page"`
	HasMore  bool              `json:"has_more"`
}

type TimelineResponse struct {
	SessionID int            `json:"session_id"`
	Commands  []CommandEntry `json:"commands"`
}

type CommandEntry struct {
	Timestamp string `json:"timestamp"`
	Command   string `json:"command"`
	Output    string `json:"output"`
}

type NoteResponse struct {
	Timestamp  string `json:"timestamp"`
	Content    string `json:"content"`
	ByteOffset int64  `json:"byte_offset"`
}

type VulnResponse struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Severity      string    `json:"severity"`
	SeverityColor string    `json:"severity_color"`
	Status        string    `json:"status"`
	Description   string    `json:"description"`
	Remediation   string    `json:"remediation"`
	References    []string  `json:"references"`
	Evidence      []string  `json:"evidence"`
	Phase         string    `json:"phase"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type VulnListResponse struct {
	Vulns      []VulnResponse `json:"vulns"`
	Total      int            `json:"total"`
	ByStatus   map[string]int `json:"by_status"`
	BySeverity map[string]int `json:"by_severity"`
}

type VulnCreateRequest struct {
	Title       string   `json:"title"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Remediation string   `json:"remediation"`
	References  []string `json:"references"`
	Evidence    []string `json:"evidence"`
	Phase       string   `json:"phase"`
}

type DashboardStatsResponse struct {
	TotalSessions     int            `json:"total_sessions"`
	TotalSize         int64          `json:"total_size"`
	TotalSizeHuman    string         `json:"total_size_human"`
	UniqueClients     int            `json:"unique_clients"`
	UniqueEngagements int            `json:"unique_engagements"`
	TotalNotes        int            `json:"total_notes"`
	TotalVulns        int            `json:"total_vulns"`
	PhaseCounts       map[string]int `json:"phase_counts"`
	SeverityCounts    map[string]int `json:"severity_counts"`
}

type ActivityResponse struct {
	RecentSessions []SessionResponse `json:"recent_sessions"`
	RecentVulns    []VulnResponse    `json:"recent_vulns"`
}

type ClientStatsResponse struct {
	Clients []ClientStats `json:"clients"`
}

type ClientStats struct {
	Name           string `json:"name"`
	SessionsCount  int    `json:"sessions_count"`
	TotalSize      int64  `json:"total_size"`
	TotalSizeHuman string `json:"total_size_human"`
}

type ContextResponse struct {
	Client     string `json:"client"`
	Engagement string `json:"engagement"`
	Scope      string `json:"scope"`
	Operator   string `json:"operator"`
	Phase      string `json:"phase"`
	Target     string `json:"target"`
	TargetIP   string `json:"target_ip"`
	Timestamp  string `json:"timestamp"`
	Type       string `json:"type"`
}

type ContextCreateRequest struct {
	Type       string `json:"type"`
	Client     string `json:"client"`
	Engagement string `json:"engagement"`
	Scope      string `json:"scope"`
	Operator   string `json:"operator"`
	Phase      string `json:"phase"`
	Target     string `json:"target,omitempty"`
	TargetIP   string `json:"target_ip,omitempty"`
}

type TargetResponse struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	IsCurrent bool   `json:"is_current"`
}

type TargetsListResponse struct {
	Targets []TargetResponse `json:"targets"`
	Current *TargetResponse  `json:"current"`
}

type TargetCreateRequest struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

type SearchResponse struct {
	Results      []SearchResult `json:"results"`
	TotalMatches int            `json:"total_matches"`
	Query        string         `json:"query"`
	IsRegex      bool           `json:"is_regex"`
	Limit        int            `json:"limit"`
	Offset       int            `json:"offset"`
	HasMore      bool           `json:"has_more"`
}

type SearchResult struct {
	SessionID        int      `json:"session_id"`
	SessionPath      string   `json:"session_path"`
	LineNum          int      `json:"line_num"`
	Content          string   `json:"content"`
	Context          []string `json:"context"`
	ContextStartLine int      `json:"context_start_line"`
	IsNote           bool     `json:"is_note"`
	NoteTimestamp    string   `json:"note_timestamp,omitempty"`
}

type Match struct {
	LineNum   int    `json:"line_num"`
	Content   string `json:"content"`
	Highlight string `json:"highlight"`
}

type ArchiveResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	SizeHuman string    `json:"size_human"`
	Encrypted bool      `json:"encrypted"`
	CreatedAt time.Time `json:"created_at"`
}

type ArchiveCreateRequest struct {
	Client         string `json:"client"`
	Engagement     string `json:"engagement,omitempty"`
	Phase          string `json:"phase,omitempty"`
	Password       string `json:"password,omitempty"`
	IncludeReports bool   `json:"include_reports"`
}

type ImportPreviewResponse struct {
	ArchiveType   string   `json:"archive_type"`
	Files         []string `json:"files"`
	SessionsCount int      `json:"sessions_count"`
	Structure     struct {
		Clients     []string `json:"clients"`
		Engagements []string `json:"engagements"`
		Phases      []string `json:"phases"`
	} `json:"structure"`
}

type ImportResultResponse struct {
	TotalFiles       int               `json:"total_files"`
	ImportedFiles    int               `json:"imported_files"`
	SkippedFiles     int               `json:"skipped_files"`
	ImportedSessions []SessionResponse `json:"imported_sessions"`
	Errors           []string          `json:"errors"`
}

type RecoveryStatusResponse struct {
	Crashed  []SessionResponse `json:"crashed"`
	Active   []SessionResponse `json:"active"`
	Orphaned []SessionResponse `json:"orphaned"`
}

type RecoveryResultResponse struct {
	MarkedCount    int    `json:"marked_count"`
	RecoveredCount int    `json:"recovered_count"`
	FailedCount    int    `json:"failed_count"`
	DeletedCount   int    `json:"deleted_count"`
	Message        string `json:"message"`
}

type ReportResponse struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	SizeHuman  string    `json:"size_human"`
	CreatedAt  time.Time `json:"created_at"`
	Client     string    `json:"client"`
	Engagement string    `json:"engagement"`
	Phase      string    `json:"phase"`
}

type ReportGenerateRequest struct {
	Client      string `json:"client"`
	Engagement  string `json:"engagement,omitempty"`
	Phase       string `json:"phase,omitempty"`
	Format      string `json:"format"`
	IncludeGIFs bool   `json:"include_gifs"`
}

type AIAnalysisResponse struct {
	Analysis    string `json:"analysis"`
	Provider    string `json:"provider"`
	Model       string `json:"model"`
	GeneratedAt string `json:"generated_at"`
}

type HashResponse struct {
	Filename string `json:"filename"`
	SHA256   string `json:"sha256"`
}

type HashesResponse struct {
	Hashes      []HashResponse `json:"hashes"`
	GeneratedAt string         `json:"generated_at"`
	HashesFile  string         `json:"hashes_file"`
}

type VerifyResponse struct {
	Verified   bool           `json:"verified"`
	Mismatches []HashMismatch `json:"mismatches"`
	CheckedAt  string         `json:"checked_at"`
}

type HashMismatch struct {
	Filename string `json:"filename"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

type SystemStatusResponse struct {
	HasContext    bool             `json:"has_context"`
	Context       *ContextResponse `json:"context"`
	Version       string           `json:"version"`
	DBPath        string           `json:"db_path"`
	TotalSessions int              `json:"total_sessions"`
}

type SystemInfoResponse struct {
	Version string    `json:"version"`
	Paths   PathsInfo `json:"paths"`
	Uptime  string    `json:"uptime"`
}

type PathsInfo struct {
	Home         string `json:"home"`
	LogsDir      string `json:"logs_dir"`
	ReportsDir   string `json:"reports_dir"`
	ArchiveDir   string `json:"archive_dir"`
	DatabaseFile string `json:"database_file"`
}

type ShareCreateResponse struct {
	Token     string    `json:"token"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ShareStatusResponse struct {
	Active      bool         `json:"active"`
	Token       string       `json:"token"`
	ClientCount int          `json:"client_count"`
	Clients     []ClientInfo `json:"clients"`
}

type ClientInfo struct {
	IP          string `json:"ip"`
	ConnectedAt string `json:"connected_at"`
}

type AIConfigResponse struct {
	Provider string       `json:"provider"`
	Gemini   GeminiConfig `json:"gemini"`
	Ollama   OllamaConfig `json:"ollama"`
}

type GeminiConfig struct {
	HasKey bool `json:"has_key"`
}

type OllamaConfig struct {
	Model string `json:"model"`
	URL   string `json:"url"`
}

type AIConfigUpdateRequest struct {
	Provider     string `json:"provider,omitempty"`
	GeminiAPIKey string `json:"gemini_api_key,omitempty"`
	OllamaModel  string `json:"ollama_model,omitempty"`
	OllamaURL    string `json:"ollama_url,omitempty"`
}

type TagResponse struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type TagsListResponse struct {
	Tags []TagResponse `json:"tags"`
}

type GIFGenerateRequest struct {
	Speed      float64 `json:"speed"`
	Cols       int     `json:"cols"`
	Rows       int     `json:"rows"`
	Resolution string  `json:"resolution"`
}

type GIFResponse struct {
	Path      string `json:"path"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    int    `json:"code"`
	Details string `json:"details,omitempty"`
}
