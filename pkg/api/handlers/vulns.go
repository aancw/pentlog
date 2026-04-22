package handlers

import (
	"encoding/json"
	"net/http"

	"pentlog/pkg/vulns"

	"github.com/go-chi/chi/v5"
)

func VulnRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handleVulnsList)
	r.Post("/", handleVulnCreate)
	r.Get("/{id}", handleVulnGet)
	r.Put("/{id}", handleVulnUpdate)
	r.Delete("/{id}", handleVulnDelete)
	return r
}

func handleVulnsList(w http.ResponseWriter, r *http.Request) {
	client := r.URL.Query().Get("client")
	engagement := r.URL.Query().Get("engagement")
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")

	if client == "" {
		mgr := configManager()
		ctx, err := mgr.LoadContext()
		if err != nil {
			http.Error(w, `{"error":"No context. Specify client parameter."}`, http.StatusBadRequest)
			return
		}
		client = ctx.Client
		engagement = ctx.Engagement
	}

	mgr := vulns.NewManager(client, engagement)
	vlist, err := mgr.List()
	if err != nil {
		http.Error(w, `{"error":"Failed to list vulnerabilities"}`, http.StatusInternalServerError)
		return
	}

	if severity != "" {
		filtered := []vulns.Vuln{}
		for _, v := range vlist {
			if string(v.Severity) == severity {
				filtered = append(filtered, v)
			}
		}
		vlist = filtered
	}

	if status != "" {
		filtered := []vulns.Vuln{}
		for _, v := range vlist {
			if string(v.Status) == status {
				filtered = append(filtered, v)
			}
		}
		vlist = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"vulns": vulnsToList(vlist),
		"total": len(vlist),
	})
}

func handleVulnCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title       string   `json:"title"`
		Severity    string   `json:"severity"`
		Description string   `json:"description"`
		Remediation string   `json:"remediation"`
		References  []string `json:"references"`
		Evidence    []string `json:"evidence"`
		Phase       string   `json:"phase"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, `{"error":"Title is required"}`, http.StatusBadRequest)
		return
	}

	mgr := configManager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		http.Error(w, `{"error":"No active context"}`, http.StatusBadRequest)
		return
	}

	vmgr := vulns.NewManager(ctx.Client, ctx.Engagement)
	id, _ := vmgr.GenerateID(req.Title)

	severity := vulns.SeverityInfo
	switch req.Severity {
	case "Critical":
		severity = vulns.SeverityCritical
	case "High":
		severity = vulns.SeverityHigh
	case "Medium":
		severity = vulns.SeverityMedium
	case "Low":
		severity = vulns.SeverityLow
	}

	v := vulns.Vuln{
		ID:          id,
		Title:       req.Title,
		Severity:    severity,
		Status:      vulns.StatusOpen,
		Description: req.Description,
		Remediation: req.Remediation,
		References:  req.References,
		Evidence:    req.Evidence,
		Phase:       req.Phase,
	}

	if err := vmgr.Save(v); err != nil {
		http.Error(w, `{"error":"Failed to save vulnerability"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vulnToMap(v))
}

func handleVulnGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	mgr := configManager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		http.Error(w, `{"error":"No active context"}`, http.StatusBadRequest)
		return
	}

	vmgr := vulns.NewManager(ctx.Client, ctx.Engagement)
	v, err := vmgr.Get(id)
	if err != nil {
		http.Error(w, `{"error":"Vulnerability not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vulnToMap(*v))
}

func handleVulnUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	mgr := configManager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		http.Error(w, `{"error":"No active context"}`, http.StatusBadRequest)
		return
	}

	vmgr := vulns.NewManager(ctx.Client, ctx.Engagement)
	v, err := vmgr.Get(id)
	if err != nil {
		http.Error(w, `{"error":"Vulnerability not found"}`, http.StatusNotFound)
		return
	}

	if title, ok := req["title"].(string); ok {
		v.Title = title
	}
	if severity, ok := req["severity"].(string); ok {
		switch severity {
		case "Critical":
			v.Severity = vulns.SeverityCritical
		case "High":
			v.Severity = vulns.SeverityHigh
		case "Medium":
			v.Severity = vulns.SeverityMedium
		case "Low":
			v.Severity = vulns.SeverityLow
		case "Info":
			v.Severity = vulns.SeverityInfo
		}
	}
	if status, ok := req["status"].(string); ok {
		switch status {
		case "Open":
			v.Status = vulns.StatusOpen
		case "Closed":
			v.Status = vulns.StatusClosed
		case "Verified":
			v.Status = vulns.StatusVerified
		}
	}
	if desc, ok := req["description"].(string); ok {
		v.Description = desc
	}
	if rem, ok := req["remediation"].(string); ok {
		v.Remediation = rem
	}
	if phase, ok := req["phase"].(string); ok {
		v.Phase = phase
	}

	if err := vmgr.Save(*v); err != nil {
		http.Error(w, `{"error":"Failed to update vulnerability"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vulnToMap(*v))
}

func handleVulnDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	mgr := configManager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		http.Error(w, `{"error":"No active context"}`, http.StatusBadRequest)
		return
	}

	vmgr := vulns.NewManager(ctx.Client, ctx.Engagement)
	if err := vmgr.Delete(id); err != nil {
		http.Error(w, `{"error":"Failed to delete vulnerability"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Vulnerability deleted"})
}

func vulnsToList(vlist []vulns.Vuln) []map[string]interface{} {
	var result []map[string]interface{}
	for _, v := range vlist {
		result = append(result, vulnToMap(v))
	}
	return result
}
