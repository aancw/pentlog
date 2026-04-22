package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pentlog/pkg/config"

	"github.com/go-chi/chi/v5"
)

type archiveRecord struct {
	Name         string `json:"name"`
	Client       string `json:"client"`
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Size         int64  `json:"size"`
	SizeHuman    string `json:"size_human"`
	ModTime      string `json:"mod_time"`
	Encrypted    bool   `json:"encrypted"`
	DownloadURL  string `json:"download_url"`
}

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
			json.NewEncoder(w).Encode(map[string]interface{}{"archives": []archiveRecord{}})
			return
		}
		http.Error(w, `{"error":"Failed to read archive directory"}`, http.StatusInternalServerError)
		return
	}

	archives := make([]archiveRecord, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if client != "" && entry.Name() != client {
			continue
		}

		clientDir := filepath.Join(archiveDir, entry.Name())
		clientEntries, err := os.ReadDir(clientDir)
		if err != nil {
			continue
		}

		for _, ce := range clientEntries {
			if ce.IsDir() || !strings.HasSuffix(ce.Name(), ".zip") {
				continue
			}

			info, err := ce.Info()
			if err != nil {
				continue
			}

			relPath := filepath.ToSlash(filepath.Join(entry.Name(), ce.Name()))
			archives = append(archives, archiveRecord{
				Name:         ce.Name(),
				Client:       entry.Name(),
				Path:         filepath.Join(clientDir, ce.Name()),
				RelativePath: relPath,
				Size:         info.Size(),
				SizeHuman:    formatSize(info.Size()),
				ModTime:      info.ModTime().Format("2006-01-02 15:04:05"),
				Encrypted:    false,
				DownloadURL:  "/files/archives/" + url.PathEscape(entry.Name()) + "/" + url.PathEscape(ce.Name()),
			})
		}
	}

	sort.Slice(archives, func(i, j int) bool {
		return archives[i].ModTime > archives[j].ModTime
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"archives": archives})
}
