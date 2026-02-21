package api_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kristofferrisa/confluence-cli/internal/api"
)

func TestNewClient(t *testing.T) {
	c := api.NewClient("https://example.atlassian.net", "user@example.com", "secret-token")
	if c == nil {
		t.Fatal("NewClient() returned nil")
	}
}

func TestClient_AuthHeader(t *testing.T) {
	const (
		email = "user@example.com"
		token = "my-api-token"
	)

	// Track the Authorization header sent by the client.
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, email, token)

	// Perform any request to trigger header injection.
	var result map[string]string
	ctx := context.Background()
	if err := api.DoJSONExported(c, ctx, http.MethodGet, srv.URL, nil, &result); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	expected := "Basic " + base64.StdEncoding.EncodeToString([]byte(email+":"+token))
	if gotAuth != expected {
		t.Errorf("Authorization header = %q, want %q", gotAuth, expected)
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Full authentication is required to access this resource.",
		})
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "bad@example.com", "wrong-token")

	var result map[string]interface{}
	err := api.DoJSONExported(c, context.Background(), http.MethodGet, srv.URL, nil, &result)
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}

	want := "401"
	if !containsString(err.Error(), want) {
		t.Errorf("error %q does not contain %q", err.Error(), want)
	}

	wantMsg := "Full authentication is required"
	if !containsString(err.Error(), wantMsg) {
		t.Errorf("error %q does not contain %q", err.Error(), wantMsg)
	}
}

func TestClient_Timeout(t *testing.T) {
	// Hang indefinitely — the client should time out.
	hangCh := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-hangCh // block forever
	}))
	defer func() {
		close(hangCh)
		srv.Close()
	}()

	// Create a client with a very short timeout via a context deadline instead
	// of modifying the private httpClient timeout, so we don't need to export it.
	c := api.NewClient(srv.URL, "user@example.com", "token")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var result map[string]interface{}
	err := api.DoJSONExported(c, ctx, http.MethodGet, srv.URL, nil, &result)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// containsString reports whether s contains substr.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
