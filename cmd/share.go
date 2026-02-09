package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/logs"
	"pentlog/pkg/share"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	sharePort       int
	shareBind       string
	shareFromStart  bool
	shareDaemonMode bool
)

type ShareSession struct {
	PID       int    `json:"pid"`
	SessionID string `json:"session_id"`
	LogFile   string `json:"log_file"`
	Token     string `json:"token"`
	URL       string `json:"url"`
	Port      int    `json:"port"`
}

var shareCmd = &cobra.Command{
	Use:   "share [session-id]",
	Short: "Share a live session for read-only viewing",
	Long: `Start a WebSocket server that streams a session to remote viewers via a browser.
Viewers connect with a secure token URL and see the terminal output in real-time
using xterm.js. No input is accepted from viewers.

The server runs in the background and doesn't block the terminal.

USAGE:
  From within 'pentlog shell':
    $ pentlog share                          (no session-id needed)

  From another terminal:
    $ pentlog share <session-id>             (specify session ID)
    $ pentlog share <session-id> --port 8080 (custom port)

  Manage sharing:
    $ pentlog share status                   (check active share)
    $ pentlog share stop                     (stop sharing)`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if shareDaemonMode {
			logFilePath := resolveLogFile(args)
			if logFilePath == "" {
				os.Exit(1)
			}
			startShareServerDaemon(logFilePath)
			return
		}

		logFilePath := resolveLogFile(args)
		if logFilePath == "" {
			return
		}

		if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: session file not found: %s\n", logFilePath)
			os.Exit(1)
		}

		existing, _ := loadShareSession()
		if existing != nil {
			fmt.Fprintf(os.Stderr, "Error: Share already active on PID %d\n", existing.PID)
			fmt.Fprintf(os.Stderr, "Stop it first with: pentlog share stop\n")
			os.Exit(1)
		}

		spawnShareDaemon(logFilePath)
	},
}

var shareStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the active sharing session",
	Run: func(cmd *cobra.Command, args []string) {
		session, err := loadShareSession()
		if err != nil || session == nil {
			fmt.Println("No active share session found.")
			return
		}

		process, err := os.FindProcess(session.PID)
		if err != nil {
			fmt.Printf("Error: Could not find process (PID %d): %v\n", session.PID, err)
			removeShareSession()
			return
		}

		if err := process.Signal(syscall.SIGTERM); err != nil {
			fmt.Printf("Error: Could not signal process: %v\n", err)
			return
		}

		fmt.Printf("Stopped share server (PID %d)\n", session.PID)
		removeShareSession()
	},
}

var shareStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show active share session info",
	Run: func(cmd *cobra.Command, args []string) {
		session, err := loadShareSession()
		if err != nil || session == nil {
			fmt.Println("No active share session.")
			return
		}

		process, err := os.FindProcess(session.PID)
		if err != nil || process == nil {
			fmt.Println("Share session file exists but process not found. Cleaning up...")
			removeShareSession()
			return
		}

		fmt.Println()
		fmt.Println("  PentLog Live Share (Active)")
		fmt.Println("  ──────────────────────────────────────────")
		fmt.Printf("  URL:     http://%s/watch?token=%s\n", session.URL, session.Token)
		fmt.Printf("  PID:     %d\n", session.PID)

		statusURL := fmt.Sprintf("http://%s/status?token=%s", session.URL, session.Token)
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get(statusURL)
		if err == nil {
			defer resp.Body.Close()
			var status struct {
				Clients   int      `json:"clients"`
				ClientIPs []string `json:"client_ips"`
			}
			if json.NewDecoder(resp.Body).Decode(&status) == nil {
				fmt.Printf("  Viewers: %d\n", status.Clients)
				if len(status.ClientIPs) > 0 {
					for _, ip := range status.ClientIPs {
						fmt.Printf("           - %s\n", ip)
					}
				}
			}
		}

		fmt.Println("  ──────────────────────────────────────────")
		fmt.Println()
	},
}

func displayShareInfo(session *ShareSession) {
	fmt.Println()
	fmt.Println("  PentLog Live Share")
	fmt.Println("  ──────────────────────────────────────────")
	fmt.Printf("  URL:   http://%s/watch?token=%s\n", session.URL, session.Token)
	fmt.Println("  ──────────────────────────────────────────")
	fmt.Println()
}

func spawnShareDaemon(logFilePath string) {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not get executable path: %v\n", err)
		os.Exit(1)
	}

	args := []string{
		"share",
		"--daemon-mode",
	}

	if sharePort != 0 {
		args = append(args, "--port", fmt.Sprintf("%d", sharePort))
	}
	if shareBind != "0.0.0.0" {
		args = append(args, "--bind", shareBind)
	}
	if !shareFromStart {
		args = append(args, "--from-start=false")
	}

	cmd := exec.Command(exe, args...)

	cmd.Env = os.Environ()
	if logFilePath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PENTLOG_SESSION_LOG_PATH=%s", logFilePath))
	}

	logPath := filepath.Join(config.Manager().GetPaths().Home, ".share_server.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err == nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		defer logFile.Close()
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting share server: %v\n", err)
		os.Exit(1)
	}

	pid := cmd.Process.Pid

	for i := 0; i < 100; i++ {
		time.Sleep(100 * time.Millisecond)
		session, _ := loadShareSession()
		if session != nil && session.PID == pid {
			displayShareInfo(session)
			return
		}
	}

	fmt.Printf("Share server started (PID %d)\n", pid)
	fmt.Println("Initializing... Use 'pentlog share status' to check status.")
}

func startShareServerDaemon(logFilePath string) {
	token, err := share.GenerateToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating access token: %v\n", err)
		os.Exit(1)
	}

	hub := share.NewHub()
	go hub.Run()

	srv := share.NewServer(hub, token)

	listenAddr := fmt.Sprintf("%s:%d", shareBind, sharePort)
	boundAddr, err := srv.Start(listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}

	displayAddr := boundAddr
	if shareBind == "0.0.0.0" {
		if ip := getOutboundIP(); ip != "" {
			_, port, _ := net.SplitHostPort(boundAddr)
			displayAddr = net.JoinHostPort(ip, port)
		}
	}

	_, portStr, _ := net.SplitHostPort(boundAddr)
	port := 0
	fmt.Sscanf(portStr, "%d", &port)

	session := &ShareSession{
		PID:     os.Getpid(),
		LogFile: logFilePath,
		Token:   token,
		URL:     displayAddr,
		Port:    port,
	}
	saveShareSession(session)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go runTailBroadcast(ctx, hub, logFilePath)

	<-sigChan
	cancel()
	srv.Stop()
	hub.Stop()
	removeShareSession()
	fmt.Println("Share server stopped.")
}

func getShareSessionPath() string {
	mgr := config.Manager()
	return filepath.Join(mgr.GetPaths().Home, ".share_session")
}

func saveShareSession(session *ShareSession) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(getShareSessionPath(), data, 0600)
}

func loadShareSession() (*ShareSession, error) {
	data, err := ioutil.ReadFile(getShareSessionPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var session ShareSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func removeShareSession() error {
	return os.Remove(getShareSessionPath())
}

func resolveLogFile(args []string) string {
	if len(args) == 1 {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid session ID: %s\n", args[0])
			os.Exit(1)
		}
		session, err := logs.GetSession(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return session.Path
	}

	if envPath := os.Getenv("PENTLOG_SESSION_LOG_PATH"); envPath != "" {
		return envPath
	}

	fmt.Fprintln(os.Stderr, "Error: No session specified.")
	fmt.Fprintln(os.Stderr, "  Usage: pentlog share <session-id>")
	fmt.Fprintln(os.Stderr, "  Or run inside a pentlog shell session.")
	os.Exit(1)
	return ""
}

func runTailBroadcast(ctx context.Context, hub *share.Hub, logFilePath string) {
	tailer := share.NewTailer(logFilePath, shareFromStart)
	rawCh := make(chan []byte, 64)

	go func() {
		if err := tailer.Run(ctx, rawCh); err != nil && ctx.Err() == nil {
			fmt.Fprintf(os.Stderr, "Tailer error: %v\n", err)
		}
	}()

	parser := share.NewTtyrecParser()

	for {
		select {
		case <-ctx.Done():
			return
		case data, ok := <-rawCh:
			if !ok {
				return
			}
			parser.Feed(data)
			for {
				frame, err := parser.Next()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
					return
				}
				if frame == nil {
					break
				}
				hub.Broadcast(frame.Payload)
			}
		}
	}
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}

func init() {
	shareCmd.AddCommand(shareStopCmd)
	shareCmd.AddCommand(shareStatusCmd)

	shareCmd.Flags().IntVarP(&sharePort, "port", "p", 0, "Port to listen on (0 = random)")
	shareCmd.Flags().StringVar(&shareBind, "bind", "0.0.0.0", "Address to bind to")
	shareCmd.Flags().BoolVar(&shareFromStart, "from-start", true, "Stream from the beginning of the file")
	shareCmd.Flags().BoolVar(&shareDaemonMode, "daemon-mode", false, "Internal: run as daemon process")
	shareCmd.Flags().MarkHidden("daemon-mode")
	rootCmd.AddCommand(shareCmd)
}
