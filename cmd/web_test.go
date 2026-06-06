package cmd

import (
	"testing"

	"pentlog/pkg/httpauth"

	"github.com/spf13/cobra"
)

func TestResolveWebAuthConfigRejectsImplicitNonLoopbackExposure(t *testing.T) {
	cmd := &cobra.Command{Use: "web"}
	cmd.Flags().String("auth-token", "", "")
	cmd.Flags().String("basic-auth", "", "")

	webPublic = false
	webAuthToken = ""
	webBasicAuth = ""

	_, err := resolveWebAuthConfig(cmd, "0.0.0.0")
	if err == nil {
		t.Fatalf("expected non-loopback bind without auth/public to fail")
	}
}

func TestResolveWebAuthConfigGeneratesTokenWhenFlagPresentWithoutValue(t *testing.T) {
	cmd := &cobra.Command{Use: "web"}
	cmd.Flags().String("auth-token", "", "")
	cmd.Flags().String("basic-auth", "", "")
	if err := cmd.Flags().Set("auth-token", autoGenerateAuthValue); err != nil {
		t.Fatalf("set auth-token: %v", err)
	}

	webPublic = false
	webAuthToken = autoGenerateAuthValue
	webBasicAuth = ""

	cfg, err := resolveWebAuthConfig(cmd, "0.0.0.0")
	if err != nil {
		t.Fatalf("resolve auth config: %v", err)
	}
	if cfg.Mode != httpauth.ModeToken || cfg.Token == "" {
		t.Fatalf("expected generated token config, got %+v", cfg)
	}
}

func TestResolveWebAuthConfigGeneratesBasicCredentialsWhenFlagPresentWithoutValue(t *testing.T) {
	cmd := &cobra.Command{Use: "web"}
	cmd.Flags().String("auth-token", "", "")
	cmd.Flags().String("basic-auth", "", "")
	if err := cmd.Flags().Set("basic-auth", autoGenerateAuthValue); err != nil {
		t.Fatalf("set basic-auth: %v", err)
	}

	webPublic = false
	webAuthToken = ""
	webBasicAuth = autoGenerateAuthValue

	cfg, err := resolveWebAuthConfig(cmd, "0.0.0.0")
	if err != nil {
		t.Fatalf("resolve auth config: %v", err)
	}
	if cfg.Mode != httpauth.ModeBasic {
		t.Fatalf("expected basic auth mode, got %s", cfg.Mode)
	}
	if cfg.Username != "pentlog" || cfg.Password == "" {
		t.Fatalf("expected generated basic credentials, got %+v", cfg)
	}
}

func TestResolveWebAuthConfigRejectsConflictingModes(t *testing.T) {
	cmd := &cobra.Command{Use: "web"}
	cmd.Flags().String("auth-token", "", "")
	cmd.Flags().String("basic-auth", "", "")
	if err := cmd.Flags().Set("auth-token", "token"); err != nil {
		t.Fatalf("set auth-token: %v", err)
	}
	if err := cmd.Flags().Set("basic-auth", "user:pass"); err != nil {
		t.Fatalf("set basic-auth: %v", err)
	}

	webPublic = false
	webAuthToken = "token"
	webBasicAuth = "user:pass"

	_, err := resolveWebAuthConfig(cmd, "0.0.0.0")
	if err == nil {
		t.Fatalf("expected conflicting auth modes to fail")
	}
}
