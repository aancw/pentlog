package deps

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Dependency struct {
	Name        string
	InstallCmds map[string][]string // OS/Distro -> Command parts
	CheckCmd    string
}

type Manager struct {
	Dependencies []Dependency
}

func NewManager() *Manager {
	return &Manager{
		Dependencies: []Dependency{
			{
				Name: "ttyrec",
				InstallCmds: map[string][]string{
					"darwin":  {"brew", "install", "ttyrec"},
					"ubuntu":  {"sudo", "apt-get", "install", "-y", "ttyrec"},
					"debian":  {"sudo", "apt-get", "install", "-y", "ttyrec"},
					"fedora":  {"sudo", "dnf", "install", "-y", "https://github.com/ovh/ovh-ttyrec/releases/download/v1.1.7.1/ovh-ttyrec-1.1.7.1-1.x86_64.rpm"},
					"alpine":  {"sudo", "apk", "add", "ttyrec"},
					"unknown": nil, // Manual install required
				},
			},
			{
				Name: "ttyplay",
				InstallCmds: map[string][]string{
					"darwin": {"brew", "install", "ttyrec"}, // ttyplay comes with ttyrec on brew usually
					"ubuntu": {"sudo", "apt-get", "install", "-y", "ttyrec"},
					"debian": {"sudo", "apt-get", "install", "-y", "ttyrec"},
					"fedora": {"sudo", "dnf", "install", "-y", "ttyrec"},
					"alpine": {"sudo", "apk", "add", "ttyrec"},
				},
			},
		},
	}
}

func (m *Manager) Check(name string) (bool, string) {
	path, err := exec.LookPath(name)
	if err != nil {
		return false, ""
	}
	return true, path
}

func (m *Manager) Install(name string) error {
	dep := m.findDependency(name)
	if dep == nil {
		return fmt.Errorf("dependency %s not known", name)
	}

	distro := detectDistro()
	cmdParts, ok := dep.InstallCmds[distro]
	if !ok {
		cmdParts, ok = dep.InstallCmds[runtime.GOOS]
	}

	if !ok || len(cmdParts) == 0 {
		if name == "ttyrec" {
			return fmt.Errorf(
				"no install command found for %s on %s/%s. Please install ttyrec manually for your distro (see https://github.com/ovh/ovh-ttyrec/releases/tag/v1.1.7.1) or open a PentLog issue to request support",
				name,
				runtime.GOOS,
				distro,
			)
		}
		return fmt.Errorf("no install command found for %s on %s/%s", name, runtime.GOOS, distro)
	}

	fmt.Printf("Installing %s using: %s\n", name, strings.Join(cmdParts, " "))
	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *Manager) findDependency(name string) *Dependency {
	for _, d := range m.Dependencies {
		if d.Name == name {
			return &d
		}
	}
	return nil
}

func detectDistro() string {
	if runtime.GOOS == "darwin" {
		return "darwin"
	}
	if _, err := exec.LookPath("apt-get"); err == nil {
		return "ubuntu"
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return "fedora"
	}
	if _, err := exec.LookPath("apk"); err == nil {
		return "alpine"
	}
	return "unknown"
}
