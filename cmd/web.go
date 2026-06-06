package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"pentlog/pkg/api"
	"pentlog/pkg/api/handlers"
	"pentlog/pkg/httpauth"
	"pentlog/pkg/utils"

	"github.com/spf13/cobra"
)

var (
	webPort      int
	webBind      string
	webOpen      bool
	webRebuild   bool
	webPublic    bool
	webAuthToken string
	webBasicAuth string
)

const autoGenerateAuthValue = "__auto_generate__"

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start web dashboard server",
	Long: `Start the PentLog web dashboard server to view sessions, reports,
and manage your penetration testing evidence through a web interface.

The web dashboard provides:
  - Dashboard with statistics and activity
  - Session browser with search and filters
  - Session replay with timeline
  - Vulnerability management
  - Report generation and AI analysis
  - Archive management

Examples:
  pentlog web                    # Start on default port 8080
  pentlog web --port 3000        # Start on port 3000
  pentlog web --open             # Open in browser after starting
  pentlog web --bind 0.0.0.0 --auth-token
  pentlog web --bind 0.0.0.0 --basic-auth pentester:changeme`,
	Run: func(cmd *cobra.Command, args []string) {
		handlers.Version = Version

		distDir, distErr := findBuiltWebAssets()
		if webRebuild || distErr != nil {
			if rebuiltDir, err := rebuildWebAssets(); err != nil {
				fmt.Printf("Warning: failed to rebuild web assets: %v\n", err)
			} else {
				distDir = rebuiltDir
			}
		}
		if distDir != "" {
			api.SetStaticDir(distDir)
		}

		authConfig, err := resolveWebAuthConfig(cmd, webBind)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		server := api.NewServer(webBind, webPort, authConfig)

		if webOpen {
			url := buildWebOpenURL(webBind, webPort, authConfig)
			go func() {
				time.Sleep(750 * time.Millisecond)
				if err := utils.OpenURL(url); err != nil {
					fmt.Printf("Warning: failed to open browser: %v\n", err)
				}
			}()
		}

		printWebExposureSummary(webBind, webPort, authConfig)

		if err := server.Start(); err != nil {
			fmt.Printf("Error starting server: %v\n", err)
		}
	},
}

func init() {
	webCmd.Flags().IntVarP(&webPort, "port", "p", 8080, "Port to listen on")
	webCmd.Flags().StringVar(&webBind, "bind", "127.0.0.1", "Address to bind to")
	webCmd.Flags().BoolVarP(&webOpen, "open", "o", false, "Open in browser after starting")
	webCmd.Flags().BoolVar(&webRebuild, "rebuild", false, "Rebuild frontend assets before starting the web server")
	webCmd.Flags().BoolVar(&webPublic, "public", false, "Explicitly allow non-loopback exposure without auth")
	webCmd.Flags().StringVar(&webAuthToken, "auth-token", "", "Require token auth for UI and API routes; provide a token or omit the value to auto-generate one")
	webCmd.Flags().StringVar(&webBasicAuth, "basic-auth", "", "Require HTTP basic auth for UI and API routes; format username:password or omit the value to auto-generate credentials")
	webCmd.Flags().Lookup("auth-token").NoOptDefVal = autoGenerateAuthValue
	webCmd.Flags().Lookup("basic-auth").NoOptDefVal = autoGenerateAuthValue
	rootCmd.AddCommand(webCmd)
}

func resolveWebAuthConfig(cmd *cobra.Command, bind string) (httpauth.Config, error) {
	tokenMode := cmd.Flags().Lookup("auth-token").Changed
	basicMode := cmd.Flags().Lookup("basic-auth").Changed

	if webPublic && (tokenMode || basicMode) {
		return httpauth.Config{}, errors.New("--public cannot be combined with --auth-token or --basic-auth")
	}
	if tokenMode && basicMode {
		return httpauth.Config{}, errors.New("--auth-token cannot be combined with --basic-auth")
	}
	if !httpauth.IsLoopbackBind(bind) && !webPublic && !tokenMode && !basicMode {
		return httpauth.Config{}, fmt.Errorf("refusing non-loopback bind on %s without an explicit exposure mode; use --public, --auth-token, or --basic-auth", bind)
	}

	if tokenMode {
		token, err := resolveTokenValue(webAuthToken)
		if err != nil {
			return httpauth.Config{}, err
		}
		return httpauth.TokenConfig(token), nil
	}

	if basicMode {
		username, password, err := resolveBasicAuthValue(webBasicAuth)
		if err != nil {
			return httpauth.Config{}, err
		}
		return httpauth.BasicConfig(username, password), nil
	}

	return httpauth.Config{}, nil
}

func resolveTokenValue(raw string) (string, error) {
	if raw == autoGenerateAuthValue {
		return httpauth.GenerateToken(32)
	}
	if raw == "" {
		return "", errors.New("--auth-token requires a token value or no value at all for auto-generation")
	}
	return raw, nil
}

func resolveBasicAuthValue(raw string) (string, string, error) {
	if raw == autoGenerateAuthValue {
		password, err := httpauth.GenerateToken(18)
		if err != nil {
			return "", "", err
		}
		return "pentlog", password, nil
	}

	username, password, ok := httpauth.ParseBasicAuthCredentials(raw)
	if !ok {
		return "", "", errors.New("--basic-auth must use username:password format or no value at all for auto-generated credentials")
	}

	return username, password, nil
}

func buildWebOpenURL(bind string, port int, authConfig httpauth.Config) string {
	openHost := bind
	switch bind {
	case "", "0.0.0.0", "::":
		openHost = "127.0.0.1"
	}

	base := fmt.Sprintf("http://%s:%d", openHost, port)
	if authConfig.Mode != httpauth.ModeToken {
		return base
	}
	return authConfig.AppendTokenQuery(base)
}

func printWebExposureSummary(bind string, port int, authConfig httpauth.Config) {
	address := fmt.Sprintf("%s:%d", bind, port)
	if httpauth.IsLoopbackBind(bind) && !authConfig.Enabled() {
		fmt.Printf(" Access mode: loopback only on %s\n", address)
		return
	}

	switch authConfig.Mode {
	case httpauth.ModeToken:
		fmt.Printf(" Access mode: token auth on %s\n", address)
		fmt.Printf(" Browser URL: %s\n", authConfig.AppendTokenQuery(fmt.Sprintf("http://%s", address)))
		fmt.Printf(" API token: %s\n", authConfig.Token)
	case httpauth.ModeBasic:
		fmt.Printf(" Access mode: basic auth on %s\n", address)
		fmt.Printf(" Username: %s\n", authConfig.Username)
		fmt.Printf(" Password: %s\n", authConfig.Password)
	default:
		fmt.Printf(" Access mode: public on %s\n", address)
	}
}

func findBuiltWebAssets() (string, error) {
	frontendDir, err := findFrontendDir()
	if err != nil {
		return "", err
	}

	distDir := filepath.Clean(filepath.Join(frontendDir, "..", "dist"))
	indexPath := filepath.Join(distDir, "index.html")
	if info, err := os.Stat(indexPath); err == nil && !info.IsDir() {
		return distDir, nil
	}

	return "", fmt.Errorf("built frontend assets not found")
}

func rebuildWebAssets() (string, error) {
	frontendDir, err := findFrontendDir()
	if err != nil {
		return "", err
	}

	fmt.Println("Rebuilding web assets...")
	buildCmd := exec.Command("npm", "run", "build")
	buildCmd.Dir = frontendDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return "", err
	}

	return filepath.Clean(filepath.Join(frontendDir, "..", "dist")), nil
}

func findFrontendDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(cwd, "pkg", "web", "frontend", "package.json")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return filepath.Dir(candidate), nil
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", fmt.Errorf("frontend source not found from %s", cwd)
		}
		cwd = parent
	}
}
