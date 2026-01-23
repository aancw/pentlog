package vulns

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"pentlog/pkg/config"
)

type Severity string

const (
	SeverityCritical Severity = "Critical"
	SeverityHigh     Severity = "High"
	SeverityMedium   Severity = "Medium"
	SeverityLow      Severity = "Low"
	SeverityInfo     Severity = "Info"
)

type Status string

const (
	StatusOpen     Status = "Open"
	StatusClosed   Status = "Closed"
	StatusVerified Status = "Verified"
)

type Vuln struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Severity    Severity  `json:"severity"`
	Status      Status    `json:"status"`
	Description string    `json:"description"`
	Remediation string    `json:"remediation"`
	References  []string  `json:"references"`
	Evidence    []string  `json:"evidence"`
	Phase       string    `json:"phase"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Manager struct {
	Client     string
	Engagement string
}

func NewManager(client, engagement string) *Manager {
	return &Manager{
		Client:     client,
		Engagement: engagement,
	}
}

func NewManagerFromContext() (*Manager, error) {
	mgr := config.Manager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		return nil, err
	}
	return NewManager(ctx.Client, ctx.Engagement), nil
}

func (m *Manager) GetVulnsDir() (string, error) {
	mgr := config.Manager()
	baseDir := mgr.GetPaths().Home
	return filepath.Join(baseDir, "vulns", m.Client, m.Engagement), nil
}

func (m *Manager) GetVulnsFile() (string, error) {
	dir, err := m.GetVulnsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "vulns.json"), nil
}

func (m *Manager) Save(vuln Vuln) error {
	vulns, err := m.List()
	if err != nil {
		return err
	}

	found := false
	for i, v := range vulns {
		if v.ID == vuln.ID {
			vuln.UpdatedAt = time.Now()
			vulns[i] = vuln
			found = true
			break
		}
	}
	if !found {
		if vuln.CreatedAt.IsZero() {
			vuln.CreatedAt = time.Now()
		}
		if vuln.UpdatedAt.IsZero() {
			vuln.UpdatedAt = time.Now()
		}
		vulns = append(vulns, vuln)
	}

	return m.writeVulns(vulns)
}

func (m *Manager) List() ([]Vuln, error) {
	if m.Engagement == "" {
		return m.listAllEngagements()
	}

	path, err := m.GetVulnsFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []Vuln{}, nil
	}

	return m.readVulnsFromFile(path)
}

func (m *Manager) listAllEngagements() ([]Vuln, error) {
	mgr := config.Manager()
	baseDir := mgr.GetPaths().Home
	clientDir := filepath.Join(baseDir, "vulns", m.Client)

	entries, err := os.ReadDir(clientDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Vuln{}, nil
		}
		return nil, err
	}

	var allVulns []Vuln
	for _, entry := range entries {
		if entry.IsDir() {
			engagementPath := filepath.Join(clientDir, entry.Name(), "vulns.json")
			if _, err := os.Stat(engagementPath); err == nil {
				vulns, err := m.readVulnsFromFile(engagementPath)
				if err != nil {
					continue
				}
				allVulns = append(allVulns, vulns...)
			}
		}
	}

	sort.Slice(allVulns, func(i, j int) bool {
		return allVulns[i].CreatedAt.After(allVulns[j].CreatedAt)
	})

	return allVulns, nil
}

func (m *Manager) readVulnsFromFile(path string) ([]Vuln, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var vulns []Vuln
	if err := json.Unmarshal(data, &vulns); err != nil {
		return nil, err
	}

	sort.Slice(vulns, func(i, j int) bool {
		return vulns[i].CreatedAt.After(vulns[j].CreatedAt)
	})

	return vulns, nil
}

func (m *Manager) Get(id string) (*Vuln, error) {
	vulns, err := m.List()
	if err != nil {
		return nil, err
	}
	for _, v := range vulns {
		if v.ID == id {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("vuln not found: %s", id)
}

func (m *Manager) Delete(id string) error {
	vulns, err := m.List()
	if err != nil {
		return err
	}

	newVulns := []Vuln{}
	for _, v := range vulns {
		if v.ID != id {
			newVulns = append(newVulns, v)
		}
	}

	return m.writeVulns(newVulns)
}

func (m *Manager) writeVulns(vulns []Vuln) error {
	path, err := m.GetVulnsFile()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(vulns, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (v Vuln) SeverityColor() string {
	switch v.Severity {
	case SeverityCritical:
		return "#ff0000"
	case SeverityHigh:
		return "#ff6600"
	case SeverityMedium:
		return "#ffcc00"
	case SeverityLow:
		return "#00cc00"
	case SeverityInfo:
		return "#0099ff"
	default:
		return "#ffffff"
	}
}

func (m *Manager) GenerateID(title string) (string, error) {
	vulns, err := m.List()
	if err != nil {
		return "", err
	}

	count := len(vulns) + 1
	slug := strings.ToLower(strings.Join(strings.Fields(title), "-"))
	if len(slug) > 20 {
		slug = slug[:20]
	}

	return fmt.Sprintf("vuln-%03d", count), nil
}
