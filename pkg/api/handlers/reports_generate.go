package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"pentlog/pkg/httpauth"
	"pentlog/pkg/logs"
	"pentlog/pkg/recorder"
	"pentlog/pkg/utils"
	"pentlog/pkg/vulns"

	"github.com/go-chi/chi/v5"
)

type reportGenerateRequest struct {
	Client        string `json:"client"`
	Engagement    string `json:"engagement"`
	Phase         string `json:"phase"`
	Format        string `json:"format"`
	IncludeGIFs   bool   `json:"include_gifs"`
	GIFResolution string `json:"gif_resolution"`
	OutputName    string `json:"output_name"`
}

type reportGenerateJob struct {
	ID                    string `json:"id"`
	Status                string `json:"status"`
	Message               string `json:"message"`
	Error                 string `json:"error,omitempty"`
	Client                string `json:"client"`
	Engagement            string `json:"engagement,omitempty"`
	Phase                 string `json:"phase,omitempty"`
	Format                string `json:"format"`
	IncludeGIFs           bool   `json:"include_gifs"`
	GIFResolution         string `json:"gif_resolution,omitempty"`
	OutputName            string `json:"output_name"`
	ReportPath            string `json:"report_path,omitempty"`
	RelativePath          string `json:"relative_path,omitempty"`
	ViewURL               string `json:"view_url,omitempty"`
	SessionsCount         int    `json:"sessions_count"`
	GIFGenerated          int    `json:"gif_generated"`
	GIFFailed             int    `json:"gif_failed"`
	AvgTimePerSessionSecs int    `json:"avg_time_per_session_secs,omitempty"`
	EstTimeRemainingSecs  int    `json:"est_time_remaining_secs,omitempty"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}

var reportJobSeq uint64

var reportJobs = struct {
	mu   sync.RWMutex
	jobs map[string]reportGenerateJob
}{
	jobs: map[string]reportGenerateJob{},
}

func handleReportsGenerate(w http.ResponseWriter, r *http.Request) {
	var req reportGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	req.Client = strings.TrimSpace(req.Client)
	req.Engagement = strings.TrimSpace(req.Engagement)
	req.Phase = strings.TrimSpace(req.Phase)
	req.Format = strings.ToLower(strings.TrimSpace(req.Format))
	req.GIFResolution = strings.ToLower(strings.TrimSpace(req.GIFResolution))
	req.OutputName = strings.TrimSpace(req.OutputName)

	if req.Format == "" {
		req.Format = "html"
	}
	if req.Format != "html" && req.Format != "md" {
		http.Error(w, `{"error":"format must be 'html' or 'md'"}`, http.StatusBadRequest)
		return
	}
	if req.IncludeGIFs && req.Format != "html" {
		http.Error(w, `{"error":"GIF generation is only supported for HTML reports"}`, http.StatusBadRequest)
		return
	}

	if req.GIFResolution == "" {
		req.GIFResolution = "720p"
	}
	if req.GIFResolution != "720p" && req.GIFResolution != "1080p" {
		http.Error(w, `{"error":"gif_resolution must be '720p' or '1080p'"}`, http.StatusBadRequest)
		return
	}

	if req.Client == "" {
		ctx, err := configManager().LoadContext()
		if err != nil || strings.TrimSpace(ctx.Client) == "" {
			http.Error(w, `{"error":"client is required when no active context exists"}`, http.StatusBadRequest)
			return
		}
		req.Client = strings.TrimSpace(ctx.Client)
		if req.Engagement == "" {
			req.Engagement = strings.TrimSpace(ctx.Engagement)
		}
		if req.Phase == "" {
			req.Phase = strings.TrimSpace(ctx.Phase)
		}
	}

	sessions, err := logs.ListSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	filtered := filterSessionsForReport(sessions, req.Client, req.Engagement, req.Phase)
	if len(filtered) == 0 {
		http.Error(w, `{"error":"No sessions found for selected filters"}`, http.StatusBadRequest)
		return
	}

	if req.OutputName != "" {
		req.OutputName = filepath.Base(req.OutputName)
		if req.OutputName == "." || req.OutputName == "" {
			req.OutputName = ""
		}
	}

	jobID := fmt.Sprintf("report-%d-%d", time.Now().Unix(), atomic.AddUint64(&reportJobSeq, 1))
	now := time.Now().Format(time.RFC3339)

	job := reportGenerateJob{
		ID:            jobID,
		Status:        "queued",
		Message:       "Queued report generation",
		Client:        req.Client,
		Engagement:    req.Engagement,
		Phase:         req.Phase,
		Format:        req.Format,
		IncludeGIFs:   req.IncludeGIFs,
		GIFResolution: req.GIFResolution,
		OutputName:    req.OutputName,
		SessionsCount: len(filtered),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	setReportJob(job)

	go runReportJob(jobID, req, filtered)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"job": job,
	})
}

func handleReportsJobByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, ok := getReportJob(id)
	if !ok {
		http.Error(w, `{"error":"Job not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"job": job,
	})
}

func handleReportsActiveJob(w http.ResponseWriter, r *http.Request) {
	reportJobs.mu.RLock()
	defer reportJobs.mu.RUnlock()

	var activeJob *reportGenerateJob
	for _, job := range reportJobs.jobs {
		if job.Status == "queued" || job.Status == "running" {
			if activeJob == nil || job.CreatedAt > activeJob.CreatedAt {
				activeJob = &job
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if activeJob != nil {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"job": activeJob,
		})
	} else {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"job": nil,
		})
	}
}

func filterSessionsForReport(sessions []logs.Session, client, engagement, phase string) []logs.Session {
	filtered := make([]logs.Session, 0, len(sessions))
	for _, session := range sessions {
		if client != "" && !strings.EqualFold(session.Metadata.Client, client) {
			continue
		}
		if engagement != "" && !strings.EqualFold(session.Metadata.Engagement, engagement) {
			continue
		}
		if phase != "" && !strings.EqualFold(session.Metadata.Phase, phase) {
			continue
		}
		filtered = append(filtered, session)
	}
	return filtered
}

func runReportJob(id string, req reportGenerateRequest, sessions []logs.Session) {
	updateReportJob(id, func(job *reportGenerateJob) {
		job.Status = "running"
		job.Message = "Generating report"
		job.OutputName = resolveReportFilename(req)
	})

	mgr := configManager()
	reportDir := filepath.Join(mgr.GetPaths().ReportsDir, utils.Slugify(req.Client))
	if err := os.MkdirAll(reportDir, 0700); err != nil {
		failReportJob(id, fmt.Errorf("failed to prepare report directory: %w", err))
		return
	}

	filename := resolveReportFilename(req)
	fullPath := filepath.Join(reportDir, filename)

	if req.Format == "md" {
		markdownReport, err := logs.GenerateReport(sessions, req.Client)
		if err != nil {
			failReportJob(id, err)
			return
		}
		if err := utils.WritePrivateFile(fullPath, []byte(markdownReport)); err != nil {
			failReportJob(id, fmt.Errorf("failed to write markdown report: %w", err))
			return
		}

		completeReportJob(id, fullPath, req.Client, filename, "Markdown report generated")
		return
	}

	var gifPaths map[int]string
	gifGenerated := 0
	gifFailed := 0
	var totalTimeGif time.Duration

	if req.IncludeGIFs {
		gifPaths = make(map[int]string)
		gifsDir := filepath.Join(reportDir, "gifs")
		if err := os.MkdirAll(gifsDir, 0700); err != nil {
			failReportJob(id, fmt.Errorf("failed to create gifs directory: %w", err))
			return
		}

		for _, session := range sessions {
			if session.Path == "" {
				continue
			}

			sessionStart := time.Now()

			updateReportJob(id, func(job *reportGenerateJob) {
				job.Message = fmt.Sprintf("Generating GIF for session %d (%d/%d)", session.ID, gifGenerated+gifFailed+1, len(sessions))
			})

			gifName := fmt.Sprintf("session-%d.gif", session.ID)
			gifPath := filepath.Join(gifsDir, gifName)
			relativePath := filepath.ToSlash(filepath.Join("gifs", gifName))

			cfg := recorder.DefaultConfig()
			cfg.Speed = 2.0
			cfg.Resolution = req.GIFResolution
			if req.GIFResolution == "1080p" {
				cfg.Cols = 274
				cfg.Rows = 83
			} else {
				cfg.Cols = 183
				cfg.Rows = 55
			}

			if err := recorder.RenderToGIF(session.Path, gifPath, cfg); err != nil {
				gifFailed++
				updateReportJob(id, func(job *reportGenerateJob) {
					job.Message = fmt.Sprintf("GIF failed for session %d: %s", session.ID, err.Error())
					job.GIFGenerated = gifGenerated
					job.GIFFailed = gifFailed
				})
				continue
			}

			sessionDuration := time.Since(sessionStart)
			totalTimeGif += sessionDuration
			gifGenerated++
			gifPaths[session.ID] = relativePath

			processed := gifGenerated + gifFailed
			remaining := len(sessions) - processed
			var avgSecs, estRemainingSecs int
			if processed > 0 {
				avgSecs = int(totalTimeGif.Seconds() / float64(processed))
				estRemainingSecs = avgSecs * remaining
			}

			updateReportJob(id, func(job *reportGenerateJob) {
				job.Message = fmt.Sprintf("Generating GIFs (%d/%d)", processed, len(sessions))
				job.GIFGenerated = gifGenerated
				job.GIFFailed = gifFailed
				job.AvgTimePerSessionSecs = avgSecs
				job.EstTimeRemainingSecs = estRemainingSecs
			})
		}
	}

	findings := collectReportFindings(req.Client, req.Engagement, req.Phase)
	htmlReport, err := logs.GenerateHTMLReport(sessions, req.Client, findings, "", gifPaths)
	if err != nil {
		failReportJob(id, err)
		return
	}

	if err := utils.WritePrivateFile(fullPath, []byte(htmlReport)); err != nil {
		failReportJob(id, fmt.Errorf("failed to write html report: %w", err))
		return
	}

	updateReportJob(id, func(job *reportGenerateJob) {
		job.GIFGenerated = gifGenerated
		job.GIFFailed = gifFailed
	})
	completeReportJob(id, fullPath, req.Client, filename, "HTML report generated")
}

func resolveReportFilename(req reportGenerateRequest) string {
	ext := "." + req.Format
	if req.Format == "md" {
		ext = ".md"
	}

	if req.OutputName != "" {
		name := req.OutputName
		if !strings.HasSuffix(strings.ToLower(name), ext) {
			name += ext
		}
		return name
	}

	fileNamePhase := req.Phase
	if fileNamePhase == "" {
		fileNamePhase = "all-phases"
	}
	fileNameEng := req.Engagement
	if fileNameEng == "" {
		fileNameEng = "all-engagements"
	}

	return fmt.Sprintf("%s_%s_%s_report%s",
		utils.Slugify(req.Client),
		utils.Slugify(fileNameEng),
		utils.Slugify(fileNamePhase),
		ext,
	)
}

func collectReportFindings(client, engagement, phase string) []vulns.Vuln {
	manager := vulns.NewManager(client, engagement)
	findings, err := manager.List()
	if err != nil || len(findings) == 0 {
		return nil
	}

	if phase == "" {
		return findings
	}

	filtered := make([]vulns.Vuln, 0, len(findings))
	for _, finding := range findings {
		if strings.EqualFold(strings.TrimSpace(finding.Phase), strings.TrimSpace(phase)) {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

func completeReportJob(id, fullPath, client, filename, message string) {
	relativePath := filepath.ToSlash(filepath.Join(utils.Slugify(client), filename))
	viewURL := httpauth.BuildArtifactURL("/files/reports", utils.Slugify(client), filename)
	updateReportJob(id, func(job *reportGenerateJob) {
		job.Status = "completed"
		job.Message = message
		job.ReportPath = fullPath
		job.RelativePath = relativePath
		job.ViewURL = viewURL
	})
}

func failReportJob(id string, err error) {
	updateReportJob(id, func(job *reportGenerateJob) {
		job.Status = "failed"
		job.Message = "Report generation failed"
		job.Error = err.Error()
	})
}

func setReportJob(job reportGenerateJob) {
	reportJobs.mu.Lock()
	defer reportJobs.mu.Unlock()
	reportJobs.jobs[job.ID] = job
}

func getReportJob(id string) (reportGenerateJob, bool) {
	reportJobs.mu.RLock()
	defer reportJobs.mu.RUnlock()
	job, ok := reportJobs.jobs[id]
	return job, ok
}

func updateReportJob(id string, update func(*reportGenerateJob)) {
	reportJobs.mu.Lock()
	defer reportJobs.mu.Unlock()

	job, ok := reportJobs.jobs[id]
	if !ok {
		return
	}

	update(&job)
	job.UpdatedAt = time.Now().Format(time.RFC3339)
	reportJobs.jobs[id] = job
}
