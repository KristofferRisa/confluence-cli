package output_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kristofferrisa/confluence-cli/internal/models"
	"github.com/kristofferrisa/confluence-cli/internal/output"
)

// --------------------------------------------------------------------------
// sample data helpers
// --------------------------------------------------------------------------

func samplePage() *models.Page {
	return &models.Page{
		ID:        "123456",
		Status:    "current",
		Title:     "Getting Started",
		SpaceID:   "SPACE001",
		ParentID:  "100000",
		AuthorID:  "user-abc",
		CreatedAt: "2024-01-15T10:00:00.000Z",
		Version: &models.Version{
			Number:    3,
			CreatedAt: "2024-06-01T08:30:00.000Z",
		},
		Body: &models.Body{
			Storage: &models.BodyContent{
				Value:          "<p>Hello world</p>",
				Representation: "storage",
			},
		},
		Links: &models.Links{
			WebUI: "/wiki/spaces/SPACE001/pages/123456",
		},
	}
}

func samplePages() []models.Page {
	return []models.Page{
		{
			ID:      "111",
			Status:  "current",
			Title:   "Page One",
			SpaceID: "SPACE001",
		},
		{
			ID:      "222",
			Status:  "draft",
			Title:   "Page Two (Draft)",
			SpaceID: "SPACE001",
		},
	}
}

func sampleSpace() *models.Space {
	return &models.Space{
		ID:         "SPACE001",
		Key:        "ENG",
		Name:       "Engineering",
		Type:       "global",
		Status:     "current",
		HomepageID: "100000",
		CreatedAt:  "2023-03-10T09:00:00.000Z",
		Description: &models.SpaceDescription{
			Plain: &models.BodyContent{
				Value:          "The engineering team space.",
				Representation: "plain",
			},
		},
		Links: &models.Links{
			WebUI: "/wiki/spaces/ENG",
		},
	}
}

func sampleSpaces() []models.Space {
	return []models.Space{
		{ID: "S1", Key: "ENG", Name: "Engineering", Type: "global", Status: "current"},
		{ID: "S2", Key: "HR", Name: "Human Resources", Type: "global", Status: "current"},
	}
}

func sampleSearchResult() *models.SearchResult {
	return &models.SearchResult{
		TotalSize: 2,
		Size:      2,
		Limit:     25,
		Start:     0,
		Results: []models.SearchEntry{
			{
				Title:   "Getting Started",
				Excerpt: "This page explains how to get started with the platform.",
				URL:     "https://example.atlassian.net/wiki/spaces/ENG/pages/123",
				Content: models.SearchContent{
					ID:    "123",
					Type:  "page",
					Title: "Getting Started",
					Space: &models.SearchSpace{Key: "ENG", Name: "Engineering"},
					Links: &models.Links{WebUI: "/wiki/spaces/ENG/pages/123"},
				},
				LastModified: "2024-05-01",
			},
			{
				Title:   "API Reference",
				Excerpt: "Complete API reference documentation for all endpoints.",
				URL:     "https://example.atlassian.net/wiki/spaces/ENG/pages/456",
				Content: models.SearchContent{
					ID:    "456",
					Type:  "page",
					Title: "API Reference",
					Space: &models.SearchSpace{Key: "ENG", Name: "Engineering"},
				},
				LastModified: "2024-06-15",
			},
		},
	}
}

func sampleLabels() []models.Label {
	return []models.Label{
		{ID: "l1", Prefix: "global", Name: "getting-started"},
		{ID: "l2", Prefix: "global", Name: "api"},
		{ID: "l3", Prefix: "global", Name: "documentation"},
	}
}

func sampleAttachments() []models.Attachment {
	return []models.Attachment{
		{
			ID:    "att-1",
			Type:  "attachment",
			Title: "diagram.png",
			Metadata: models.AttachmentMeta{
				MediaType: "image/png",
				Comment:   "Architecture diagram",
			},
			Extensions: models.AttachmentExt{
				MediaType: "image/png",
				FileSize:  204800, // 200 KB
			},
		},
		{
			ID:    "att-2",
			Type:  "attachment",
			Title: "report.pdf",
			Metadata: models.AttachmentMeta{
				MediaType: "application/pdf",
			},
			Extensions: models.AttachmentExt{
				MediaType: "application/pdf",
				FileSize:  2097152, // 2 MB
			},
		},
	}
}

func samplePageTree() *models.PageTree {
	return &models.PageTree{
		Page: models.Page{
			ID:    "100",
			Title: "Root Page",
			Links: &models.Links{WebUI: "/wiki/spaces/ENG/pages/100"},
		},
		Children: []models.PageTree{
			{
				Page: models.Page{
					ID:    "101",
					Title: "Child One",
					Links: &models.Links{WebUI: "/wiki/spaces/ENG/pages/101"},
				},
				Children: []models.PageTree{
					{
						Page: models.Page{
							ID:    "102",
							Title: "Grandchild",
							Links: &models.Links{WebUI: "/wiki/spaces/ENG/pages/102"},
						},
					},
				},
			},
			{
				Page: models.Page{
					ID:    "103",
					Title: "Child Two",
					Links: &models.Links{WebUI: "/wiki/spaces/ENG/pages/103"},
				},
			},
		},
	}
}

// --------------------------------------------------------------------------
// TestNew — factory tests
// --------------------------------------------------------------------------

func TestNew_JSON(t *testing.T) {
	f := output.New("json")
	if _, ok := f.(*output.JSONFormatter); !ok {
		t.Fatalf("expected *JSONFormatter, got %T", f)
	}
}

func TestNew_Markdown(t *testing.T) {
	f := output.New("markdown")
	if _, ok := f.(*output.MarkdownFormatter); !ok {
		t.Fatalf("expected *MarkdownFormatter for 'markdown', got %T", f)
	}
}

func TestNew_MarkdownAlias(t *testing.T) {
	f := output.New("md")
	if _, ok := f.(*output.MarkdownFormatter); !ok {
		t.Fatalf("expected *MarkdownFormatter for 'md', got %T", f)
	}
}

func TestNew_Pretty(t *testing.T) {
	f := output.New("pretty")
	if _, ok := f.(*output.PrettyFormatter); !ok {
		t.Fatalf("expected *PrettyFormatter for 'pretty', got %T", f)
	}
}

func TestNew_Default(t *testing.T) {
	f := output.New("")
	if _, ok := f.(*output.PrettyFormatter); !ok {
		t.Fatalf("expected *PrettyFormatter for empty string, got %T", f)
	}
}

func TestNew_Unknown(t *testing.T) {
	f := output.New("xml")
	if _, ok := f.(*output.PrettyFormatter); !ok {
		t.Fatalf("expected *PrettyFormatter for unknown format, got %T", f)
	}
}

// --------------------------------------------------------------------------
// JSONFormatter tests
// --------------------------------------------------------------------------

func TestJSONFormatter_FormatPage(t *testing.T) {
	f := &output.JSONFormatter{}
	result := f.FormatPage(samplePage())

	if !json.Valid([]byte(result)) {
		t.Fatalf("FormatPage returned invalid JSON:\n%s", result)
	}

	var page models.Page
	if err := json.Unmarshal([]byte(result), &page); err != nil {
		t.Fatalf("failed to unmarshal page JSON: %v", err)
	}
	if page.ID != "123456" {
		t.Errorf("expected ID 123456, got %s", page.ID)
	}
	if page.Title != "Getting Started" {
		t.Errorf("expected title 'Getting Started', got %s", page.Title)
	}
}

func TestJSONFormatter_FormatPage_Nil(t *testing.T) {
	f := &output.JSONFormatter{}
	result := f.FormatPage(nil)
	if !json.Valid([]byte(result)) {
		t.Fatalf("FormatPage(nil) returned invalid JSON: %s", result)
	}
}

func TestJSONFormatter_FormatPages(t *testing.T) {
	f := &output.JSONFormatter{}
	result := f.FormatPages(samplePages())

	if !json.Valid([]byte(result)) {
		t.Fatalf("FormatPages returned invalid JSON:\n%s", result)
	}

	var pages []models.Page
	if err := json.Unmarshal([]byte(result), &pages); err != nil {
		t.Fatalf("failed to unmarshal pages JSON: %v", err)
	}
	if len(pages) != 2 {
		t.Errorf("expected 2 pages, got %d", len(pages))
	}
}

func TestJSONFormatter_FormatPages_Empty(t *testing.T) {
	f := &output.JSONFormatter{}
	result := f.FormatPages([]models.Page{})
	if !json.Valid([]byte(result)) {
		t.Fatalf("FormatPages([]) returned invalid JSON: %s", result)
	}
	if result != "[]" {
		t.Errorf("expected '[]', got %s", result)
	}
}

func TestJSONFormatter_FormatSpace(t *testing.T) {
	f := &output.JSONFormatter{}
	result := f.FormatSpace(sampleSpace())

	if !json.Valid([]byte(result)) {
		t.Fatalf("FormatSpace returned invalid JSON:\n%s", result)
	}

	var space models.Space
	if err := json.Unmarshal([]byte(result), &space); err != nil {
		t.Fatalf("failed to unmarshal space JSON: %v", err)
	}
	if space.Key != "ENG" {
		t.Errorf("expected Key 'ENG', got %s", space.Key)
	}
}

func TestJSONFormatter_FormatSearchResults(t *testing.T) {
	f := &output.JSONFormatter{}
	result := f.FormatSearchResults(sampleSearchResult())

	if !json.Valid([]byte(result)) {
		t.Fatalf("FormatSearchResults returned invalid JSON:\n%s", result)
	}

	var sr models.SearchResult
	if err := json.Unmarshal([]byte(result), &sr); err != nil {
		t.Fatalf("failed to unmarshal search result JSON: %v", err)
	}
	if sr.TotalSize != 2 {
		t.Errorf("expected TotalSize 2, got %d", sr.TotalSize)
	}
}

func TestJSONFormatter_FormatPageTree(t *testing.T) {
	f := &output.JSONFormatter{}
	result := f.FormatPageTree(samplePageTree(), "https://example.atlassian.net")

	if !json.Valid([]byte(result)) {
		t.Fatalf("FormatPageTree returned invalid JSON:\n%s", result)
	}
}

// --------------------------------------------------------------------------
// PrettyFormatter tests
// --------------------------------------------------------------------------

func TestPrettyFormatter_FormatPage(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatPage(samplePage())

	if !strings.Contains(result, "Getting Started") {
		t.Error("expected output to contain page title 'Getting Started'")
	}
	// Should contain ANSI codes (bold = ESC[1m)
	if !strings.Contains(result, "\033[") {
		t.Error("expected output to contain ANSI escape codes")
	}
}

func TestPrettyFormatter_FormatPage_Nil(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatPage(nil)
	// Should not panic; should return a graceful error string
	if result == "" {
		t.Error("expected non-empty result for nil page")
	}
}

func TestPrettyFormatter_FormatPage_NilVersionAndBody(t *testing.T) {
	f := &output.PrettyFormatter{}
	page := &models.Page{
		ID:      "999",
		Status:  "current",
		Title:   "Minimal Page",
		SpaceID: "SPACE001",
		// Version and Body are nil
	}
	// Must not panic
	result := f.FormatPage(page)
	if !strings.Contains(result, "Minimal Page") {
		t.Error("expected output to contain page title")
	}
}

func TestPrettyFormatter_FormatPages(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatPages(samplePages())

	if !strings.Contains(result, "Page One") {
		t.Error("expected output to contain 'Page One'")
	}
	if !strings.Contains(result, "Page Two (Draft)") {
		t.Error("expected output to contain 'Page Two (Draft)'")
	}
}

func TestPrettyFormatter_FormatPages_Empty(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatPages([]models.Page{})
	if result == "" {
		t.Error("expected non-empty result for empty pages slice")
	}
}

func TestPrettyFormatter_FormatSpace(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatSpace(sampleSpace())

	if !strings.Contains(result, "Engineering") {
		t.Error("expected output to contain space name 'Engineering'")
	}
	if !strings.Contains(result, "ENG") {
		t.Error("expected output to contain space key 'ENG'")
	}
}

func TestPrettyFormatter_FormatSearchResults(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatSearchResults(sampleSearchResult())

	if !strings.Contains(result, "Getting Started") {
		t.Error("expected output to contain 'Getting Started'")
	}
	if !strings.Contains(result, "2") {
		t.Error("expected output to contain result count")
	}
}

func TestPrettyFormatter_FormatLabels(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatLabels(sampleLabels())

	if !strings.Contains(result, "getting-started") {
		t.Error("expected output to contain label 'getting-started'")
	}
	if !strings.Contains(result, "api") {
		t.Error("expected output to contain label 'api'")
	}
}

func TestPrettyFormatter_FormatLabels_Empty(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatLabels([]models.Label{})
	if result == "" {
		t.Error("expected non-empty result for empty labels slice")
	}
}

func TestPrettyFormatter_FormatAttachments(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatAttachments(sampleAttachments())

	if !strings.Contains(result, "diagram.png") {
		t.Error("expected output to contain 'diagram.png'")
	}
	if !strings.Contains(result, "report.pdf") {
		t.Error("expected output to contain 'report.pdf'")
	}
	// Human-readable sizes
	if !strings.Contains(result, "KB") && !strings.Contains(result, "MB") {
		t.Error("expected output to contain human-readable file size units")
	}
}

func TestPrettyFormatter_FormatPageTree(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatPageTree(samplePageTree(), "https://example.atlassian.net")

	if !strings.Contains(result, "Root Page") {
		t.Error("expected tree output to contain 'Root Page'")
	}
	if !strings.Contains(result, "Child One") {
		t.Error("expected tree output to contain 'Child One'")
	}
	if !strings.Contains(result, "Grandchild") {
		t.Error("expected tree output to contain 'Grandchild'")
	}
	// Box-drawing characters
	if !strings.Contains(result, "├──") && !strings.Contains(result, "└──") {
		t.Error("expected tree output to contain box-drawing characters (├── or └──)")
	}
}

func TestPrettyFormatter_FormatPageTree_Nil(t *testing.T) {
	f := &output.PrettyFormatter{}
	result := f.FormatPageTree(nil, "https://example.atlassian.net")
	if result == "" {
		t.Error("expected non-empty result for nil tree")
	}
}

// --------------------------------------------------------------------------
// MarkdownFormatter tests
// --------------------------------------------------------------------------

func TestMarkdownFormatter_FormatPage(t *testing.T) {
	f := &output.MarkdownFormatter{}
	result := f.FormatPage(samplePage())

	if !strings.HasPrefix(result, "# Getting Started") {
		t.Errorf("expected Markdown heading '# Getting Started', got: %s", result[:min(50, len(result))])
	}
	if !strings.Contains(result, "| ID |") {
		t.Error("expected output to contain Markdown table with ID column")
	}
}

func TestMarkdownFormatter_FormatPages(t *testing.T) {
	f := &output.MarkdownFormatter{}
	result := f.FormatPages(samplePages())

	// Check header row
	if !strings.Contains(result, "| ID | Title | Status | Space ID |") {
		t.Error("expected Markdown table header with ID, Title, Status, Space ID columns")
	}
	if !strings.Contains(result, "Page One") {
		t.Error("expected table to contain 'Page One'")
	}
	if !strings.Contains(result, "Page Two (Draft)") {
		t.Error("expected table to contain 'Page Two (Draft)'")
	}
	// Separator row should contain ---
	if !strings.Contains(result, "| --- |") {
		t.Error("expected Markdown table separator row")
	}
}

func TestMarkdownFormatter_FormatPages_Empty(t *testing.T) {
	f := &output.MarkdownFormatter{}
	result := f.FormatPages([]models.Page{})
	if !strings.Contains(result, "No pages") {
		t.Error("expected 'No pages' message for empty slice")
	}
}

func TestMarkdownFormatter_FormatSpace(t *testing.T) {
	f := &output.MarkdownFormatter{}
	result := f.FormatSpace(sampleSpace())

	if !strings.HasPrefix(result, "# Engineering") {
		t.Errorf("expected Markdown heading '# Engineering'")
	}
	if !strings.Contains(result, "| Key |") {
		t.Error("expected output to contain Key field")
	}
}

func TestMarkdownFormatter_FormatSpaces(t *testing.T) {
	f := &output.MarkdownFormatter{}
	result := f.FormatSpaces(sampleSpaces())

	if !strings.Contains(result, "| Key | Name | Type | Status |") {
		t.Error("expected Markdown table header for spaces")
	}
	if !strings.Contains(result, "ENG") {
		t.Error("expected table to contain 'ENG'")
	}
}

func TestMarkdownFormatter_FormatSearchResults(t *testing.T) {
	f := &output.MarkdownFormatter{}
	result := f.FormatSearchResults(sampleSearchResult())

	if !strings.Contains(result, "**2 result(s) found**") {
		t.Error("expected bold result count header")
	}
	if !strings.Contains(result, "| Title | Space | Excerpt |") {
		t.Error("expected Markdown table header with Title, Space, Excerpt")
	}
	// Linked title
	if !strings.Contains(result, "[Getting Started]") {
		t.Error("expected linked title for search result")
	}
}

func TestMarkdownFormatter_FormatLabels(t *testing.T) {
	f := &output.MarkdownFormatter{}
	result := f.FormatLabels(sampleLabels())

	if !strings.Contains(result, "`getting-started`") {
		t.Error("expected backtick-quoted label 'getting-started'")
	}
	if !strings.Contains(result, "`api`") {
		t.Error("expected backtick-quoted label 'api'")
	}
}

func TestMarkdownFormatter_FormatAttachments(t *testing.T) {
	f := &output.MarkdownFormatter{}
	result := f.FormatAttachments(sampleAttachments())

	if !strings.Contains(result, "| ID | Title | Type | Size |") {
		t.Error("expected Markdown table header for attachments")
	}
	if !strings.Contains(result, "diagram.png") {
		t.Error("expected table to contain 'diagram.png'")
	}
}

func TestMarkdownFormatter_FormatPageTree(t *testing.T) {
	f := &output.MarkdownFormatter{}
	result := f.FormatPageTree(samplePageTree(), "https://example.atlassian.net")

	if !strings.Contains(result, "- [Root Page]") {
		t.Error("expected Markdown list item with linked root page")
	}
	if !strings.Contains(result, "  - [Child One]") {
		t.Error("expected indented child page")
	}
	if !strings.Contains(result, "    - [Grandchild]") {
		t.Error("expected doubly-indented grandchild page")
	}
}

// min is a helper for Go versions before 1.21 generic min.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
