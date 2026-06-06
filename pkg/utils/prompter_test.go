package utils

import (
	"strings"
	"testing"
)

func TestReadSecretFromReaderTrimsTrailingNewlinesOnly(t *testing.T) {
	secret, err := ReadSecretFromReader(strings.NewReader("  hunter2  \n"))
	if err != nil {
		t.Fatalf("ReadSecretFromReader returned error: %v", err)
	}
	if secret != "  hunter2  " {
		t.Fatalf("expected surrounding spaces to be preserved, got %q", secret)
	}
}

func TestReadSecretFromReaderRejectsEmptyInput(t *testing.T) {
	if _, err := ReadSecretFromReader(strings.NewReader("\n")); err == nil {
		t.Fatal("expected empty stdin secret to fail")
	}
}
