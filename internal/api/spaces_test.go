package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kristofferrisa/confluence-cli/internal/api"
	"github.com/kristofferrisa/confluence-cli/internal/models"
)

func TestListSpaces(t *testing.T) {
	want := models.SpaceList{
		Results: []models.Space{
			{ID: "1", Key: "TEAM", Name: "Team Space", Type: "global", Status: "current"},
			{ID: "2", Key: "PROJ", Name: "Project Space", Type: "global", Status: "current"},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !containsString(r.URL.Path, "/spaces") {
			t.Errorf("path %q does not contain /spaces", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	list, err := c.ListSpaces(context.Background(), &models.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListSpaces() error: %v", err)
	}

	if len(list.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2", len(list.Results))
	}
	if list.Results[0].Key != "TEAM" {
		t.Errorf("first space key = %q, want TEAM", list.Results[0].Key)
	}
	if list.Results[1].Key != "PROJ" {
		t.Errorf("second space key = %q, want PROJ", list.Results[1].Key)
	}
}

func TestGetSpaceByKey(t *testing.T) {
	want := models.SpaceList{
		Results: []models.Space{
			{ID: "42", Key: "ENG", Name: "Engineering", Type: "global", Status: "current"},
		},
	}

	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	space, err := c.GetSpaceByKey(context.Background(), "ENG")
	if err != nil {
		t.Fatalf("GetSpaceByKey() error: %v", err)
	}

	if space.Key != "ENG" {
		t.Errorf("Key = %q, want ENG", space.Key)
	}
	if space.ID != "42" {
		t.Errorf("ID = %q, want 42", space.ID)
	}

	if gotQuery != "keys=ENG" {
		t.Errorf("query = %q, want keys=ENG", gotQuery)
	}
}

func TestGetSpaceByKey_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models.SpaceList{Results: []models.Space{}})
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	_, err := c.GetSpaceByKey(context.Background(), "MISSING")
	if err == nil {
		t.Fatal("expected error for missing space, got nil")
	}

	if !containsString(err.Error(), "MISSING") {
		t.Errorf("error %q does not mention the key", err.Error())
	}
}
