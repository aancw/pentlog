package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"pentlog/pkg/config"
	"pentlog/pkg/dashboard"
	"pentlog/pkg/logs"
	"pentlog/pkg/utils"
	"pentlog/pkg/vulns"

	"github.com/go-chi/chi/v5"
)

func DashboardRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/stats", handleDashboardStats)
	r.Get("/activity", handleDashboardActivity)
	r.Get("/clients", handleDashboardClients)
	r.Get("/engagements", handleDashboardEngagements)
	return r
}

func handleDashboardStats(w http.ResponseWriter, r *http.Request) {
	sessions, err := logs.ListSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	stats := dashboard.Stats{
		TotalSessions:    len(sessions),
		PhaseCounts:      make(map[string]int),
		EngagementCounts: make(map[string]int),
		ClientSizes:      make(map[string]int64),
		EngagementSizes:  make(map[string]int64),
	}

	clients := make(map[string]bool)
	engagements := make(map[string]bool)
	noteCount := 0

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
		if s.NotesPath != "" {
			notes, err := logs.ReadNotes(s.NotesPath)
			if err == nil {
				noteCount += len(notes)
			}
		}
	}

	stats.UniqueClients = len(clients)
	stats.UniqueEngagements = len(engagements)
	stats.TotalNotes = noteCount

	var allVulns []vulns.Vuln
	uniqueContexts := make(map[string]bool)
	for _, s := range sessions {
		key := s.Metadata.Client + "|" + s.Metadata.Engagement
		if !uniqueContexts[key] && s.Metadata.Client != "" {
			uniqueContexts[key] = true
			mgr := vulns.NewManager(s.Metadata.Client, s.Metadata.Engagement)
			if vList, err := mgr.List(); err == nil {
				allVulns = append(allVulns, vList...)
			}
		}
	}

	severityCounts := make(map[string]int)
	for _, v := range allVulns {
		severityCounts[string(v.Severity)]++
	}

	resp := map[string]interface{}{
		"total_sessions":     stats.TotalSessions,
		"total_size":         stats.TotalSize,
		"total_size_human":   utils.FormatBytes(stats.TotalSize),
		"unique_clients":     stats.UniqueClients,
		"unique_engagements": stats.UniqueEngagements,
		"total_notes":        stats.TotalNotes,
		"total_vulns":        len(allVulns),
		"phase_counts":       stats.PhaseCounts,
		"severity_counts":    severityCounts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleDashboardActivity(w http.ResponseWriter, r *http.Request) {
	sessions, err := logs.ListSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	limit := 10
	if sessions == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"recent_sessions": []interface{}{},
			"recent_vulns":    []interface{}{},
		})
		return
	}

	reversedSessions := make([]logs.Session, len(sessions))
	for i, j := 0, len(sessions)-1; i < len(sessions); i, j = i+1, j-1 {
		reversedSessions[i] = sessions[j]
	}

	countSession := limit
	if len(reversedSessions) < limit {
		countSession = len(reversedSessions)
	}
	recent := reversedSessions[:countSession]

	var recentSessions []map[string]interface{}
	for _, s := range recent {
		recentSessions = append(recentSessions, sessionToMap(s))
	}

	var allVulns []vulns.Vuln
	uniqueContexts := make(map[string]bool)
	for _, s := range sessions {
		key := s.Metadata.Client + "|" + s.Metadata.Engagement
		if !uniqueContexts[key] && s.Metadata.Client != "" {
			uniqueContexts[key] = true
			mgr := vulns.NewManager(s.Metadata.Client, s.Metadata.Engagement)
			if vList, err := mgr.List(); err == nil {
				allVulns = append(allVulns, vList...)
			}
		}
	}

	sort.Slice(allVulns, func(i, j int) bool {
		return allVulns[i].CreatedAt.After(allVulns[j].CreatedAt)
	})

	maxCount := limit
	if len(allVulns) < limit {
		maxCount = len(allVulns)
	}

	var recentVulns []map[string]interface{}
	for i := 0; i < maxCount; i++ {
		recentVulns = append(recentVulns, vulnToMap(allVulns[i]))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recent_sessions": recentSessions,
		"recent_vulns":    recentVulns,
	})
}

func handleDashboardClients(w http.ResponseWriter, r *http.Request) {
	sessions, err := logs.ListSessions()
	if err != nil {
		http.Error(w, `{"error":"Failed to list sessions"}`, http.StatusInternalServerError)
		return
	}

	clientSizes := make(map[string]int64)
	clientCounts := make(map[string]int)

	for _, s := range sessions {
		if s.Metadata.Client != "" {
			clientSizes[s.Metadata.Client] += s.Size
			clientCounts[s.Metadata.Client]++
		}
	}

	var clients []map[string]interface{}
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"clients": clients})
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
	json.NewEncoder(w).Encode(map[string]interface{}{"clients": engagements})
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
		"id":           s.ID,
		"filename":     s.Filename,
		"path":         s.Path,
		"display_path": s.DisplayPath,
		"size":         s.Size,
		"size_human":   utils.FormatBytes(s.Size),
		"mod_time":     s.ModTime,
		"state":        string(s.State),
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

func getContext() (*config.ContextData, error) {
	mgr := config.Manager()
	return mgr.LoadContext()
}
