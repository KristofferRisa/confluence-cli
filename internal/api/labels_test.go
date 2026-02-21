package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kristofferrisa/confluence-cli/internal/api"
	"github.com/kristofferrisa/confluence-cli/internal/models"
)

func TestGetLabels(t *testing.T) {
	want := models.LabelList{
		Results: []models.Label{
			{ID: "1", Prefix: "global", Name: "docs"},
			{ID: "2", Prefix: "global", Name: "api"},
		},
		Size:  2,
		Limit: 200,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/content/42/label") {
			t.Errorf("path = %q, want suffix /content/42/label", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	labels, err := c.GetLabels(context.Background(), "42")
	if err != nil {
		t.Fatalf("GetLabels() error: %v", err)
	}

	if len(labels) != 2 {
		t.Fatalf("len(labels) = %d, want 2", len(labels))
	}
	if labels[0].Name != "docs" {
		t.Errorf("labels[0].Name = %q, want docs", labels[0].Name)
	}
	if labels[1].Name != "api" {
		t.Errorf("labels[1].Name = %q, want api", labels[1].Name)
	}
}

func TestAddLabels(t *testing.T) {
	type labelReq struct {
		Prefix string `json:"prefix"`
		Name   string `json:"name"`
	}

	var gotBody []labelReq
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/content/10/label") {
			t.Errorf("path = %q, want suffix /content/10/label", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	err := c.AddLabels(context.Background(), "10", []string{"backend", "v2"})
	if err != nil {
		t.Fatalf("AddLabels() error: %v", err)
	}

	if len(gotBody) != 2 {
		t.Fatalf("request body len = %d, want 2", len(gotBody))
	}
	if gotBody[0].Prefix != "global" {
		t.Errorf("gotBody[0].Prefix = %q, want global", gotBody[0].Prefix)
	}
	if gotBody[0].Name != "backend" {
		t.Errorf("gotBody[0].Name = %q, want backend", gotBody[0].Name)
	}
	if gotBody[1].Name != "v2" {
		t.Errorf("gotBody[1].Name = %q, want v2", gotBody[1].Name)
	}
}

func TestRemoveLabel(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	if err := c.RemoveLabel(context.Background(), "10", "backend"); err != nil {
		t.Fatalf("RemoveLabel() error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if !strings.HasSuffix(gotPath, "/content/10/label/backend") {
		t.Errorf("path = %q, want suffix /content/10/label/backend", gotPath)
	}
}
