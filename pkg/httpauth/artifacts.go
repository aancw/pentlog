package httpauth

import (
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
)

const artifactTokenQueryParam = "access_token"

var artifactTokenState struct {
	once  sync.Once
	token string
	err   error
}

func ArtifactToken() (string, error) {
	artifactTokenState.once.Do(func() {
		token, err := GenerateToken(32)
		if err != nil {
			artifactTokenState.err = err
			return
		}
		artifactTokenState.token = token
	})

	return artifactTokenState.token, artifactTokenState.err
}

func BuildArtifactURL(prefix string, segments ...string) string {
	token, err := ArtifactToken()
	if err != nil {
		return prefix
	}

	escaped := make([]string, 0, len(segments))
	for _, segment := range segments {
		escaped = append(escaped, url.PathEscape(segment))
	}

	base := strings.TrimRight(prefix, "/")
	if len(escaped) > 0 {
		base += "/" + strings.Join(escaped, "/")
	}

	return base + "?" + artifactTokenQueryParam + "=" + url.QueryEscape(token)
}

func ValidateArtifactRequest(r *http.Request) bool {
	token, err := ArtifactToken()
	if err != nil || token == "" {
		return false
	}

	return r.URL.Query().Get(artifactTokenQueryParam) == token
}

func JoinURLPath(prefix string, parts ...string) string {
	segments := append([]string{prefix}, parts...)
	return path.Join(segments...)
}
