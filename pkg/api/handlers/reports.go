package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pentlog/pkg/config"

	"github.com/go-chi/chi/v5"
)

func ReportRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handleReportsList)
	return r
}

func handleReportsList(w http.ResponseWriter, r *http.Request) {
	client := r.URL.Query().Get("client")

	mgr := config.Manager()
	reportsDir := mgr.GetPaths().ReportsDir

	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"reports": []interface{}{}})
			return
		}
		http.Error(w, `{"error":"Failed to read reports directory"}`, http.StatusInternalServerError)
		return
	}

	var reports []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() {
			if client != "" && entry.Name() != client {
				continue
			}
			clientDir := filepath.Join(reportsDir, entry.Name())
			clientEntries, err := os.ReadDir(clientDir)
			if err != nil {
				continue
			}
			for _, ce := range clientEntries {
				if !ce.IsDir() && (strings.HasSuffix(ce.Name(), ".md") || strings.HasSuffix(ce.Name(), ".html")) {
					info, err := ce.Info()
					if err != nil {
						continue
					}
					reports = append(reports, map[string]interface{}{
						"name":       ce.Name(),
						"client":     entry.Name(),
						"path":       filepath.Join(clientDir, ce.Name()),
						"size":       info.Size(),
						"size_human": formatSize(info.Size()),
						"mod_time":   info.ModTime().Format("2006-01-02 15:04:05"),
						"type":       getFileExt(ce.Name()),
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"reports": reports})
}

func getFileExt(name string) string {
	if strings.HasSuffix(name, ".html") {
		return "html"
	}
	return "md"
}
