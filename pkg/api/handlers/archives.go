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

func ArchiveRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handleArchivesList)
	return r
}

func handleArchivesList(w http.ResponseWriter, r *http.Request) {
	client := r.URL.Query().Get("client")

	mgr := config.Manager()
	archiveDir := mgr.GetPaths().ArchiveDir

	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"archives": []interface{}{}})
			return
		}
		http.Error(w, `{"error":"Failed to read archive directory"}`, http.StatusInternalServerError)
		return
	}

	var archives []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() {
			if client != "" && entry.Name() != client {
				continue
			}
			clientDir := filepath.Join(archiveDir, entry.Name())
			clientEntries, err := os.ReadDir(clientDir)
			if err != nil {
				continue
			}
			for _, ce := range clientEntries {
				if !ce.IsDir() && strings.HasSuffix(ce.Name(), ".zip") {
					info, err := ce.Info()
					if err != nil {
						continue
					}
					archives = append(archives, map[string]interface{}{
						"name":       ce.Name(),
						"client":     entry.Name(),
						"path":       filepath.Join(clientDir, ce.Name()),
						"size":       info.Size(),
						"size_human": formatSize(info.Size()),
						"mod_time":   info.ModTime().Format("2006-01-02 15:04:05"),
						"encrypted":  false,
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"archives": archives})
}
