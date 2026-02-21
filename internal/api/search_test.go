package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/kristofferrisa/confluence-cli/internal/api"
	"github.com/kristofferrisa/confluence-cli/internal/models"
)

func TestSearch(t *testing.T) {
	want := models.SearchResult{
		Results: []models.SearchEntry{
			{
				Title:   "Found Page",
				Excerpt: "some excerpt",
				Content: models.SearchContent{
					ID:    "77",
					Type:  "page",
					Title: "Found Page",
				},
			},
		},
		Size:      1,
		TotalSize: 1,
		Limit:     25,
	}

	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	cql := `type = "page" AND space.key = "ENG" AND title ~ "Found"`
	c := api.NewClient(srv.URL, "user@example.com", "token")

	result, err := c.Search(context.Background(), cql, &models.ListOptions{Limit: 25})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	if len(result.Results) != 1 {
		t.Errorf("len(Results) = %d, want 1", len(result.Results))
	}
	if result.Results[0].Title != "Found Page" {
		t.Errorf("Title = %q, want Found Page", result.Results[0].Title)
	}

	// Verify CQL was correctly URL-encoded in the query string.
	gotCQL := gotQuery.Get("cql")
	if gotCQL != cql {
		t.Errorf("cql query param = %q, want %q", gotCQL, cql)
	}

	if gotQuery.Get("limit") != "25" {
		t.Errorf("limit query param = %q, want 25", gotQuery.Get("limit"))
	}
}

func TestSearch_URLEncoding(t *testing.T) {
	// Verify that special characters in CQL are properly URL-encoded.
	var gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRawQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.SearchResult{})
	}))
	defer srv.Close()

	cql := `space = "MY SPACE" AND type = "page"`
	c := api.NewClient(srv.URL, "user@example.com", "token")
	_, err := c.Search(context.Background(), cql, nil)
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}

	// The raw query string must not contain unencoded spaces.
	if containsString(gotRawQuery, " ") {
		t.Errorf("raw query %q contains unencoded spaces", gotRawQuery)
	}

	// Decode and verify the cql value round-trips correctly.
	parsed, err := url.ParseQuery(gotRawQuery)
	if err != nil {
		t.Fatalf("parse raw query: %v", err)
	}
	if parsed.Get("cql") != cql {
		t.Errorf("decoded cql = %q, want %q", parsed.Get("cql"), cql)
	}
}
