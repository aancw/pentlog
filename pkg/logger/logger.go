package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"pentlog/pkg/utils"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

var (
	logger     *slog.Logger
	loggerOnce sync.Once
	output     io.Writer
	fileOutput *os.File
	level      Level
	mu         sync.RWMutex
)

func init() {
	SetOutput(os.Stderr)
	SetLevel(LevelInfo)
}

type outputType int

const (
	outputTypeText outputType = iota
	outputTypeJSON
)

var outType outputType = outputTypeText

func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	output = w
}

func SetLevel(l Level) {
	mu.Lock()
	defer mu.Unlock()
	level = l
}

func SetJSONOutput(enabled bool) {
	mu.Lock()
	defer mu.Unlock()
	if enabled {
		outType = outputTypeJSON
	} else {
		outType = outputTypeText
	}
}

func SetLogFile(path string) error {
	mu.Lock()
	defer mu.Unlock()

	if fileOutput != nil {
		fileOutput.Close()
	}

	f, err := utils.OpenPrivateFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND)
	if err != nil {
		return err
	}

	fileOutput = f
	output = io.MultiWriter(os.Stderr, f)
	loggerOnce = sync.Once{}

	return nil
}

func CloseLogFile() {
	mu.Lock()
	defer mu.Unlock()

	if fileOutput != nil {
		fileOutput.Close()
		fileOutput = nil
	}
	output = os.Stderr
	loggerOnce = sync.Once{}
}

func InitWithHome(homeDir string) error {
	if err := os.MkdirAll(homeDir, 0700); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(homeDir, "pentlog.log")
	if err := SetLogFile(logPath); err != nil {
		return fmt.Errorf("failed to set log file: %w", err)
	}

	Info("logging initialized", "file", logPath)
	return nil
}

func GetLogFilePath() string {
	mu.RLock()
	defer mu.RUnlock()
	if fileOutput != nil {
		return fileOutput.Name()
	}
	return ""
}

func getLogger() *slog.Logger {
	loggerOnce.Do(func() {
		mu.RLock()
		lvl := level
		w := output
		t := outType
		mu.RUnlock()

		var handler slog.Handler
		opts := &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.Level(lvl),
		}

		switch t {
		case outputTypeJSON:
			handler = slog.NewJSONHandler(w, opts)
		default:
			handler = slog.NewTextHandler(w, &slog.HandlerOptions{
				AddSource: true,
				Level:     slog.Level(lvl),
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.Attr{Key: "time", Value: slog.StringValue(a.Value.Time().Format("2006-01-02 15:04:05"))}
					}
					if a.Key == slog.MessageKey {
						return slog.Attr{Key: "msg", Value: a.Value}
					}
					return a
				},
			})
		}

		logger = slog.New(handler)
		slog.SetDefault(logger)
	})

	return logger
}

func Debug(msg string, args ...any) {
	getLogger().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	getLogger().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	getLogger().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	getLogger().Error(msg, args...)
}

func ErrorContext(_ context.Context, msg string, args ...any) {
	getLogger().Error(msg, args...)
}

func With(args ...any) *slog.Logger {
	return getLogger().With(args...)
}

func WithGroup(name string) *slog.Logger {
	return getLogger().WithGroup(name)
}

func Log(level Level, msg string, args ...any) {
	switch level {
	case LevelDebug:
		Debug(msg, args...)
	case LevelInfo:
		Info(msg, args...)
	case LevelWarn:
		Warn(msg, args...)
	case LevelError:
		Error(msg, args...)
	}
}

func LogIfErr(err error, msg string) {
	if err != nil {
		Error(msg, "error", err)
	}
}

func GetFile() *os.File {
	mu.RLock()
	defer mu.RUnlock()
	if f, ok := output.(*os.File); ok {
		return f
	}
	return os.Stderr
}

type ProgressLogger struct {
	mu      sync.Mutex
	enabled bool
}

func NewProgressLogger() *ProgressLogger {
	return &ProgressLogger{enabled: true}
}

func (p *ProgressLogger) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

func (p *ProgressLogger) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

func (p *ProgressLogger) Print(msg string, args ...any) {
	if !p.enabled {
		return
	}
	fmt.Fprintf(output, msg+"\n", args...)
}

func (p *ProgressLogger) Printf(format string, args ...any) {
	if !p.enabled {
		return
	}
	fmt.Fprintf(output, format+"\n", args...)
}

type UserOutput struct {
	mu   sync.Mutex
	out  io.Writer
	err  io.Writer
	json bool
}

func NewUserOutput() *UserOutput {
	return &UserOutput{
		out: os.Stdout,
		err: os.Stderr,
	}
}

func (u *UserOutput) Success(msg string, args ...any) {
	u.mu.Lock()
	defer u.mu.Unlock()
	fmt.Fprintf(u.out, "✓ "+msg+"\n", args...)
}

func (u *UserOutput) Info(msg string, args ...any) {
	u.mu.Lock()
	defer u.mu.Unlock()
	fmt.Fprintf(u.out, msg+"\n", args...)
}

func (u *UserOutput) Warning(msg string, args ...any) {
	u.mu.Lock()
	defer u.mu.Unlock()
	fmt.Fprintf(u.err, "⚠ "+msg+"\n", args...)
}

func (u *UserOutput) Error(msg string, args ...any) {
	u.mu.Lock()
	defer u.mu.Unlock()
	fmt.Fprintf(u.err, "✗ "+msg+"\n", args...)
}

func (u *UserOutput) Debug(msg string, args ...any) {
	if !isDebugEnabled() {
		return
	}
	u.mu.Lock()
	defer u.mu.Unlock()
	fmt.Fprintf(u.out, "▸ "+msg+"\n", args...)
}

func isDebugEnabled() bool {
	mu.RLock()
	defer mu.RUnlock()
	return level <= LevelDebug
}

var DefaultUserOutput = NewUserOutput()

func Success(msg string, args ...any) {
	DefaultUserOutput.Success(msg, args...)
}

func InfoMsg(msg string, args ...any) {
	DefaultUserOutput.Info(msg, args...)
}

func Warning(msg string, args ...any) {
	DefaultUserOutput.Warning(msg, args...)
}

func ErrorMsg(msg string, args ...any) {
	DefaultUserOutput.Error(msg, args...)
}

func DebugMsg(msg string, args ...any) {
	DefaultUserOutput.Debug(msg, args...)
}

type StructuredLogEntry struct {
	Time    time.Time
	Level   string
	Message string
	Source  string
	Args    map[string]interface{}
}

func ParseLogLine(line string) (*StructuredLogEntry, bool) {
	if !strings.HasPrefix(line, "{") {
		return nil, false
	}

	entry := &StructuredLogEntry{
		Args: make(map[string]interface{}),
	}

	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.HasPrefix(part, "time=") {
			entry.Time, _ = time.Parse("2006-01-02 15:04:05", strings.TrimPrefix(part, "time="))
		} else if strings.HasPrefix(part, "level=") {
			entry.Level = strings.TrimPrefix(part, "level=")
		} else if strings.HasPrefix(part, "msg=") {
			entry.Message = strings.Trim(strings.TrimPrefix(part, "msg="), `"`)
		} else if strings.HasPrefix(part, "src=") {
			entry.Source = strings.Trim(strings.TrimPrefix(part, "src="), `"`)
		}
	}

	return entry, entry.Message != ""
}

func Caller(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		return fmt.Sprintf("%s:%d", parts[len(parts)-2:][0], line)
	}
	return fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
}
