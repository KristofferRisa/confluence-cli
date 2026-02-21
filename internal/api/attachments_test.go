package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kristofferrisa/confluence-cli/internal/api"
	"github.com/kristofferrisa/confluence-cli/internal/models"
)

func TestListAttachments(t *testing.T) {
	want := models.AttachmentList{
		Results: []models.Attachment{
			{
				ID:    "att1",
				Type:  "attachment",
				Title: "diagram.png",
				Links: &models.AttachmentLinks{
					Download: "/wiki/download/attachments/42/diagram.png",
				},
			},
			{
				ID:    "att2",
				Type:  "attachment",
				Title: "spec.pdf",
				Links: &models.AttachmentLinks{
					Download: "/wiki/download/attachments/42/spec.pdf",
				},
			},
		},
		Size: 2,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/content/42/child/attachment") {
			t.Errorf("path = %q, want suffix /content/42/child/attachment", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	attachments, err := c.ListAttachments(context.Background(), "42")
	if err != nil {
		t.Fatalf("ListAttachments() error: %v", err)
	}

	if len(attachments) != 2 {
		t.Fatalf("len(attachments) = %d, want 2", len(attachments))
	}
	if attachments[0].Title != "diagram.png" {
		t.Errorf("attachments[0].Title = %q, want diagram.png", attachments[0].Title)
	}
}

func TestUploadAttachment(t *testing.T) {
	wantAttachment := models.Attachment{
		ID:    "att99",
		Type:  "attachment",
		Title: "notes.txt",
		Links: &models.AttachmentLinks{
			Download: "/wiki/download/attachments/10/notes.txt",
		},
	}
	response := models.AttachmentList{
		Results: []models.Attachment{wantAttachment},
		Size:    1,
	}

	var (
		gotAtlassianToken string
		gotContentType    string
		gotFileContent    string
		gotFileName       string
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/content/10/child/attachment") {
			t.Errorf("path = %q, want suffix /content/10/child/attachment", r.URL.Path)
		}

		gotAtlassianToken = r.Header.Get("X-Atlassian-Token")
		gotContentType = r.Header.Get("Content-Type")

		// Parse the multipart body to verify file field name and content.
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
			t.Errorf("Content-Type is not multipart: %q", r.Header.Get("Content-Type"))
		} else {
			mr := multipart.NewReader(r.Body, params["boundary"])
			for {
				part, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Errorf("read multipart part: %v", err)
					break
				}
				if part.FormName() == "file" {
					gotFileName = part.FileName()
					data, _ := io.ReadAll(part)
					gotFileContent = string(data)
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	fileContent := "these are my notes"
	c := api.NewClient(srv.URL, "user@example.com", "token")
	att, err := c.UploadAttachment(context.Background(), "10", "notes.txt", strings.NewReader(fileContent))
	if err != nil {
		t.Fatalf("UploadAttachment() error: %v", err)
	}

	if att.ID != wantAttachment.ID {
		t.Errorf("attachment ID = %q, want %q", att.ID, wantAttachment.ID)
	}

	// Verify X-Atlassian-Token header.
	if gotAtlassianToken != "nocheck" {
		t.Errorf("X-Atlassian-Token = %q, want nocheck", gotAtlassianToken)
	}

	// Verify Content-Type is multipart.
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Errorf("Content-Type = %q, want multipart/form-data", gotContentType)
	}

	// Verify file name and content.
	if gotFileName != "notes.txt" {
		t.Errorf("file name = %q, want notes.txt", gotFileName)
	}
	if gotFileContent != fileContent {
		t.Errorf("file content = %q, want %q", gotFileContent, fileContent)
	}
}

func TestDownloadAttachment(t *testing.T) {
	fileData := []byte("binary file content here")

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fileData)
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")

	downloadPath := "/wiki/download/attachments/42/diagram.png"
	var buf bytes.Buffer
	if err := c.DownloadAttachment(context.Background(), downloadPath, &buf); err != nil {
		t.Fatalf("DownloadAttachment() error: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), fileData) {
		t.Errorf("downloaded content = %q, want %q", buf.Bytes(), fileData)
	}
	if gotPath != downloadPath {
		t.Errorf("request path = %q, want %q", gotPath, downloadPath)
	}
}

func TestDownloadAttachment_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Attachment not found"}`))
	}))
	defer srv.Close()

	c := api.NewClient(srv.URL, "user@example.com", "token")
	var buf bytes.Buffer
	err := c.DownloadAttachment(context.Background(), "/wiki/download/missing.png", &buf)
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if !containsString(err.Error(), "404") {
		t.Errorf("error %q does not contain 404", err.Error())
	}
}
