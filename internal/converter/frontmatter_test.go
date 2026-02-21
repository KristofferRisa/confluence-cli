package converter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseFrontmatter_Normal(t *testing.T) {
	input := `---
title: My Page
space: MYSPACE
page_id: "12345"
parent_id: "99999"
labels:
  - go
  - cli
---
# Hello

Body content here.
`
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Title != "My Page" {
		t.Errorf("Title: want %q, got %q", "My Page", fm.Title)
	}
	if fm.Space != "MYSPACE" {
		t.Errorf("Space: want %q, got %q", "MYSPACE", fm.Space)
	}
	if fm.PageID != "12345" {
		t.Errorf("PageID: want %q, got %q", "12345", fm.PageID)
	}
	if fm.ParentID != "99999" {
		t.Errorf("ParentID: want %q, got %q", "99999", fm.ParentID)
	}
	if len(fm.Labels) != 2 || fm.Labels[0] != "go" || fm.Labels[1] != "cli" {
		t.Errorf("Labels: want [go cli], got %v", fm.Labels)
	}
	if !strings.Contains(body, "# Hello") {
		t.Errorf("body should contain '# Hello', got: %q", body)
	}
	if !strings.Contains(body, "Body content here.") {
		t.Errorf("body should contain 'Body content here.', got: %q", body)
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	input := `# Plain Markdown

No frontmatter here.
`
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm != nil {
		t.Errorf("expected nil frontmatter, got %+v", fm)
	}
	if body != input {
		t.Errorf("body should equal full input when no frontmatter present\nwant: %q\ngot:  %q", input, body)
	}
}

func TestParseFrontmatter_EmptyFrontmatter(t *testing.T) {
	input := `---
---
Body here.
`
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected non-nil frontmatter struct for empty frontmatter")
	}
	if fm.Title != "" || fm.Space != "" || fm.PageID != "" {
		t.Errorf("expected all fields empty, got %+v", fm)
	}
	if !strings.Contains(body, "Body here.") {
		t.Errorf("body should contain 'Body here.', got: %q", body)
	}
}

func TestParseFrontmatter_PartialFields(t *testing.T) {
	input := `---
title: Just Title
space: MYSPACE
---
Content.
`
	fm, body, err := ParseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Title != "Just Title" {
		t.Errorf("Title: want %q, got %q", "Just Title", fm.Title)
	}
	if fm.Space != "MYSPACE" {
		t.Errorf("Space: want %q, got %q", "MYSPACE", fm.Space)
	}
	if fm.PageID != "" {
		t.Errorf("PageID: want empty, got %q", fm.PageID)
	}
	if fm.ParentID != "" {
		t.Errorf("ParentID: want empty, got %q", fm.ParentID)
	}
	if len(fm.Labels) != 0 {
		t.Errorf("Labels: want empty, got %v", fm.Labels)
	}
	if !strings.Contains(body, "Content.") {
		t.Errorf("body should contain 'Content.', got: %q", body)
	}
}

func TestRenderFrontmatter(t *testing.T) {
	fm := &Frontmatter{
		Title:    "Test Page",
		Space:    "TESTSPACE",
		PageID:   "42",
		ParentID: "10",
		Labels:   []string{"alpha", "beta"},
	}

	out, err := RenderFrontmatter(fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(out, "---\n") {
		t.Errorf("output should start with '---\\n', got: %q", out[:10])
	}
	if !strings.HasSuffix(out, "---\n") {
		t.Errorf("output should end with '---\\n', got tail: %q", out[len(out)-10:])
	}
	if !strings.Contains(out, "title: Test Page") {
		t.Errorf("output should contain 'title: Test Page', got:\n%s", out)
	}
	if !strings.Contains(out, "space: TESTSPACE") {
		t.Errorf("output should contain 'space: TESTSPACE', got:\n%s", out)
	}
	if !strings.Contains(out, "page_id:") {
		t.Errorf("output should contain 'page_id:', got:\n%s", out)
	}
	if !strings.Contains(out, "alpha") {
		t.Errorf("output should contain 'alpha' label, got:\n%s", out)
	}
}

func TestRenderFrontmatter_Nil(t *testing.T) {
	out, err := RenderFrontmatter(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty string for nil frontmatter, got: %q", out)
	}
}

func TestRoundTrip(t *testing.T) {
	fm := &Frontmatter{
		Title:  "Round Trip Page",
		Space:  "RT",
		PageID: "777",
		Labels: []string{"roundtrip"},
	}
	body := "# Heading\n\nSome body content.\n"

	rendered, err := RenderFrontmatter(fm)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	fullContent := rendered + body

	parsed, parsedBody, err := ParseFrontmatter(fullContent)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed frontmatter, got nil")
	}
	if parsed.Title != fm.Title {
		t.Errorf("Title round-trip: want %q, got %q", fm.Title, parsed.Title)
	}
	if parsed.Space != fm.Space {
		t.Errorf("Space round-trip: want %q, got %q", fm.Space, parsed.Space)
	}
	if parsed.PageID != fm.PageID {
		t.Errorf("PageID round-trip: want %q, got %q", fm.PageID, parsed.PageID)
	}
	if len(parsed.Labels) != 1 || parsed.Labels[0] != "roundtrip" {
		t.Errorf("Labels round-trip: want [roundtrip], got %v", parsed.Labels)
	}
	if parsedBody != body {
		t.Errorf("body round-trip mismatch\nwant: %q\ngot:  %q", body, parsedBody)
	}
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")

	content := `---
title: File Test
space: FS
---
File body.
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	fm, body, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Title != "File Test" {
		t.Errorf("Title: want %q, got %q", "File Test", fm.Title)
	}
	if fm.Space != "FS" {
		t.Errorf("Space: want %q, got %q", "FS", fm.Space)
	}
	if !strings.Contains(body, "File body.") {
		t.Errorf("body should contain 'File body.', got: %q", body)
	}
}

func TestParseFile_NotFound(t *testing.T) {
	_, _, err := ParseFile("/nonexistent/path/file.md")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "output.md")

	fm := &Frontmatter{
		Title: "Written Page",
		Space: "WP",
	}
	body := "# Written\n\nContent.\n"

	if err := WriteFile(path, fm, body); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	// Read back and verify.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back error: %v", err)
	}

	parsedFM, parsedBody, err := ParseFrontmatter(string(data))
	if err != nil {
		t.Fatalf("parse written file: %v", err)
	}
	if parsedFM == nil {
		t.Fatal("expected frontmatter in written file")
	}
	if parsedFM.Title != "Written Page" {
		t.Errorf("Title: want %q, got %q", "Written Page", parsedFM.Title)
	}
	if parsedFM.Space != "WP" {
		t.Errorf("Space: want %q, got %q", "WP", parsedFM.Space)
	}
	if parsedBody != body {
		t.Errorf("body mismatch\nwant: %q\ngot:  %q", body, parsedBody)
	}

	// Verify file permissions.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat written file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0644 {
		t.Errorf("file permissions: want 0644, got %04o", perm)
	}
}
