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

type reportRecord struct {
	Name         string `json:"name"`
	Client       string `json:"client"`
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
	Size         int64  `json:"size"`
	SizeHuman    string `json:"size_human"`
	ModTime      string `json:"mod_time"`
	Type         string `json:"type"`
	ViewURL      string `json:"view_url"`
}

func ReportRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handleReportsList)
	r.Post("/generate", handleReportsGenerate)
	r.Get("/jobs/{id}", handleReportsJobByID)
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
			json.NewEncoder(w).Encode(map[string]interface{}{"reports": []reportRecord{}})
			return
		}
		http.Error(w, `{"error":"Failed to read reports directory"}`, http.StatusInternalServerError)
		return
	}

	reports := make([]reportRecord, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if client != "" && entry.Name() != client {
			continue
		}

		clientDir := filepath.Join(reportsDir, entry.Name())
		clientEntries, err := os.ReadDir(clientDir)
		if err != nil {
			continue
		}

		for _, ce := range clientEntries {
			if ce.IsDir() || (!strings.HasSuffix(ce.Name(), ".md") && !strings.HasSuffix(ce.Name(), ".html")) {
				continue
			}

			info, err := ce.Info()
			if err != nil {
				continue
			}

			relPath := filepath.ToSlash(filepath.Join(entry.Name(), ce.Name()))
			reports = append(reports, reportRecord{
				Name:         ce.Name(),
				Client:       entry.Name(),
				Path:         filepath.Join(clientDir, ce.Name()),
				RelativePath: relPath,
				Size:         info.Size(),
				SizeHuman:    formatSize(info.Size()),
				ModTime:      info.ModTime().Format("2006-01-02 15:04:05"),
				Type:         getFileExt(ce.Name()),
				ViewURL:      "/files/reports/" + url.PathEscape(entry.Name()) + "/" + url.PathEscape(ce.Name()),
			})
		}
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].ModTime > reports[j].ModTime
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"reports": reports})
}

func getFileExt(name string) string {
	if strings.HasSuffix(name, ".html") {
		return "html"
	}
	return "md"
}
