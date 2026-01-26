package errors

import (
	"strings"
	"testing"
)

func TestNoContextError(t *testing.T) {
	err := NoContext()

	if err.Type != NoActiveContext {
		t.Errorf("Expected NoActiveContext, got %v", err.Type)
	}

	output := err.Format()
	if !strings.Contains(output, "No active engagement context found") {
		t.Errorf("Output missing main message: %s", output)
	}

	if !strings.Contains(output, "haven't started") {
		t.Errorf("Output missing reason: %s", output)
	}

	if !strings.Contains(output, "pentlog create") {
		t.Errorf("Output missing solution")
	}
}

func TestAlreadyInShellError(t *testing.T) {
	err := AlreadyInShell()

	if err.Type != AlreadyInSession {
		t.Errorf("Expected AlreadyInSession, got %v", err.Type)
	}

	output := err.Format()
	if !strings.Contains(output, "already in") {
		t.Errorf("Output missing context: %s", output)
	}
}

func TestMissingDependencyError(t *testing.T) {
	err := MissingDependency("ttyrec", "brew install ttyrec")

	if err.Type != DependencyMissing {
		t.Errorf("Expected DependencyMissing, got %v", err.Type)
	}

	output := err.Format()
	if !strings.Contains(output, "ttyrec") {
		t.Errorf("Output missing tool name: %s", output)
	}

	if !strings.Contains(output, "brew install ttyrec") {
		t.Errorf("Output missing install command: %s", output)
	}
}

func TestCustomError(t *testing.T) {
	err := NewError(Generic, "Something went wrong").
		AddReason("First reason").
		AddReason("Second reason").
		AddSolution("Try this fix").
		AddSolution("Or try this")

	output := err.Format()
	if !strings.Contains(output, "Something went wrong") {
		t.Errorf("Output missing message: %s", output)
	}

	if !strings.Contains(output, "First reason") || !strings.Contains(output, "Second reason") {
		t.Errorf("Output missing reasons: %s", output)
	}

	if !strings.Contains(output, "Try this fix") || !strings.Contains(output, "Or try this") {
		t.Errorf("Output missing solutions: %s", output)
	}
}

func TestErrorDetails(t *testing.T) {
	err := NewError(FileNotFound, "File missing").
		WithDetails("Expected file at /path/to/file")

	output := err.Format()
	if !strings.Contains(output, "Details: Expected file at /path/to/file") {
		t.Errorf("Output missing details: %s", output)
	}
}

func TestErrorFormatting(t *testing.T) {
	err := NoContext()
	output := err.Format()

	if !strings.Contains(output, "❌ Error") {
		t.Errorf("Output missing error marker: %s", output)
	}

	if !strings.Contains(output, "This can happen when:") {
		t.Errorf("Output missing reasons section: %s", output)
	}

	if !strings.Contains(output, "To fix:") {
		t.Errorf("Output missing solutions section: %s", output)
	}
}

func TestDatabaseErrorWithOriginal(t *testing.T) {
	originalErr := NewError(Generic, "original error")
	err := DatabaseErr("insert_session", originalErr)

	if err.Type != DatabaseError {
		t.Errorf("Expected DatabaseError, got %v", err.Type)
	}

	if err.Original == nil {
		t.Errorf("Original error should be set")
	}
}

func TestArchiveErrors(t *testing.T) {
	tests := []struct {
		name      string
		fn        func() *Error
		wantType  ErrorType
		wantInMsg string
	}{
		{
			name:      "ArchiveNotFound",
			fn:        func() *Error { return NewError(ArchiveNotFound, "Archive not found: test.zip") },
			wantType:  ArchiveNotFound,
			wantInMsg: "pentlog archive --list",
		},
		{
			name:      "ArchiveEncrypted",
			fn:        func() *Error { return ArchivePasswordErr("test.zip") },
			wantType:  ArchiveEncrypted,
			wantInMsg: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err.Type != tt.wantType {
				t.Errorf("Expected %v, got %v", tt.wantType, err.Type)
			}

			output := err.Format()
			if !strings.Contains(strings.ToLower(output), strings.ToLower(tt.wantInMsg)) {
				t.Errorf("Output missing expected text '%s': %s", tt.wantInMsg, output)
			}
		})
	}
}
