package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/deps"
	"pentlog/pkg/errors"
	"pentlog/pkg/logger"
	"pentlog/pkg/logs"
	"pentlog/pkg/share"
	"pentlog/pkg/system"
	"pentlog/pkg/utils"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

const heartbeatInterval = 30 * time.Second

var (
	shellShare          bool
	shellSharePort      int
	shellShareBind      string
	shellPhaseOverride  string
	shellTargetOverride string
	shellReview         bool
)

type shellPreflightPlan struct {
	Saved         config.ContextData
	Effective     config.ContextData
	Targets       []config.Target
	RecentChanges []string
	Warnings      []string
	PendingChange string
}

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start a recorded shell with the engagement context loaded",
	Run: func(cmd *cobra.Command, args []string) {
		runShellLaunch(shellReview)
	},
}

var shellReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review shell context before starting a recording",
	Run: func(cmd *cobra.Command, args []string) {
		runShellLaunch(true)
	},
}

func runShellLaunch(withReview bool) {
	if os.Getenv("PENTLOG_SESSION_LOG_PATH") != "" {
		errors.AlreadyInShell().Fatal()
	}

	dm := deps.NewManager()
	if ok, _ := dm.Check("ttyrec"); !ok {
		errors.MissingDependency("ttyrec", "brew install ttyrec || apt-get install ttyrec").Fatal()
	}

	mgr := config.Manager()
	ctx, err := mgr.LoadContext()
	if err != nil {
		errors.NoContext().Fatal()
	}

	if withReview {
		proceed := false
		var reviewErr error
		ctx, proceed, reviewErr = runShellPreflight(mgr, ctx)
		if reviewErr != nil {
			errors.FromError(errors.InvalidContext, "failed to prepare shell context", reviewErr).Fatal()
		}
		if !proceed {
			fmt.Println("Shell launch cancelled.")
			return
		}
	} else {
		ctx, err = applyShellContextOverrides(mgr, ctx)
		if err != nil {
			errors.FromError(errors.InvalidContext, "failed to prepare shell context", err).Fatal()
		}
	}

	crashed, err := logs.GetCrashedSessionsForContext(ctx.Client, ctx.Engagement, ctx.Phase)
	if err == nil && len(crashed) > 0 {
		resumeSession := promptResumeSession(crashed)
		if resumeSession != nil {
			startResumedSession(ctx, resumeSession)
			return
		}
	}

	logDir, err := system.EnsureLogDir()
	if err != nil {
		errors.FromError(errors.DirectoryNotFound, "failed to prepare log directory", err).Fatal()
	}

	sessionDir := getSessionDir(logDir, ctx)
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		errors.FromError(errors.PermissionDenied, "failed to create session directory", err).Fatal()
	}

	timestamp := time.Now().Format("20060102-150405")
	baseName := fmt.Sprintf("session-%s-%s", utils.Slugify(ctx.Operator), timestamp)
	if ctx.Target != "" {
		baseName = fmt.Sprintf("session-%s-%s-%s", utils.Slugify(ctx.Operator), utils.Slugify(ctx.Target), timestamp)
	}
	logFilePath := filepath.Join(sessionDir, baseName+".tty")
	metaFilePath := filepath.Join(sessionDir, baseName+".json")

	meta := logs.SessionMetadata{
		Client:     ctx.Client,
		Engagement: ctx.Engagement,
		Scope:      ctx.Scope,
		Operator:   ctx.Operator,
		Phase:      ctx.Phase,
		Target:     ctx.Target,
		TargetIP:   ctx.TargetIP,
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	if err := writeMetadata(metaFilePath, meta); err != nil {
		errors.FromError(errors.FileNotFound, "failed to write session metadata", err).Fatal()
	}

	sessionID, err := logs.AddSessionToDB(meta, logFilePath)
	if err != nil {
		logger.Warn("failed to add session to database", "error", err)
	}

	newEnv, tempDir, shellArgs, err := prepareShellEnv(ctx, sessionDir, metaFilePath, logFilePath, sessionID)
	if err != nil {
		errors.FromError(errors.Generic, "failed to prepare shell environment", err).Fatal()
	}
	if tempDir != "" {
		defer os.RemoveAll(tempDir)
	}

	recorder := system.NewRecorder()
	c, err := recorder.BuildCommand("", logFilePath, shellArgs...)
	if err != nil {
		errors.FromError(errors.Generic, "failed to create recorder command", err).Fatal()
	}

	hbCtx, hbCancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	cfg := mgr.GetMonitor()
	monitorConfig := logs.MonitorConfig{
		WarnThreshold:  int64(cfg.WarnThresholdMB) * 1024 * 1024,
		AlertThreshold: int64(cfg.AlertThresholdMB) * 1024 * 1024,
		CheckInterval:  time.Duration(cfg.CheckIntervalSec) * time.Second,
		AlertCooldown:  time.Duration(cfg.AlertCooldownMin) * time.Minute,
	}
	sessionMonitor := logs.NewSessionMonitor(logFilePath, monitorConfig)
	sessionMonitor.Start()

	if sessionID > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runHeartbeat(hbCtx, sessionID, logFilePath)
		}()
	}

	var shareCancel context.CancelFunc
	var shareSrv *share.Server
	var shareHub *share.Hub
	if shellShare {
		shareHub, shareSrv, shareCancel = startShareServer(logFilePath)
		if shareCancel != nil {
			defer shareCancel()
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	var signalReceived bool
	var mu sync.Mutex

	go func() {
		sig, ok := <-sigChan
		if !ok || sig == nil {
			return
		}
		mu.Lock()
		signalReceived = true
		mu.Unlock()

		hbCancel()

		if c.Process != nil {
			if c.SysProcAttr != nil && c.SysProcAttr.Setpgid {
				syscall.Kill(-c.Process.Pid, sig.(syscall.Signal))
			} else {
				c.Process.Signal(sig)
			}
		}
	}()

	runErr := startRecording(c, newEnv, ctx)

	hbCancel()
	wg.Wait()
	signal.Stop(sigChan)
	close(sigChan)

	sessionMonitor.Stop()

	if shareCancel != nil {
		shareCancel()
	}
	if shareSrv != nil {
		shareSrv.Stop()
	}
	if shareHub != nil {
		shareHub.Stop()
	}

	wasNormalExit := runErr == nil
	if sessionID > 0 {
		mu.Lock()
		if signalReceived {
			wasNormalExit = false
		}
		mu.Unlock()
		updateSessionOnExit(sessionID, logFilePath, wasNormalExit)
	}

	if runErr != nil {
		fmt.Fprintf(os.Stderr, "Error running recorder: %v\n", runErr)
		return
	}

	exitMsg := "\nLeaving pentlog shell session."
	if shellShare {
		exitMsg += " (share server stopped)"
	}
	fmt.Println(exitMsg)
}

func getSessionDir(logDir string, ctx *config.ContextData) string {
	if ctx.Type == "Log Only" {
		return filepath.Join(logDir, utils.Slugify(ctx.Client))
	}
	return filepath.Join(
		logDir,
		utils.Slugify(ctx.Client),
		utils.Slugify(ctx.Engagement),
		utils.Slugify(ctx.Phase),
	)
}

func prepareShellEnv(ctx *config.ContextData, sessionDir, metaFilePath, logFilePath string, sessionID int64) ([]string, string, []string, error) {
	tempDir, err := os.MkdirTemp("", "pentlog-shell-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create temp dir for shell config: %v\n", err)
	}

	newEnv := os.Environ()
	newEnv = append(newEnv, fmt.Sprintf("PENTLOG_CLIENT=%s", ctx.Client))
	newEnv = append(newEnv, fmt.Sprintf("PENTLOG_ENGAGEMENT=%s", ctx.Engagement))
	newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SCOPE=%s", ctx.Scope))
	newEnv = append(newEnv, fmt.Sprintf("PENTLOG_OPERATOR=%s", ctx.Operator))
	newEnv = append(newEnv, fmt.Sprintf("PENTLOG_PHASE=%s", ctx.Phase))
	newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SESSION_DIR=%s", sessionDir))
	newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SESSION_METADATA_PATH=%s", metaFilePath))
	newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SESSION_LOG_PATH=%s", logFilePath))
	if ctx.Target != "" {
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_TARGET=%s", ctx.Target))
	}
	if ctx.TargetIP != "" {
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_TARGET_IP=%s", ctx.TargetIP))
	}
	if sessionID > 0 {
		newEnv = append(newEnv, fmt.Sprintf("PENTLOG_SESSION_ID=%d", sessionID))
	}

	promptSegment := fmt.Sprintf("(pentlog:%s/%s)", ctx.Client, ctx.Phase)
	if ctx.Target != "" {
		promptSegment = fmt.Sprintf("(pentlog:%s/%s@%s)", ctx.Client, ctx.Phase, ctx.Target)
	}
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	baseShell := filepath.Base(shell)
	var shellArgs []string

	pentlogBin, _ := os.Executable()
	if pentlogBin == "" {
		pentlogBin = "pentlog"
	}

	if baseShell == "zsh" && tempDir != "" {
		zshrcPath := filepath.Join(tempDir, ".zshrc")
		userZshrc := filepath.Join(os.Getenv("HOME"), ".zshrc")
		zshContent := ""
		if _, err := os.Stat(userZshrc); err == nil {
			zshContent += fmt.Sprintf("source %s\n", userZshrc)
		}
		zshContent += "\n# Pentlog Prompt Injection\n"
		zshContent += "setopt TRANSIENT_RPROMPT\n"
		zshContent += fmt.Sprintf("RPROMPT=\"%%F{cyan}%s%%f $RPROMPT\"\n", promptSegment)
		zshContent += fmt.Sprintf(`
# Pentlog quick commands alias
alias pentlog="%s"

# Zsh widget bindings for Ctrl+N (note), Ctrl+G (vuln), Ctrl+O (pause), Ctrl+T (resume)
zle -N _pentlog_quicknote_widget
_pentlog_quicknote_widget() {
    zle push-line
    BUFFER="pentlog quicknote"
    zle accept-line
}
zle -N _pentlog_quickvuln_widget
_pentlog_quickvuln_widget() {
    zle push-line
    BUFFER="pentlog quickvuln"
    zle accept-line
}
zle -N _pentlog_pause_widget
_pentlog_pause_widget() {
    zle push-line
    BUFFER="pentlog pause"
    zle accept-line
}
zle -N _pentlog_resume_widget
_pentlog_resume_widget() {
    zle push-line
    BUFFER="pentlog resume"
    zle accept-line
}
bindkey '^N' _pentlog_quicknote_widget
bindkey '^G' _pentlog_quickvuln_widget
# Ctrl+O for pause, Ctrl+T for resume
bindkey '^O' _pentlog_pause_widget
bindkey '^T' _pentlog_resume_widget
`, pentlogBin)

		if err := os.WriteFile(zshrcPath, []byte(zshContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not write .zshrc: %v\n", err)
		} else {
			newEnv = append(newEnv, fmt.Sprintf("ZDOTDIR=%s", tempDir))
		}
	} else if baseShell == "bash" && tempDir != "" {
		bashrcPath := filepath.Join(tempDir, ".bashrc")
		userBashrc := filepath.Join(os.Getenv("HOME"), ".bashrc")
		bashContent := ""
		if _, err := os.Stat(userBashrc); err == nil {
			bashContent += fmt.Sprintf("source %s\n", userBashrc)
		}
		bashContent += fmt.Sprintf("\n# Pentlog quick commands alias\nalias pentlog=\"%s\"\n", pentlogBin)
		bashContent += "\n# Pentlog Transient RPROMPT (right side of input line)\n"
		bashContent += fmt.Sprintf(`_pentlog_prompt_text="%s"
_pentlog_marker="/tmp/.pentlog_rprompt_$$"

# Show RPROMPT on right side of input line
_pentlog_rprompt() {
    local cols=$(tput cols 2>/dev/null || echo 80)
    local text_len=${#_pentlog_prompt_text}
    local col=$((cols - text_len + 1))
    printf '\033[s\033[%%dG\033[0;36m%%s\033[0m\033[u' "$col" "$_pentlog_prompt_text"
    touch "$_pentlog_marker"
}

# Clear RPROMPT before command runs (transient)
_pentlog_preexec() {
    [[ -n "$COMP_LINE" ]] && return
    [[ ! -f "$_pentlog_marker" ]] && return
    [[ "$BASH_COMMAND" == *"_pentlog_"* ]] && return
    
    local cols=$(tput cols 2>/dev/null || echo 80)
    local text_len=${#_pentlog_prompt_text}
    local col=$((cols - text_len + 1))
    # Move up 1 line, go to RPROMPT column, clear to EOL, move back down
    printf '\033[A\033[%%dG\033[K\033[B\r' "$col"
    rm -f "$_pentlog_marker"
}

trap '_pentlog_preexec' DEBUG
trap 'rm -f "$_pentlog_marker"' EXIT

# Append RPROMPT display to end of PS1
PS1="$PS1"'\[$(_pentlog_rprompt)\]'

# Bash keybindings for Ctrl+N (note), Ctrl+G (vuln), Ctrl+O (pause), Ctrl+T (resume)
bind '"\C-n": "\C-apentlog quicknote\C-m"'
bind '"\C-g": "\C-apentlog quickvuln\C-m"'
# Ctrl+O for pause, Ctrl+T for resume
bind '"\C-o": "\C-apentlog pause\C-m"'
bind '"\C-t": "\C-apentlog resume\C-m"'
`, promptSegment)

		if err := os.WriteFile(bashrcPath, []byte(bashContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not write .bashrc: %v\n", err)
		} else {
			shellArgs = []string{"bash", "--rcfile", bashrcPath}
		}
	} else if baseShell == "sh" && tempDir != "" {
		shrcPath := filepath.Join(tempDir, ".shrc")
		userShrc := filepath.Join(os.Getenv("HOME"), ".shrc")
		shContent := ""
		if _, err := os.Stat(userShrc); err == nil {
			shContent += fmt.Sprintf(". %s\n", userShrc)
		}
		shContent += "\n# Pentlog Prompt Injection with RPROMPT support\n"
		shContent += fmt.Sprintf(`# Right-aligned prompt for RPROMPT support (RPROMPT emulation for sh)
_pentlog_rprompt="\033[0;36m%s\033[0m"
_pentlog_prompt_cmd() {
    # Print right-aligned prompt segment
    printf "%%s\n" "$_pentlog_rprompt" >&2
}
# sh uses ENV variable, set PROMPT_COMMAND equivalent if supported
PROMPT_COMMAND="_pentlog_prompt_cmd${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
PS1="\033[0;36m%s\033[0m $PS1"
`, promptSegment, promptSegment)

		if err := os.WriteFile(shrcPath, []byte(shContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not write .shrc: %v\n", err)
		} else {
			newEnv = append(newEnv, fmt.Sprintf("ENV=%s", shrcPath))
		}
	}

	return newEnv, tempDir, shellArgs, nil
}

func startRecording(c *exec.Cmd, env []string, ctx *config.ContextData) error {
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = env
	c.SysProcAttr = &syscall.SysProcAttr{
		Setpgid:    true,
		Foreground: true,
		Ctty:       int(os.Stdin.Fd()),
	}

	fmt.Println()
	fmt.Print(Banner)

	summary := buildContextSummaryLines(*ctx)
	utils.PrintBox("Active Session", summary)

	fmt.Println()
	utils.PrintCenteredBlock([]string{"⚠️  WARNING: All input (including passwords) is logged."})
	fmt.Println()
	hints := []string{
		"Type 'exit' or Ctrl+D to stop recording.",
		"Hotkeys: Ctrl+N = Note | Ctrl+G = Vuln | Ctrl+O = Pause | Ctrl+T = Resume",
	}
	utils.PrintCenteredBlock(hints)

	if shellShare {
		shareURL := os.Getenv("PENTLOG_SHARE_URL")
		if shareURL != "" {
			fmt.Println()
			printShareInfo(shareURL)
		}
	}

	if err := c.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0 {
				exitMsg := "\nLeaving pentlog shell session."
				if shellShare {
					exitMsg += " (share server stopped)"
				}
				fmt.Println(exitMsg)
				return nil
			}
		}
		return err
	}
	return nil
}

func printShareInfo(shareURL string) {
	width := utils.GetTerminalWidth()

	lines := []string{
		"PentLog Live Share",
		"──────────────────────────────────────────",
		fmt.Sprintf("URL:   %s", shareURL),
		"──────────────────────────────────────────",
		"",
		"Live sharing is active — viewers can watch this session.",
	}

	for _, line := range lines {
		strippedLen := runewidth.StringWidth(utils.StripANSI(line))
		padding := (width - strippedLen) / 2
		if padding < 0 {
			padding = 0
		}
		fmt.Println(strings.Repeat(" ", padding) + line)
	}
}

func startShareServer(logFilePath string) (*share.Hub, *share.Server, context.CancelFunc) {
	token, err := share.GenerateToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not start share server: %v\n", err)
		return nil, nil, nil
	}

	if ip := net.ParseIP(shellShareBind); ip != nil && !ip.IsLoopback() {
		fmt.Fprintf(os.Stderr, "Warning: live share is exposed on %s. Anyone with the URL can view the session.\n", shellShareBind)
	}

	hub := share.NewHub()
	go hub.Run()

	srv := share.NewServer(hub, token)
	listenAddr := fmt.Sprintf("%s:%d", shellShareBind, shellSharePort)
	boundAddr, err := srv.Start(listenAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not start share server: %v\n", err)
		hub.Stop()
		return nil, nil, nil
	}

	displayAddr := boundAddr
	if shellShareBind == "0.0.0.0" {
		if ip := getOutboundIP(); ip != "" {
			_, port, _ := net.SplitHostPort(boundAddr)
			displayAddr = net.JoinHostPort(ip, port)
		}
	}

	shareURL := fmt.Sprintf("http://%s/watch?token=%s", displayAddr, token)
	os.Setenv("PENTLOG_SHARE_URL", shareURL)

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
	go runTailBroadcast(ctx, hub, logFilePath)

	return hub, srv, cancel
}

func promptResumeSession(crashed []logs.Session) *logs.Session {
	if len(crashed) == 0 {
		return nil
	}

	fmt.Println()
	fmt.Println("⚠️  Found crashed session(s) for this context:")
	fmt.Println()

	for i, s := range crashed {
		lastSync := "unknown"
		if s.LastSyncAt != "" {
			if t, err := time.Parse(time.RFC3339, s.LastSyncAt); err == nil {
				lastSync = utils.FormatRelativeTime(t)
			}
		}
		fileSize := s.Size
		if info, err := os.Stat(s.Path); err == nil {
			fileSize = info.Size()
		}

		fmt.Printf("  [%d] Session ID: %d\n", i+1, s.ID)
		fmt.Printf("      File: %s (%s)\n", s.Filename, utils.FormatSize(fileSize))
		fmt.Printf("      Crashed: %s\n", lastSync)
		if i < len(crashed)-1 {
			fmt.Println()
		}
	}

	fmt.Println()

	prompt := promptui.Select{
		Label: "Resume crashed session or start new?",
		Items: []string{"Resume most recent", "Start new session"},
	}

	idx, _, err := prompt.Run()
	if err != nil || idx == 1 {
		return nil
	}

	return &crashed[0]
}

func startResumedSession(ctx *config.ContextData, session *logs.Session) {
	fmt.Printf("\n✓ Resuming session ID: %d\n", session.ID)
	fmt.Printf("  File: %s\n", session.Path)
	fmt.Println()

	if err := logs.ResumeSession(int64(session.ID)); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to update session state: %v\n", err)
	}

	markerFile := session.Path + ".resume_marker"
	if err := utils.WritePrivateFile(markerFile, []byte(fmt.Sprintf("%d", time.Now().Unix()))); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create resume marker: %v\n", err)
	}

	// Insert the "Session Resumed" banner into the tty file
	if err := logs.InsertResumeMarker(session.Path); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to insert resume marker: %v\n", err)
	}

	sessionDir := filepath.Dir(session.Path)

	newEnv, tempDir, shellArgs, err := prepareShellEnv(ctx, sessionDir, session.MetaPath, session.Path, int64(session.ID))
	if err != nil {
		errors.FromError(errors.Generic, "failed to prepare shell environment for resumed session", err).Fatal()
	}
	if tempDir != "" {
		defer os.RemoveAll(tempDir)
	}

	recorder := system.NewRecorder()
	c, err := recorder.BuildCommand("", session.Path, shellArgs...)
	if err != nil {
		errors.FromError(errors.Generic, "failed to create recorder command for resumed session", err).Fatal()
	}

	hbCtx, hbCancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// Start session size monitoring for resumed session
	cfgMgr := config.Manager()
	cfg := cfgMgr.GetMonitor()
	monitorConfig := logs.MonitorConfig{
		WarnThreshold:  int64(cfg.WarnThresholdMB) * 1024 * 1024,
		AlertThreshold: int64(cfg.AlertThresholdMB) * 1024 * 1024,
		CheckInterval:  time.Duration(cfg.CheckIntervalSec) * time.Second,
		AlertCooldown:  time.Duration(cfg.AlertCooldownMin) * time.Minute,
	}
	sessionMonitor := logs.NewSessionMonitor(session.Path, monitorConfig)
	sessionMonitor.Start()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runHeartbeat(hbCtx, int64(session.ID), session.Path)
	}()

	var shareCancel context.CancelFunc
	var shareSrv *share.Server
	var shareHub *share.Hub
	if shellShare {
		shareHub, shareSrv, shareCancel = startShareServer(session.Path)
		if shareCancel != nil {
			defer shareCancel()
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	var signalReceived bool
	var mu sync.Mutex

	go func() {
		sig, ok := <-sigChan
		if !ok || sig == nil {
			return
		}
		mu.Lock()
		signalReceived = true
		mu.Unlock()

		hbCancel()

		if c.Process != nil {
			if c.SysProcAttr != nil && c.SysProcAttr.Setpgid {
				syscall.Kill(-c.Process.Pid, sig.(syscall.Signal))
			} else {
				c.Process.Signal(sig)
			}
		}
	}()

	runErr := startRecording(c, newEnv, ctx)

	hbCancel()
	wg.Wait()
	signal.Stop(sigChan)
	close(sigChan)

	// Stop session monitoring
	sessionMonitor.Stop()

	if shareCancel != nil {
		shareCancel()
	}
	if shareSrv != nil {
		shareSrv.Stop()
	}
	if shareHub != nil {
		shareHub.Stop()
	}

	wasNormalExit := runErr == nil
	mu.Lock()
	if signalReceived {
		wasNormalExit = false
	}
	mu.Unlock()

	markerFile = session.Path + ".resume_marker"
	if _, statErr := os.Stat(markerFile); statErr == nil {
		if normErr := logs.NormalizeResumedSession(session.Path); normErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to normalize session timestamps: %v\n", normErr)
		}
		os.Remove(markerFile)
	}

	updateSessionOnExit(int64(session.ID), session.Path, wasNormalExit)

	if runErr != nil {
		fmt.Fprintf(os.Stderr, "Error running recorder: %v\n", runErr)
		return
	}

	exitMsg := "\nLeaving pentlog shell session."
	if shellShare {
		exitMsg += " (share server stopped)"
	}
	fmt.Println(exitMsg)
}

func init() {
	shellCmd.PersistentFlags().BoolVar(&shellShare, "share", false, "Enable live sharing via browser")
	shellCmd.PersistentFlags().IntVar(&shellSharePort, "share-port", 0, "Port for share server (0 = random)")
	shellCmd.PersistentFlags().StringVar(&shellShareBind, "share-bind", "127.0.0.1", "Bind address for share server")
	shellCmd.PersistentFlags().StringVar(&shellPhaseOverride, "phase", "", "Override the current phase for this shell session")
	shellCmd.PersistentFlags().StringVar(&shellTargetOverride, "target", "", "Override the current target for this shell session")
	shellCmd.Flags().BoolVar(&shellReview, "review", false, "Run shell review before starting")
	shellCmd.AddCommand(shellReviewCmd)
	rootCmd.AddCommand(shellCmd)
}

func writeMetadata(path string, meta logs.SessionMetadata) error {
	f, err := utils.CreatePrivateFile(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(meta)
}

func runHeartbeat(ctx context.Context, sessionID int64, logFilePath string) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if info, err := os.Stat(logFilePath); err == nil {
				logs.UpdateSessionSize(sessionID, info.Size())
			} else {
				logs.UpdateSessionHeartbeat(sessionID)
			}
		}
	}
}

func applyShellContextOverrides(mgr *config.ConfigManager, ctx *config.ContextData) (*config.ContextData, error) {
	updated := *ctx

	if shellPhaseOverride != "" && !strings.EqualFold(strings.TrimSpace(updated.Phase), strings.TrimSpace(shellPhaseOverride)) {
		updated.Phase = strings.TrimSpace(shellPhaseOverride)
	}

	if shellTargetOverride != "" && !strings.EqualFold(strings.TrimSpace(updated.Target), strings.TrimSpace(shellTargetOverride)) {
		targets, err := mgr.LoadTargets()
		if err != nil {
			return nil, fmt.Errorf("load targets: %w", err)
		}

		var matched *config.Target
		for _, target := range targets.Targets {
			if strings.EqualFold(target.Name, shellTargetOverride) {
				targetCopy := target
				matched = &targetCopy
				break
			}
		}
		if matched == nil {
			return nil, fmt.Errorf("target %q not found; run 'pentlog target list' to review available targets", shellTargetOverride)
		}

		updated.Target = matched.Name
		updated.TargetIP = matched.IP
	}

	return &updated, nil
}

func runShellPreflight(mgr *config.ConfigManager, saved *config.ContextData) (*config.ContextData, bool, error) {
	effective, err := applyShellContextOverrides(mgr, saved)
	if err != nil {
		return nil, false, err
	}

	plan, err := buildShellPreflightPlan(mgr, *saved, *effective)
	if err != nil {
		return nil, false, err
	}

	for {
		printShellPreflight(plan)

		action := selectShellPreflightAction(plan)
		switch action {
		case "start":
			if !confirmShellPreflightStart(plan, false) {
				continue
			}
			return &plan.Effective, true, nil
		case "persist_start":
			if !confirmShellPreflightStart(plan, true) {
				continue
			}
			plan.Effective.Timestamp = time.Now().Format(time.RFC3339)
			if err := mgr.SaveContext(&plan.Effective); err != nil {
				return nil, false, err
			}
			return &plan.Effective, true, nil
		case "edit_phase":
			phase := strings.TrimSpace(utils.PromptString("Shell Phase", plan.Effective.Phase))
			if phase == "" {
				continue
			}
			plan.Effective.Phase = phase
		case "select_target":
			target, ok := promptShellTarget(plan.Targets, plan.Effective.Target)
			if !ok {
				continue
			}
			plan.Effective.Target = target.Name
			plan.Effective.TargetIP = target.IP
		case "clear_target":
			plan.Effective.Target = ""
			plan.Effective.TargetIP = ""
		case "revert":
			plan.Effective = plan.Saved
		default:
			return nil, false, nil
		}

		plan, err = buildShellPreflightPlan(mgr, plan.Saved, plan.Effective)
		if err != nil {
			return nil, false, err
		}
	}
}

func buildShellPreflightPlan(mgr *config.ConfigManager, saved, effective config.ContextData) (*shellPreflightPlan, error) {
	targets, err := mgr.LoadTargets()
	if err != nil {
		return nil, fmt.Errorf("load targets: %w", err)
	}

	recentChanges, err := loadRecentContextChanges(mgr, 3)
	if err != nil {
		return nil, fmt.Errorf("load context history: %w", err)
	}

	return &shellPreflightPlan{
		Saved:         saved,
		Effective:     effective,
		Targets:       targets.Targets,
		RecentChanges: recentChanges,
		Warnings:      collectShellPreflightWarningsAt(saved, effective, targets.Targets, time.Now()),
		PendingChange: describePendingContextMutation(saved, effective),
	}, nil
}

func printShellPreflight(plan *shellPreflightPlan) {
	lines := []string{}

	if plan.Effective.Type == "Exam/Lab" {
		lines = append(lines, fmt.Sprintf("Exam/Lab Name: %s", plan.Effective.Client))
		lines = append(lines, fmt.Sprintf("Target:        %s", plan.Effective.Engagement))
	} else {
		lines = append(lines, fmt.Sprintf("Client:     %s", plan.Effective.Client))
		lines = append(lines, fmt.Sprintf("Engagement: %s", plan.Effective.Engagement))
	}
	lines = append(lines, fmt.Sprintf("Phase:      %s", plan.Effective.Phase))
	lines = append(lines, fmt.Sprintf("Target/IP:  %s", formatTargetDisplay(plan.Effective.Target, plan.Effective.TargetIP)))
	lines = append(lines, fmt.Sprintf("Context Age:%s", " "+formatContextAgeDetail(plan.Saved.Timestamp)))

	if plan.PendingChange != "" {
		lines = append(lines, fmt.Sprintf("Shell Fixes: %s", plan.PendingChange))
	}

	if len(plan.RecentChanges) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Recent changes:")
		for _, change := range plan.RecentChanges {
			lines = append(lines, "  - "+change)
		}
	}

	if len(plan.Warnings) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Guardrails:")
		for _, warning := range plan.Warnings {
			lines = append(lines, "  - "+warning)
		}
	}

	fmt.Println()
	utils.PrintBox("Shell Review", lines)
}

func selectShellPreflightAction(plan *shellPreflightPlan) string {
	actions := []struct {
		Label string
		Key   string
	}{
		{Label: "Start shell with this context", Key: "start"},
	}

	if plan.PendingChange != "" {
		actions = append(actions, struct {
			Label string
			Key   string
		}{Label: "Start shell and persist these context changes", Key: "persist_start"})
	}

	actions = append(actions, struct {
		Label string
		Key   string
	}{Label: "Edit shell phase", Key: "edit_phase"})

	if len(plan.Targets) > 0 {
		actions = append(actions, struct {
			Label string
			Key   string
		}{Label: "Select shell target", Key: "select_target"})
	}

	if strings.TrimSpace(plan.Effective.Target) != "" || strings.TrimSpace(plan.Effective.TargetIP) != "" {
		actions = append(actions, struct {
			Label string
			Key   string
		}{Label: "Clear shell target", Key: "clear_target"})
	}

	if plan.PendingChange != "" {
		actions = append(actions, struct {
			Label string
			Key   string
		}{Label: "Revert to saved context", Key: "revert"})
	}

	actions = append(actions, struct {
		Label string
		Key   string
	}{Label: "Cancel", Key: "cancel"})

	items := make([]string, 0, len(actions))
	for _, action := range actions {
		items = append(items, action.Label)
	}

	choice := utils.SelectItem("Shell review action", items)
	if choice == -1 {
		return "cancel"
	}
	return actions[choice].Key
}

func confirmShellPreflightStart(plan *shellPreflightPlan, persist bool) bool {
	if len(plan.Warnings) == 0 {
		return true
	}

	label := "Proceed despite these shell-context risks?"
	if persist {
		label = "Persist the context changes and proceed despite these risks?"
	}

	return utils.PromptSelect(label, []string{"Yes", "No"}) == "Yes"
}

func promptShellTarget(targets []config.Target, current string) (config.Target, bool) {
	items := make([]string, 0, len(targets))
	for _, target := range targets {
		label := target.Name
		if strings.TrimSpace(target.IP) != "" {
			label = fmt.Sprintf("%s (%s)", target.Name, target.IP)
		}
		if strings.EqualFold(strings.TrimSpace(target.Name), strings.TrimSpace(current)) {
			label += " [current]"
		}
		items = append(items, label)
	}

	choice := utils.SelectItem("Select shell target", items)
	if choice == -1 {
		return config.Target{}, false
	}

	return targets[choice], true
}

func updateSessionOnExit(sessionID int64, logFilePath string, normalExit bool) {
	stopShareIfActive()

	state := logs.SessionStateCompleted
	if !normalExit {
		state = logs.SessionStateCrashed
	}

	if err := logs.UpdateSessionState(sessionID, state); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to update session state: %v\n", err)
	}

	if info, err := os.Stat(logFilePath); err == nil {
		logs.UpdateSessionSize(sessionID, info.Size())
	}
}

func stopShareIfActive() {
	configMgr := config.Manager()
	shareSessionPath := filepath.Join(configMgr.GetPaths().Home, ".share_session")
	os.Remove(shareSessionPath)
}
