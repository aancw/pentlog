package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"pentlog/pkg/config"
	"pentlog/pkg/dashboard"
	"pentlog/pkg/httpauth"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"pentlog/pkg/vulns"

	"github.com/go-chi/chi/v5"
)

const dashboardRecentLimit = 10

func DashboardRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/overview", handleDashboardOverview)
	r.Get("/stats", handleDashboardStats)
	r.Get("/activity", handleDashboardActivity)
	r.Get("/clients", handleDashboardClients)
	r.Get("/engagements", handleDashboardEngagements)
	r.Get("/phases", handleDashboardPhases)
	return r
}

func handleDashboardOverview(w http.ResponseWriter, r *http.Request) {
	sessions, allVulns, err := loadDashboardSessionsAndVulns()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	noteCount := countSessionNotes(sessions)
	stats := buildStatsResponse(sessions, allVulns, noteCount)
	activity := buildActivityResponse(sessions, allVulns, dashboardRecentLimit)
	clients := buildClientResponse(sessions)

	mgr := config.Manager()
	reportsTotal, latestReport, err := collectArtifactSummary(
		mgr.GetPaths().ReportsDir,
		map[string]bool{".md": true, ".html": true},
		"/files/reports/",
	)
	if err != nil {
		http.Error(w, `{"error":"Failed to read reports directory"}`, http.StatusInternalServerError)
		return
	}

	archivesTotal, latestArchive, err := collectArtifactSummary(
		mgr.GetPaths().ArchiveDir,
		map[string]bool{".zip": true},
		"/files/archives/",
	)
	if err != nil {
		http.Error(w, `{"error":"Failed to read archive directory"}`, http.StatusInternalServerError)
		return
	}

	hasContext, contextResp := loadCurrentContextPayload()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"stats":    stats,
		"activity": activity,
		"clients":  clients,
		"context": map[string]interface{}{
			"has_context": hasContext,
			"context":     contextResp,
		},
		"artifacts": map[string]interface{}{
			"reports_total":  reportsTotal,
			"archives_total": archivesTotal,
			"latest_report":  latestReport,
			"latest_archive": latestArchive,
		},
	})
}

func handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	sessions, allVulns, err := loadDashboardSessionsAndVulns()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	noteCount := countSessionNotes(sessions)
	resp := buildStatsResponse(sessions, allVulns, noteCount)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func handleDashboardActivity(w http.ResponseWriter, r *http.Request) {
	sessions, allVulns, err := loadDashboardSessionsAndVulns()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	resp := buildActivityResponse(sessions, allVulns, dashboardRecentLimit)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func handleDashboardClients(w http.ResponseWriter, r *http.Request) {
	sessions, err := logs.ListSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	clients := buildClientResponse(sessions)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"clients": clients})
}

func handleDashboardEngagements(w http.ResponseWriter, r *http.Request) {
	client := r.URL.Query().Get("client")

	sessions, err := logs.ListSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	engSizes := make(map[string]int64)
	engCounts := make(map[string]int)

	for _, s := range sessions {
		if client != "" && s.Metadata.Client != client {
			continue
		}
		if s.Metadata.Engagement != "" {
			engSizes[s.Metadata.Engagement] += s.Size
			engCounts[s.Metadata.Engagement]++
		}
	}

	var engagements []map[string]interface{}
	for name, size := range engSizes {
		engagements = append(engagements, map[string]interface{}{
			"name":             name,
			"sessions_count":   engCounts[name],
			"total_size":       size,
			"total_size_human": utils.FormatBytes(size),
		})
	}

	sort.Slice(engagements, func(i, j int) bool {
		return engagements[i]["total_size"].(int64) > engagements[j]["total_size"].(int64)
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"engagements": engagements})
}

func handleDashboardPhases(w http.ResponseWriter, r *http.Request) {
	client := r.URL.Query().Get("client")
	engagement := r.URL.Query().Get("engagement")

	sessions, err := logs.ListSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	phaseCounts := make(map[string]int)

	for _, s := range sessions {
		if client != "" && s.Metadata.Client != client {
			continue
		}
		if engagement != "" && s.Metadata.Engagement != engagement {
			continue
		}
		if s.Metadata.Phase != "" {
			phaseCounts[s.Metadata.Phase]++
		}
	}

	var phases []map[string]interface{}
	for name, count := range phaseCounts {
		phases = append(phases, map[string]interface{}{
			"name":           name,
			"sessions_count": count,
		})
	}

	sort.Slice(phases, func(i, j int) bool {
		return phases[i]["sessions_count"].(int) > phases[j]["sessions_count"].(int)
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"phases": phases})
}

func loadDashboardSessionsAndVulns() ([]logs.Session, []vulns.Vuln, error) {
	sessions, err := logs.ListSessions()
	if err != nil {
		return nil, nil, err
	}

	var allVulns []vulns.Vuln
	uniqueContexts := make(map[string]bool)
	for _, s := range sessions {
		if s.Metadata.Client == "" {
			continue
		}

		key := s.Metadata.Client + "|" + s.Metadata.Engagement
		if uniqueContexts[key] {
			continue
		}
		uniqueContexts[key] = true

		mgr := vulns.NewManager(s.Metadata.Client, s.Metadata.Engagement)
		if vList, err := mgr.List(); err == nil {
			allVulns = append(allVulns, vList...)
		}
	}

	return sessions, allVulns, nil
}

func countSessionNotes(sessions []logs.Session) int {
	noteCount := 0
	for _, s := range sessions {
		if s.NotesPath == "" {
			continue
		}
		notes, err := logs.ReadNotes(s.NotesPath)
		if err == nil {
			noteCount += len(notes)
		}
	}
	return noteCount
}

func buildStatsResponse(sessions []logs.Session, allVulns []vulns.Vuln, noteCount int) map[string]interface{} {
	stats := dashboard.Stats{
		TotalSessions:    len(sessions),
		PhaseCounts:      make(map[string]int),
		EngagementCounts: make(map[string]int),
		ClientSizes:      make(map[string]int64),
		EngagementSizes:  make(map[string]int64),
	}

	clients := make(map[string]bool)
	engagements := make(map[string]bool)
	stateCounts := make(map[string]int)

	for _, s := range sessions {
		stats.TotalSize += s.Size
		if s.Metadata.Client != "" {
			clients[s.Metadata.Client] = true
			stats.ClientSizes[s.Metadata.Client] += s.Size
		}
		if s.Metadata.Engagement != "" {
			engagements[s.Metadata.Engagement] = true
			stats.EngagementCounts[s.Metadata.Engagement]++
			stats.EngagementSizes[s.Metadata.Engagement] += s.Size
		}
		if s.Metadata.Phase != "" {
			stats.PhaseCounts[s.Metadata.Phase]++
		}

		state := strings.TrimSpace(string(s.State))
		if state == "" {
			state = string(logs.SessionStateCompleted)
		}
		stateCounts[state]++
	}

	stats.UniqueClients = len(clients)
	stats.UniqueEngagements = len(engagements)
	stats.TotalNotes = noteCount

	severityCounts := make(map[string]int)
	for _, v := range allVulns {
		severityCounts[string(v.Severity)]++
	}

	lastActivity := ""
	if len(sessions) > 0 {
		lastActivity = sessions[0].ModTime
	}

	return map[string]interface{}{
		"total_sessions":     stats.TotalSessions,
		"total_size":         stats.TotalSize,
		"total_size_human":   utils.FormatBytes(stats.TotalSize),
		"unique_clients":     stats.UniqueClients,
		"unique_engagements": stats.UniqueEngagements,
		"total_notes":        stats.TotalNotes,
		"total_vulns":        len(allVulns),
		"phase_counts":       stats.PhaseCounts,
		"severity_counts":    severityCounts,
		"state_counts":       stateCounts,
		"last_activity":      lastActivity,
	}
}

func buildActivityResponse(sessions []logs.Session, allVulns []vulns.Vuln, limit int) map[string]interface{} {
	if sessions == nil {
		return map[string]interface{}{
			"recent_sessions": []interface{}{},
			"recent_vulns":    []interface{}{},
		}
	}

	countSession := limit
	if len(sessions) < limit {
		countSession = len(sessions)
	}
	recent := sessions[:countSession]

	recentSessions := make([]map[string]interface{}, 0, len(recent))
	for _, s := range recent {
		recentSessions = append(recentSessions, sessionToMap(s))
	}

	sort.Slice(allVulns, func(i, j int) bool {
		return allVulns[i].CreatedAt.After(allVulns[j].CreatedAt)
	})

	maxCount := limit
	if len(allVulns) < limit {
		maxCount = len(allVulns)
	}

	recentVulns := make([]map[string]interface{}, 0, maxCount)
	for i := 0; i < maxCount; i++ {
		recentVulns = append(recentVulns, vulnToMap(allVulns[i]))
	}

	return map[string]interface{}{
		"recent_sessions": recentSessions,
		"recent_vulns":    recentVulns,
	}
}

func buildClientResponse(sessions []logs.Session) []map[string]interface{} {
	clientSizes := make(map[string]int64)
	clientCounts := make(map[string]int)

	for _, s := range sessions {
		if s.Metadata.Client != "" {
			clientSizes[s.Metadata.Client] += s.Size
			clientCounts[s.Metadata.Client]++
		}
	}

	clients := make([]map[string]interface{}, 0, len(clientSizes))
	for name, size := range clientSizes {
		clients = append(clients, map[string]interface{}{
			"name":             name,
			"sessions_count":   clientCounts[name],
			"total_size":       size,
			"total_size_human": utils.FormatBytes(size),
		})
	}

	sort.Slice(clients, func(i, j int) bool {
		return clients[i]["total_size"].(int64) > clients[j]["total_size"].(int64)
	})

	return clients
}

func collectArtifactSummary(root string, extensions map[string]bool, fileURLPrefix string) (int, map[string]interface{}, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil, nil
		}
		return 0, nil, err
	}

	total := 0
	hasLatest := false
	var latestModTime time.Time
	var latest map[string]interface{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		clientDir := filepath.Join(root, entry.Name())
		clientEntries, err := os.ReadDir(clientDir)
		if err != nil {
			continue
		}

		for _, ce := range clientEntries {
			if ce.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(ce.Name()))
			if !extensions[ext] {
				continue
			}

			info, err := ce.Info()
			if err != nil {
				continue
			}

			total++
			if hasLatest && !info.ModTime().After(latestModTime) {
				continue
			}

			hasLatest = true
			latestModTime = info.ModTime()
			relativePath := filepath.ToSlash(filepath.Join(entry.Name(), ce.Name()))
			latest = map[string]interface{}{
				"name":          ce.Name(),
				"client":        entry.Name(),
				"relative_path": relativePath,
				"size":          info.Size(),
				"size_human":    utils.FormatBytes(info.Size()),
				"mod_time":      info.ModTime().Format("2006-01-02 15:04:05"),
				"url":           httpauth.BuildArtifactURL(strings.TrimRight(fileURLPrefix, "/"), entry.Name(), ce.Name()),
			}
		}
	}

	if !hasLatest {
		return total, nil, nil
	}

	return total, latest, nil
}

func loadCurrentContextPayload() (bool, map[string]interface{}) {
	mgr := config.Manager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		return false, nil
	}

	return true, map[string]interface{}{
		"client":     ctx.Client,
		"engagement": ctx.Engagement,
		"scope":      ctx.Scope,
		"operator":   ctx.Operator,
		"phase":      ctx.Phase,
		"target":     ctx.Target,
		"target_ip":  ctx.TargetIP,
		"timestamp":  ctx.Timestamp,
		"type":       ctx.Type,
	}
}

func sessionToMap(s logs.Session) map[string]interface{} {
	tags, _ := logs.GetSessionTags(s.ID)
	notesCount := 0
	if s.NotesPath != "" {
		notes, err := logs.ReadNotes(s.NotesPath)
		if err == nil {
			notesCount = len(notes)
		}
	}

	return map[string]interface{}{
		"id":                      s.ID,
		"filename":                s.Filename,
		"path":                    s.Path,
		"display_path":            s.DisplayPath,
		"size":                    s.Size,
		"size_human":              utils.FormatBytes(s.Size),
		"mod_time":                s.ModTime,
		"state":                   string(s.State),
		"last_sync_at":            s.LastSyncAt,
		"recorder_pid":            s.RecorderPID,
		"host_fingerprint":        s.HostFingerprint,
		"hostname":                s.Hostname,
		"started_at":              s.StartedAt,
		"ended_at":                s.EndedAt,
		"resume_count":            s.ResumeCount,
		"archived_at":             s.ArchivedAt,
		"archive_path":            s.ArchivePath,
		"archive_manifest_sha256": s.ArchiveManifestSHA256,
		"metadata": map[string]string{
			"client":     s.Metadata.Client,
			"engagement": s.Metadata.Engagement,
			"scope":      s.Metadata.Scope,
			"operator":   s.Metadata.Operator,
			"phase":      s.Metadata.Phase,
			"target":     s.Metadata.Target,
			"target_ip":  s.Metadata.TargetIP,
		},
		"tags":        tags,
		"notes_count": notesCount,
		"has_gif":     false,
	}
}

func vulnToMap(v vulns.Vuln) map[string]interface{} {
	return map[string]interface{}{
		"id":             v.ID,
		"title":          v.Title,
		"severity":       string(v.Severity),
		"severity_color": v.SeverityColor(),
		"status":         string(v.Status),
		"description":    v.Description,
		"remediation":    v.Remediation,
		"references":     v.References,
		"evidence":       v.Evidence,
		"phase":          v.Phase,
		"created_at":     v.CreatedAt,
		"updated_at":     v.UpdatedAt,
	}
}
