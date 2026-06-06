package httpauth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type Mode string

const (
	ModeNone  Mode = ""
	ModeToken Mode = "token"
	ModeBasic Mode = "basic"
)

const (
	QueryTokenParam = "auth_token"
	TokenHeaderName = "X-Pentlog-Token"
	TokenCookieName = "pentlog_auth"
)

type Config struct {
	Mode     Mode
	Token    string
	Username string
	Password string
}

func TokenConfig(token string) Config {
	return Config{
		Mode:  ModeToken,
		Token: token,
	}
}

func BasicConfig(username string, password string) Config {
	return Config{
		Mode:     ModeBasic,
		Username: username,
		Password: password,
	}
}

func (c Config) Enabled() bool {
	return c.Mode != ModeNone
}

func (c Config) AppendTokenQuery(rawURL string) string {
	if c.Mode != ModeToken || c.Token == "" {
		return rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	values := parsed.Query()
	values.Set(QueryTokenParam, c.Token)
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

func (c Config) IsAuthorized(r *http.Request) bool {
	switch c.Mode {
	case ModeToken:
		return subtleCompare(extractToken(r), c.Token)
	case ModeBasic:
		username, password, ok := r.BasicAuth()
		return ok && subtleCompare(username, c.Username) && subtleCompare(password, c.Password)
	default:
		return true
	}
}

func (c Config) TokenShouldSetCookie(r *http.Request) bool {
	if c.Mode != ModeToken {
		return false
	}

	token := strings.TrimSpace(r.URL.Query().Get(QueryTokenParam))
	if subtleCompare(token, c.Token) {
		return true
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") && subtleCompare(strings.TrimSpace(authHeader[7:]), c.Token) {
		return true
	}

	headerToken := strings.TrimSpace(r.Header.Get(TokenHeaderName))
	return subtleCompare(headerToken, c.Token)
}

func (c Config) TokenCookie() *http.Cookie {
	return &http.Cookie{
		Name:     TokenCookieName,
		Value:    c.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func (c Config) HasTokenQuery(r *http.Request) bool {
	if c.Mode != ModeToken {
		return false
	}
	return subtleCompare(strings.TrimSpace(r.URL.Query().Get(QueryTokenParam)), c.Token)
}

func (c Config) SanitizedTokenURL(r *http.Request) (string, bool) {
	if !c.HasTokenQuery(r) {
		return "", false
	}

	clone := *r.URL
	values := clone.Query()
	values.Del(QueryTokenParam)
	clone.RawQuery = values.Encode()
	return clone.String(), true
}

func ParseBasicAuthCredentials(raw string) (string, string, bool) {
	if raw == "" {
		return "", "", false
	}

	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	username := strings.TrimSpace(parts[0])
	password := parts[1]
	if username == "" || password == "" {
		return "", "", false
	}

	return username, password, true
}

func GenerateToken(numBytes int) (string, error) {
	buf := make([]byte, numBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func IsLoopbackBind(bind string) bool {
	host := strings.TrimSpace(bind)
	switch strings.ToLower(host) {
	case "", "localhost":
		return true
	}

	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func extractToken(r *http.Request) string {
	if cookie, err := r.Cookie(TokenCookieName); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	if token := strings.TrimSpace(r.URL.Query().Get(QueryTokenParam)); token != "" {
		return token
	}

	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	return strings.TrimSpace(r.Header.Get(TokenHeaderName))
}

func subtleCompare(a string, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
