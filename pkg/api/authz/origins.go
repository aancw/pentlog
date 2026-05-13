package authz

import (
	"fmt"
	"net"
	"strings"
)

func AllowedOrigins(bind string, port int) []string {
	origins := []string{
		fmt.Sprintf("http://127.0.0.1:%d", port),
		fmt.Sprintf("http://localhost:%d", port),
	}

	host := strings.TrimSpace(bind)
	if host == "" || host == "0.0.0.0" || host == "::" {
		return origins
	}

	if ip := net.ParseIP(host); ip != nil {
		if !ip.IsLoopback() {
			origins = append(origins, fmt.Sprintf("http://%s:%d", host, port))
		}
		return origins
	}

	if !strings.EqualFold(host, "localhost") {
		origins = append(origins, fmt.Sprintf("http://%s:%d", host, port))
	}

	return origins
}
