package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"pentlog/pkg/config"

	"github.com/go-chi/chi/v5"
)

type shareSession struct {
	PID     int    `json:"pid"`
	LogFile string `json:"log_file"`
	Token   string `json:"token"`
	URL     string `json:"url"`
	Port    int    `json:"port"`
}

func ShareRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/status", handleShareStatus)
	return r
}

func handleShareStatus(w http.ResponseWriter, r *http.Request) {
	session, err := loadShareSession()
	if err != nil {
		http.Error(w, `{"error":"Failed to load share session"}`, http.StatusInternalServerError)
		return
	}
	if session == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"active": false,
		})
		return
	}

	watchURL := fmt.Sprintf("http://%s/watch?token=%s", session.URL, session.Token)
	statusURL := fmt.Sprintf("http://%s/status?token=%s", session.URL, session.Token)

	resp := map[string]interface{}{
		"active":     true,
		"pid":        session.PID,
		"log_file":   session.LogFile,
		"watch_url":  watchURL,
		"status_url": statusURL,
		"port":       session.Port,
		"reachable":  false,
		"clients":    0,
		"client_ips": []string{},
	}

	client := &http.Client{Timeout: 1500 * time.Millisecond}
	statusResp, err := client.Get(statusURL)
	if err == nil {
		defer statusResp.Body.Close()
		var status struct {
			Clients   int      `json:"clients"`
			ClientIPs []string `json:"client_ips"`
		}
		if statusResp.StatusCode == http.StatusOK && json.NewDecoder(statusResp.Body).Decode(&status) == nil {
			resp["reachable"] = true
			resp["clients"] = status.Clients
			resp["client_ips"] = status.ClientIPs
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func loadShareSession() (*shareSession, error) {
	mgr := config.Manager()
	path := filepath.Join(mgr.GetPaths().Home, ".share_session")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var session shareSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}
