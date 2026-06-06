package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"pentlog/pkg/config"
	"pentlog/pkg/httpauth"
)

func TestServerTokenAuthGatesAPIAndSetsCookie(t *testing.T) {
	server := NewServer("0.0.0.0", 8080, httpauth.TokenConfig("secret-token"))

	unauthReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	unauthRec := httptest.NewRecorder()
	server.Router.ServeHTTP(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated API request, got %d", unauthRec.Code)
	}

	authReq := httptest.NewRequest(http.MethodGet, "/api/health?auth_token=secret-token", nil)
	authRec := httptest.NewRecorder()
	server.Router.ServeHTTP(authRec, authReq)
	if authRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for token-authenticated request, got %d body=%s", authRec.Code, authRec.Body.String())
	}

	cookies := authRec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != httpauth.TokenCookieName {
		t.Fatalf("expected auth cookie to be set, got %+v", cookies)
	}

	cookieReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	cookieReq.AddCookie(cookies[0])
	cookieRec := httptest.NewRecorder()
	server.Router.ServeHTTP(cookieRec, cookieReq)
	if cookieRec.Code != http.StatusOK {
		t.Fatalf("expected cookie-authenticated request to pass, got %d", cookieRec.Code)
	}

	uiReq := httptest.NewRequest(http.MethodGet, "/", nil)
	uiRec := httptest.NewRecorder()
	server.Router.ServeHTTP(uiRec, uiReq)
	if uiRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected UI route to be gated, got %d", uiRec.Code)
	}
}

func TestServerTokenAuthScrubsTokenFromBrowserURL(t *testing.T) {
	server := NewServer("0.0.0.0", 8080, httpauth.TokenConfig("secret-token"))

	req := httptest.NewRequest(http.MethodGet, "/?auth_token=secret-token", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected browser bootstrap request to redirect, got %d", rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "/" {
		t.Fatalf("expected redirect to sanitized root URL, got %q", location)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != httpauth.TokenCookieName {
		t.Fatalf("expected auth cookie to be set on redirect, got %+v", cookies)
	}
}

func TestServerBasicAuthGatesAPI(t *testing.T) {
	server := NewServer("0.0.0.0", 8080, httpauth.BasicConfig("pentester", "secret"))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without basic auth, got %d", rec.Code)
	}

	authReq := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	authReq.SetBasicAuth("pentester", "secret")
	authRec := httptest.NewRecorder()
	server.Router.ServeHTTP(authRec, authReq)
	if authRec.Code != http.StatusOK {
		t.Fatalf("expected 200 with basic auth, got %d body=%s", authRec.Code, authRec.Body.String())
	}
}

func TestArtifactRoutesRemainTokenProtectedWhenRouteAuthIsEnabled(t *testing.T) {
	config.ResetManagerForTesting()
	defer config.ResetManagerForTesting()

	tmpDir := t.TempDir()
	os.Setenv("PENTLOG_TEST_HOME", tmpDir)
	defer os.Unsetenv("PENTLOG_TEST_HOME")

	mgr := config.Manager()
	if err := mgr.EnsureDirectories(); err != nil {
		t.Fatalf("ensure directories: %v", err)
	}

	reportDir := filepath.Join(mgr.GetPaths().ReportsDir, "Acme")
	if err := os.MkdirAll(reportDir, 0700); err != nil {
		t.Fatalf("mkdir report dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(reportDir, "report.html"), []byte("<html></html>"), 0600); err != nil {
		t.Fatalf("write report: %v", err)
	}

	server := NewServer("0.0.0.0", 8080, httpauth.TokenConfig("secret-token"))

	unauthReq := httptest.NewRequest(http.MethodGet, "/files/reports/Acme/report.html", nil)
	unauthRec := httptest.NewRecorder()
	server.Router.ServeHTTP(unauthRec, unauthReq)
	if unauthRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected artifact route without artifact token to fail, got %d", unauthRec.Code)
	}

	artifactURL := httpauth.BuildArtifactURL("/files/reports", "Acme", "report.html")
	authReq := httptest.NewRequest(http.MethodGet, artifactURL, nil)
	authRec := httptest.NewRecorder()
	server.Router.ServeHTTP(authRec, authReq)
	if authRec.Code != http.StatusOK {
		t.Fatalf("expected artifact token to authorize download, got %d body=%s", authRec.Code, authRec.Body.String())
	}
}
