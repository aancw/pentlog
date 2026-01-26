package errors

import (
	"fmt"
	"os"
	"strings"
)

type ErrorType int

const (
	NoActiveContext ErrorType = iota
	ContextNotFound
	InvalidContext
	SessionNotFound
	SessionCrashed
	AlreadyInSession
	DependencyMissing
	DependencyVersionMismatch
	DatabaseError
	DatabaseLocked
	FileNotFound
	DirectoryNotFound
	PermissionDenied
	AIConfigMissing
	AIConfigInvalid
	ArchiveNotFound
	ArchiveCorrupted
	ArchiveEncrypted
	Generic
)

type Error struct {
	Type        ErrorType
	Message     string
	Reasons     []string
	Solutions   []string
	Original    error
	Details     string
}

func NewError(errType ErrorType, message string) *Error {
	e := &Error{
		Type:      errType,
		Message:   message,
		Reasons:   []string{},
		Solutions: []string{},
	}
	e.setDefaults()
	return e
}

func FromError(errType ErrorType, message string, err error) *Error {
	e := NewError(errType, message)
	e.Original = err
	return e
}

func (e *Error) AddReason(reason string) *Error {
	e.Reasons = append(e.Reasons, reason)
	return e
}

func (e *Error) AddSolution(solution string) *Error {
	e.Solutions = append(e.Solutions, solution)
	return e
}

func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

func (e *Error) setDefaults() {
	switch e.Type {
	case NoActiveContext:
		e.Reasons = []string{
			"You haven't started an engagement yet",
			"The context was reset with 'pentlog reset'",
			"Running pentlog from a different user account",
		}
		e.Solutions = []string{
			"$ pentlog create   # Start a new engagement",
			"$ pentlog switch   # Switch to an existing context",
		}

	case ContextNotFound:
		e.Reasons = []string{
			"Context file not found at ~/.pentlog/context.json",
			"pentlog may not be initialized",
		}
		e.Solutions = []string{
			"$ pentlog setup    # Initialize pentlog",
			"$ pentlog create   # Create a new engagement",
		}

	case SessionNotFound:
		e.Reasons = []string{
			"Session file not found in logs directory",
			"Session may have been archived or deleted",
		}
		e.Solutions = []string{
			"$ pentlog sessions           # List available sessions",
			"$ pentlog archive --list     # Check archived sessions",
		}

	case SessionCrashed:
		e.Reasons = []string{
			"Shell session exited unexpectedly",
			"Recording may be incomplete",
		}
		e.Solutions = []string{
			"$ pentlog recover            # Recover crashed sessions",
			"$ pentlog sessions           # Check session status",
		}

	case AlreadyInSession:
		e.Reasons = []string{
			"You're already inside a pentlog shell session",
			"Nested sessions are not supported",
		}
		e.Solutions = []string{
			"Exit the current shell and try again",
			"$ exit   # Exit the pentlog shell",
		}

	case DependencyMissing:
		e.Reasons = []string{
			"Required tool is not installed on this system",
		}
		e.Solutions = []string{
			"$ pentlog setup    # Automatically install dependencies",
			"Install manually: brew install ttyrec ttyplay",
		}

	case DatabaseError:
		e.Reasons = []string{
			"SQLite database operation failed",
			"Database file may be corrupted",
		}
		e.Solutions = []string{
			"Try again in a moment",
			"$ pentlog setup    # Reinitialize database",
		}

	case DatabaseLocked:
		e.Reasons = []string{
			"Database is locked (another pentlog process is using it)",
		}
		e.Solutions = []string{
			"Wait for other pentlog operations to complete",
			"Close other pentlog sessions and try again",
		}

	case FileNotFound:
		e.Reasons = []string{
			"File doesn't exist at the specified path",
		}
		e.Solutions = []string{
			"Check the file path is correct",
			"Use 'pentlog sessions' to find session files",
		}

	case DirectoryNotFound:
		e.Reasons = []string{
			"Directory doesn't exist at the specified path",
		}
		e.Solutions = []string{
			"Create the directory first",
			"Check the path is correct",
		}

	case PermissionDenied:
		e.Reasons = []string{
			"You don't have permission to access this resource",
			"File/directory permissions may be restricted",
		}
		e.Solutions = []string{
			"Check file permissions: ls -la <path>",
			"Contact your system administrator if needed",
		}

	case AIConfigMissing:
		e.Reasons = []string{
			"AI configuration not found at ~/.pentlog/ai_config.yaml",
		}
		e.Solutions = []string{
			"$ pentlog analyze --setup       # Configure AI provider",
			"Supported: Google Gemini, Ollama",
		}

	case ArchiveNotFound:
		e.Reasons = []string{
			"Archive file not found at the specified path",
		}
		e.Solutions = []string{
			"Check the archive path is correct",
			"$ pentlog archive --list    # List existing archives",
		}

	case ArchiveCorrupted:
		e.Reasons = []string{
			"Archive file is corrupted or invalid",
		}
		e.Solutions = []string{
			"Try extracting with: unzip -t <archive>",
			"Recreate the archive if original is still available",
		}

	case ArchiveEncrypted:
		e.Reasons = []string{
			"Archive is password-protected",
		}
		e.Solutions = []string{
			"Use the password when extracting",
			"$ pentlog import <archive>    # Prompted for password",
		}
	}
}

func (e *Error) Error() string {
	return e.Format()
}

func (e *Error) Format() string {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("\n❌ Error: %s\n", e.Message))

	if e.Details != "" {
		buf.WriteString(fmt.Sprintf("\nDetails: %s\n", e.Details))
	}

	if e.Original != nil && e.Type != Generic {
		buf.WriteString(fmt.Sprintf("\nUnderlying: %v\n", e.Original))
	}

	if len(e.Reasons) > 0 {
		buf.WriteString("\nThis can happen when:\n")
		for i, reason := range e.Reasons {
			buf.WriteString(fmt.Sprintf("  %d. %s\n", i+1, reason))
		}
	}

	if len(e.Solutions) > 0 {
		buf.WriteString("\nTo fix:\n")
		for _, solution := range e.Solutions {
			if strings.HasPrefix(solution, "$") {
				buf.WriteString(fmt.Sprintf("  %s\n", solution))
			} else {
				buf.WriteString(fmt.Sprintf("  • %s\n", solution))
			}
		}
	}

	return buf.String()
}

func (e *Error) Print() {
	fmt.Fprint(os.Stderr, e.Format())
}

func (e *Error) Fatal() {
	e.Print()
	os.Exit(1)
}
