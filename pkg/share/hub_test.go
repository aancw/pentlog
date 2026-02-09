package share

import (
	"testing"
	"time"
)

func TestHubClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	if hub.ClientCount() != 0 {
		t.Fatalf("expected 0 clients, got %d", hub.ClientCount())
	}
}

func TestHubBroadcastNonBlocking(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	done := make(chan struct{})
	go func() {
		hub.Broadcast([]byte("test"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Broadcast blocked with no clients")
	}
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) != 32 {
		t.Fatalf("expected 32 char hex token, got %d chars: %s", len(token), token)
	}

	token2, _ := GenerateToken()
	if token == token2 {
		t.Fatal("tokens should be unique")
	}
}
