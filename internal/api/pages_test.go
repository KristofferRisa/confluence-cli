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

func TestCreatePage(t *testing.T) {
	want := models.Page{
		ID:      "12345",
		Title:   "My New Page",
		SpaceID: "SPACE1",
		Status:  "current",
	}

	var gotBody models.CreatePageRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/pages") {
			t.Errorf("path = %q, want suffix /pages", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	req := &models.CreatePageRequest{
		SpaceID: "SPACE1",
		Title:   "My New Page",
		Status:  "current",
		Body: models.CreatePageBody{
			Representation: "storage",
			Value:          "<p>Hello</p>",
		},
	}

	page, err := c.CreatePage(context.Background(), req)
	if err != nil {
		t.Fatalf("CreatePage() error: %v", err)
	}

	if page.ID != want.ID {
		t.Errorf("ID = %q, want %q", page.ID, want.ID)
	}
	if page.Title != want.Title {
		t.Errorf("Title = %q, want %q", page.Title, want.Title)
	}
	if gotBody.SpaceID != req.SpaceID {
		t.Errorf("request SpaceID = %q, want %q", gotBody.SpaceID, req.SpaceID)
	}
}

func TestGetPage(t *testing.T) {
	want := models.Page{
		ID:    "99",
		Title: "Existing Page",
		Body: &models.Body{
			Storage: &models.BodyContent{
				Value:          "<p>content</p>",
				Representation: "storage",
			},
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
	page, err := c.GetPage(context.Background(), "99")
	if err != nil {
		t.Fatalf("GetPage() error: %v", err)
	}

	if page.ID != want.ID {
		t.Errorf("ID = %q, want %q", page.ID, want.ID)
	}

	if gotQuery != "body-format=storage" {
		t.Errorf("query = %q, want body-format=storage", gotQuery)
	}
}

func TestUpdatePage(t *testing.T) {
	want := models.Page{
		ID:    "55",
		Title: "Updated Title",
		Version: &models.Version{
			Number: 2,
		},
	}

	var gotBody models.UpdatePageRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %q, want PUT", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	req := &models.UpdatePageRequest{
		ID:     "55",
		Title:  "Updated Title",
		Status: "current",
		Body: models.CreatePageBody{
			Representation: "storage",
			Value:          "<p>new content</p>",
		},
		Version: models.UpdateVersion{Number: 2},
	}

	page, err := c.UpdatePage(context.Background(), "55", req)
	if err != nil {
		t.Fatalf("UpdatePage() error: %v", err)
	}

	if page.Title != want.Title {
		t.Errorf("Title = %q, want %q", page.Title, want.Title)
	}
	if gotBody.Version.Number != 2 {
		t.Errorf("request Version.Number = %d, want 2", gotBody.Version.Number)
	}
}

func TestDeletePage(t *testing.T) {
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	if err := c.DeletePage(context.Background(), "42"); err != nil {
		t.Fatalf("DeletePage() error: %v", err)
	}

	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
}

func TestListPages(t *testing.T) {
	want := models.PageList{
		Results: []models.Page{
			{ID: "1", Title: "Page One"},
			{ID: "2", Title: "Page Two"},
		},
		Links: &models.Links{Next: "/wiki/api/v2/spaces/SPACE1/pages?cursor=abc"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	list, err := c.ListPages(context.Background(), "SPACE1", &models.ListOptions{Limit: 25})
	if err != nil {
		t.Fatalf("ListPages() error: %v", err)
	}

	if len(list.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2", len(list.Results))
	}
	if list.Links == nil || list.Links.Next == "" {
		t.Error("expected pagination next link")
	}
}

func TestCreatePage_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Title is required.",
		})
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	_, err := c.CreatePage(context.Background(), &models.CreatePageRequest{SpaceID: "X"})
	if err == nil {
		t.Fatal("expected error for 400 response, got nil")
	}

	if !containsString(err.Error(), "400") {
		t.Errorf("error %q does not contain 400", err.Error())
	}
	if !containsString(err.Error(), "Title is required") {
		t.Errorf("error %q does not contain expected message", err.Error())
	}
}
