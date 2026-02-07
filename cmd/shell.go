package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"pentlog/pkg/config"
	"pentlog/pkg/deps"
	"pentlog/pkg/errors"
	"pentlog/pkg/logs"
	"pentlog/pkg/system"
	"pentlog/pkg/utils"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

const heartbeatInterval = 30 * time.Second

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Start a recorded shell with the engagement context loaded",
	Run: func(cmd *cobra.Command, args []string) {
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
			errors.NoContext().Print()
			os.Exit(1)
		}

		logDir, err := system.EnsureLogDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error preparing log dir: %v\n", err)
			os.Exit(1)
		}

		sessionDir := getSessionDir(logDir, ctx)
		if err := os.MkdirAll(sessionDir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating session dir: %v\n", err)
			os.Exit(1)
		}

		timestamp := time.Now().Format("20060102-150405")
		baseName := fmt.Sprintf("session-%s-%s", utils.Slugify(ctx.Operator), timestamp)
		logFilePath := filepath.Join(sessionDir, baseName+".tty")
		metaFilePath := filepath.Join(sessionDir, baseName+".json")

		meta := logs.SessionMetadata{
			Client:     ctx.Client,
			Engagement: ctx.Engagement,
			Scope:      ctx.Scope,
			Operator:   ctx.Operator,
			Phase:      ctx.Phase,
			Timestamp:  time.Now().Format(time.RFC3339),
		}

		if err := writeMetadata(metaFilePath, meta); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing metadata: %v\n", err)
			os.Exit(1)
		}

		sessionID, err := logs.AddSessionToDB(meta, logFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to add session to DB: %v\n", err)
		}

		newEnv, tempDir, shellArgs, err := prepareShellEnv(ctx, sessionDir, metaFilePath, logFilePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error preparing shell environment: %v\n", err)
			os.Exit(1)
		}
		if tempDir != "" {
			defer os.RemoveAll(tempDir)
		}

		recorder := system.NewRecorder()
		c, err := recorder.BuildCommand("", logFilePath, shellArgs...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating recorder command: %v\n", err)
			os.Exit(1)
		}

		hbCtx, hbCancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		if sessionID > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				runHeartbeat(hbCtx, sessionID, logFilePath)
			}()
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		
		var signalReceived bool
		var mu sync.Mutex
		
		go func() {
			sig := <-sigChan
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

		fmt.Println("\nLeaving pentlog shell session.")
	},
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

func prepareShellEnv(ctx *config.ContextData, sessionDir, metaFilePath, logFilePath string) ([]string, string, []string, error) {
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

	promptSegment := fmt.Sprintf("(pentlog:%s/%s)", ctx.Client, ctx.Phase)
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

# Zsh widget bindings for Ctrl+N (note) and Ctrl+G (vuln)
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
bindkey '^N' _pentlog_quicknote_widget
bindkey '^G' _pentlog_quickvuln_widget
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

# Bash keybindings for Ctrl+N (note) and Ctrl+G (vuln)
bind '"\C-n": "\C-apentlog quicknote\C-m"'
bind '"\C-g": "\C-apentlog quickvuln\C-m"'
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

	summary := []string{}
	if ctx.Type == "Exam/Lab" {
		summary = append(summary, fmt.Sprintf("Exam/Lab Name: %s", ctx.Client))
		summary = append(summary, fmt.Sprintf("Target:        %s", ctx.Engagement))
	} else {
		summary = append(summary, fmt.Sprintf("Client:     %s", ctx.Client))
		summary = append(summary, fmt.Sprintf("Engagement: %s", ctx.Engagement))
		summary = append(summary, fmt.Sprintf("Scope:      %s", ctx.Scope))
	}
	summary = append(summary, fmt.Sprintf("Operator:   %s", ctx.Operator))
	summary = append(summary, fmt.Sprintf("Phase:      %s", ctx.Phase))
	utils.PrintBox("Active Session", summary)

	fmt.Println()
	fmt.Println("⚠️  WARNING: All input (including passwords) is logged.")
	fmt.Println()
	utils.PrintCenteredBlock([]string{
		"Type 'exit' or Ctrl+D to stop recording.",
		"Hotkeys: Ctrl+N = Quick Note | Ctrl+G = Quick Vuln",
	})

	if err := c.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0 {
				fmt.Println("\nLeaving pentlog shell session.")
				return nil // Expected exit
			}
		}
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(shellCmd)
}

func writeMetadata(path string, meta logs.SessionMetadata) error {
	f, err := os.Create(path)
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

func updateSessionOnExit(sessionID int64, logFilePath string, normalExit bool) {
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
