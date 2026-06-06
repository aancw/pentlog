package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidatePasswordInputFlagsRejectsMixedSources(t *testing.T) {
	if err := validatePasswordInputFlags("hunter2", true); err == nil {
		t.Fatal("expected mixed password sources to fail")
	}
}

func TestReadPasswordFromStdin(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetIn(strings.NewReader("hunter2\n"))

	password, err := readPasswordFromStdin(cmd)
	if err != nil {
		t.Fatalf("readPasswordFromStdin returned error: %v", err)
	}
	if password != "hunter2" {
		t.Fatalf("expected password to be trimmed to one line, got %q", password)
	}
}
