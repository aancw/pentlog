package errors

import (
	"fmt"
	"os"
)

func NoContext() *Error {
	return NewError(NoActiveContext, "No active engagement context found")
}

func ContextMissing() *Error {
	return NewError(ContextNotFound, "Engagement context not found").
		WithDetails("Expected at ~/.pentlog/context.json")
}

func SessionMissing(sessionID string) *Error {
	return NewError(SessionNotFound, fmt.Sprintf("Session '%s' not found", sessionID)).
		WithDetails("Check if session file exists in ~/.pentlog/logs/")
}

func SessionCrashedError(sessionID string) *Error {
	return NewError(SessionCrashed, fmt.Sprintf("Session '%s' crashed unexpectedly", sessionID))
}

func AlreadyInShell() *Error {
	return NewError(AlreadyInSession, "Already inside a pentlog shell session")
}

func MissingDependency(tool, installCmd string) *Error {
	err := NewError(DependencyMissing, fmt.Sprintf("'%s' is not installed", tool)).
		WithDetails(fmt.Sprintf("Required for pentlog to function correctly"))
	if installCmd != "" {
		err.AddSolution(fmt.Sprintf("Install: %s", installCmd))
	}
	return err
}

func DatabaseErr(op string, err error) *Error {
	return FromError(DatabaseError, fmt.Sprintf("Database operation failed: %s", op), err)
}

func DatabaseLockedErr() *Error {
	return NewError(DatabaseLocked, "Database is locked")
}

func FileErr(path string, err error) *Error {
	if os.IsNotExist(err) {
		return FromError(FileNotFound, fmt.Sprintf("File not found: %s", path), err)
	}
	if os.IsPermission(err) {
		return FromError(PermissionDenied, fmt.Sprintf("Permission denied: %s", path), err)
	}
	return FromError(Generic, fmt.Sprintf("File error: %s", path), err)
}

func DirErr(path string, err error) *Error {
	if os.IsNotExist(err) {
		return FromError(DirectoryNotFound, fmt.Sprintf("Directory not found: %s", path), err)
	}
	if os.IsPermission(err) {
		return FromError(PermissionDenied, fmt.Sprintf("Permission denied: %s", path), err)
	}
	return FromError(Generic, fmt.Sprintf("Directory error: %s", path), err)
}

func AIConfigErr() *Error {
	return NewError(AIConfigMissing, "AI configuration not found").
		WithDetails("Expected at ~/.pentlog/ai_config.yaml")
}

func ArchiveErr(path string, err error) *Error {
	if os.IsNotExist(err) {
		return FromError(ArchiveNotFound, fmt.Sprintf("Archive not found: %s", path), err)
	}
	return FromError(ArchiveCorrupted, fmt.Sprintf("Archive corrupted: %s", path), err)
}

func ArchivePasswordErr(path string) *Error {
	return NewError(ArchiveEncrypted, fmt.Sprintf("Archive is password protected: %s", path))
}
