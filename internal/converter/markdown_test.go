package converter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- MarkdownToStorage tests -----------------------------------------------

func TestMarkdownToStorage_Headings(t *testing.T) {
	input := `# Heading 1

## Heading 2

### Heading 3

#### Heading 4

##### Heading 5

###### Heading 6
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []string{"<h1", "<h2", "<h3", "<h4", "<h5", "<h6"}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("expected output to contain %q\nGot:\n%s", want, out)
		}
	}
}

func TestMarkdownToStorage_Paragraphs(t *testing.T) {
	input := `First paragraph.

Second paragraph.
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "<p>First paragraph.</p>") {
		t.Errorf("expected <p>First paragraph.</p>\nGot:\n%s", out)
	}
	if !strings.Contains(out, "<p>Second paragraph.</p>") {
		t.Errorf("expected <p>Second paragraph.</p>\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_Bold_Italic(t *testing.T) {
	input := `**bold text** and *italic text* and ***bold italic***.
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "<strong>bold text</strong>") {
		t.Errorf("expected <strong>bold text</strong>\nGot:\n%s", out)
	}
	if !strings.Contains(out, "<em>italic text</em>") {
		t.Errorf("expected <em>italic text</em>\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_Links(t *testing.T) {
	input := `Visit [Confluence](https://confluence.example.com) for more.
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `href="https://confluence.example.com"`) {
		t.Errorf("expected href attribute\nGot:\n%s", out)
	}
	if !strings.Contains(out, "Confluence") {
		t.Errorf("expected link text\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_CodeBlock_WithLanguage(t *testing.T) {
	input := "```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\n"

	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `ac:structured-macro ac:name="code"`) {
		t.Errorf("expected ac:structured-macro\nGot:\n%s", out)
	}
	if !strings.Contains(out, `ac:parameter ac:name="language"`) {
		t.Errorf("expected language parameter\nGot:\n%s", out)
	}
	if !strings.Contains(out, ">go<") {
		t.Errorf("expected 'go' language value\nGot:\n%s", out)
	}
	if !strings.Contains(out, "<![CDATA[") {
		t.Errorf("expected CDATA section\nGot:\n%s", out)
	}
	if !strings.Contains(out, "func main()") {
		t.Errorf("expected code content in CDATA\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_CodeBlock_NoLanguage(t *testing.T) {
	input := "```\nplain code here\n```\n"

	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, `ac:structured-macro ac:name="code"`) {
		t.Errorf("expected ac:structured-macro\nGot:\n%s", out)
	}
	// Should NOT have a language parameter.
	if strings.Contains(out, `ac:parameter ac:name="language"`) {
		t.Errorf("unexpected language parameter for no-language block\nGot:\n%s", out)
	}
	if !strings.Contains(out, "plain code here") {
		t.Errorf("expected code content\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_CodeBlock_HTMLEntitiesUnescaped(t *testing.T) {
	// goldmark escapes < > & in code blocks; we must unescape inside CDATA.
	input := "```go\nif x < 10 && y > 0 {\n\tfmt.Println(\"ok\")\n}\n```\n"

	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The CDATA should contain the raw characters, not HTML entities.
	if strings.Contains(out, "&lt;") {
		t.Errorf("CDATA should contain literal '<', not '&lt;'\nGot:\n%s", out)
	}
	if strings.Contains(out, "&amp;") {
		t.Errorf("CDATA should contain literal '&', not '&amp;'\nGot:\n%s", out)
	}
	if !strings.Contains(out, "x < 10 && y > 0") {
		t.Errorf("expected unescaped code in CDATA\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_InlineCode(t *testing.T) {
	input := "Use `fmt.Println` to print.\n"

	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "<code>fmt.Println</code>") {
		t.Errorf("expected <code>fmt.Println</code>\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_Table(t *testing.T) {
	input := `| Name  | Age |
|-------|-----|
| Alice | 30  |
| Bob   | 25  |
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "<table>") {
		t.Errorf("expected <table>\nGot:\n%s", out)
	}
	if !strings.Contains(out, "<th>") {
		t.Errorf("expected <th>\nGot:\n%s", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("expected 'Alice' in table\nGot:\n%s", out)
	}
	if !strings.Contains(out, "Bob") {
		t.Errorf("expected 'Bob' in table\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_Lists_Unordered(t *testing.T) {
	input := `- item one
- item two
- item three
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "<ul>") {
		t.Errorf("expected <ul>\nGot:\n%s", out)
	}
	if !strings.Contains(out, "<li>") {
		t.Errorf("expected <li>\nGot:\n%s", out)
	}
	if !strings.Contains(out, "item one") {
		t.Errorf("expected list content\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_Lists_Ordered(t *testing.T) {
	input := `1. first
2. second
3. third
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "<ol>") {
		t.Errorf("expected <ol>\nGot:\n%s", out)
	}
	if !strings.Contains(out, "first") {
		t.Errorf("expected list content\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_Blockquote(t *testing.T) {
	input := `> This is a blockquote.
> It spans multiple lines.
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "<blockquote>") {
		t.Errorf("expected <blockquote>\nGot:\n%s", out)
	}
	if !strings.Contains(out, "This is a blockquote.") {
		t.Errorf("expected blockquote content\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_Image(t *testing.T) {
	input := `![My Alt Text](https://example.com/image.png)
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "<ac:image") {
		t.Errorf("expected <ac:image\nGot:\n%s", out)
	}
	if !strings.Contains(out, "https://example.com/image.png") {
		t.Errorf("expected image URL\nGot:\n%s", out)
	}
}

func TestMarkdownToStorage_HorizontalRule(t *testing.T) {
	input := `Before rule.

---

After rule.
`
	out, err := MarkdownToStorage(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "<hr") {
		t.Errorf("expected <hr\nGot:\n%s", out)
	}
}

// ---- StorageToMarkdown tests -----------------------------------------------

func TestStorageToMarkdown_Headings(t *testing.T) {
	input := `<h1>Title One</h1>
<h2>Title Two</h2>
<h3>Title Three</h3>
<h4>Title Four</h4>
<h5>Title Five</h5>
<h6>Title Six</h6>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	checks := []struct {
		prefix string
		text   string
	}{
		{"# ", "Title One"},
		{"## ", "Title Two"},
		{"### ", "Title Three"},
		{"#### ", "Title Four"},
		{"##### ", "Title Five"},
		{"###### ", "Title Six"},
	}
	for _, c := range checks {
		line := c.prefix + c.text
		if !strings.Contains(out, line) {
			t.Errorf("expected %q in output\nGot:\n%s", line, out)
		}
	}
}

func TestStorageToMarkdown_Paragraphs(t *testing.T) {
	input := `<p>First paragraph.</p>
<p>Second paragraph.</p>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "First paragraph.") {
		t.Errorf("expected first paragraph\nGot:\n%s", out)
	}
	if !strings.Contains(out, "Second paragraph.") {
		t.Errorf("expected second paragraph\nGot:\n%s", out)
	}
	// Should have blank line between paragraphs.
	if !strings.Contains(out, "\n\n") {
		t.Errorf("expected blank line between paragraphs\nGot:\n%s", out)
	}
}

func TestStorageToMarkdown_Links(t *testing.T) {
	input := `<p>Visit <a href="https://example.com">Example Site</a> for info.</p>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "[Example Site](https://example.com)") {
		t.Errorf("expected markdown link\nGot:\n%s", out)
	}
}

func TestStorageToMarkdown_CodeMacro_WithLanguage(t *testing.T) {
	input := `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">python</ac:parameter><ac:plain-text-body><![CDATA[def hello():
    print("world")
]]></ac:plain-text-body></ac:structured-macro>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "```python") {
		t.Errorf("expected ```python fence\nGot:\n%s", out)
	}
	if !strings.Contains(out, `def hello():`) {
		t.Errorf("expected code content\nGot:\n%s", out)
	}
	if !strings.Contains(out, "```") {
		t.Errorf("expected closing fence\nGot:\n%s", out)
	}
}

func TestStorageToMarkdown_CodeMacro_NoLanguage(t *testing.T) {
	input := `<ac:structured-macro ac:name="code"><ac:plain-text-body><![CDATA[plain text code]]></ac:plain-text-body></ac:structured-macro>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "```\n") {
		t.Errorf("expected plain opening fence\nGot:\n%s", out)
	}
	if !strings.Contains(out, "plain text code") {
		t.Errorf("expected code content\nGot:\n%s", out)
	}
}

func TestStorageToMarkdown_Table(t *testing.T) {
	input := `<table>
<thead><tr><th>Name</th><th>Value</th></tr></thead>
<tbody>
<tr><td>foo</td><td>bar</td></tr>
<tr><td>baz</td><td>qux</td></tr>
</tbody>
</table>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "| Name") {
		t.Errorf("expected table header\nGot:\n%s", out)
	}
	if !strings.Contains(out, "| foo") {
		t.Errorf("expected table row with 'foo'\nGot:\n%s", out)
	}
	if !strings.Contains(out, "---") {
		t.Errorf("expected separator row\nGot:\n%s", out)
	}
}

func TestStorageToMarkdown_Image(t *testing.T) {
	input := `<ac:image ac:alt="Logo"><ri:url ri:value="https://example.com/logo.png" /></ac:image>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "![Logo](https://example.com/logo.png)") {
		t.Errorf("expected markdown image\nGot:\n%s", out)
	}
}

func TestStorageToMarkdown_BoldItalic(t *testing.T) {
	input := `<p><strong>bold</strong> and <em>italic</em></p>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "**bold**") {
		t.Errorf("expected **bold**\nGot:\n%s", out)
	}
	if !strings.Contains(out, "*italic*") {
		t.Errorf("expected *italic*\nGot:\n%s", out)
	}
}

func TestStorageToMarkdown_Lists(t *testing.T) {
	input := `<ul><li>alpha</li><li>beta</li></ul>
<ol><li>one</li><li>two</li></ol>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "- alpha") {
		t.Errorf("expected '- alpha'\nGot:\n%s", out)
	}
	if !strings.Contains(out, "1. one") {
		t.Errorf("expected '1. one'\nGot:\n%s", out)
	}
}

func TestStorageToMarkdown_HorizontalRule(t *testing.T) {
	input := `<p>Before.</p><hr /><p>After.</p>`

	out, err := StorageToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "---") {
		t.Errorf("expected '---' for hr\nGot:\n%s", out)
	}
}

// ---- Round-trip test -------------------------------------------------------

func TestRoundTrip_SimpleDocument(t *testing.T) {
	original := `# My Document

This is a paragraph with **bold** and *italic* text.

## Section Two

- List item one
- List item two

Visit [Example](https://example.com) for more info.
`

	storage, err := MarkdownToStorage(original)
	if err != nil {
		t.Fatalf("MarkdownToStorage error: %v", err)
	}

	back, err := StorageToMarkdown(storage)
	if err != nil {
		t.Fatalf("StorageToMarkdown error: %v", err)
	}

	// The round-trip won't be byte-identical but should preserve key content.
	checks := []string{
		"# My Document",
		"## Section Two",
		"**bold**",
		"*italic*",
		"- List item one",
		"[Example](https://example.com)",
	}
	for _, want := range checks {
		if !strings.Contains(back, want) {
			t.Errorf("round-trip: expected %q in output\nGot:\n%s", want, back)
		}
	}
}

// ---- Golden file tests -----------------------------------------------------
//
// Run with UPDATE_GOLDEN=1 to regenerate the golden .storage.xml files:
//
//	UPDATE_GOLDEN=1 go test ./internal/converter/...

func runGoldenTest(t *testing.T, mdPath, storagePath string) {
	t.Helper()

	mdBytes, err := os.ReadFile(mdPath)
	if err != nil {
		t.Skipf("markdown source not found, skipping golden test: %v", err)
	}

	// Strip YAML frontmatter so we only convert the body.
	_, body, err := ParseFrontmatter(string(mdBytes))
	if err != nil {
		t.Fatalf("parse frontmatter: %v", err)
	}

	got, err := MarkdownToStorage(body)
	if err != nil {
		t.Fatalf("MarkdownToStorage: %v", err)
	}

	// If UPDATE_GOLDEN is set, write the result as the new golden file.
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.WriteFile(storagePath, []byte(got), 0644); err != nil {
			t.Fatalf("write golden file %s: %v", storagePath, err)
		}
		t.Logf("updated golden file: %s", storagePath)
		return
	}

	expectedBytes, err := os.ReadFile(storagePath)
	if err != nil {
		t.Skipf("golden storage file not found (run with UPDATE_GOLDEN=1 to generate): %v", err)
	}

	expected := strings.TrimSpace(string(expectedBytes))
	gotTrimmed := strings.TrimSpace(got)

	if gotTrimmed != expected {
		t.Errorf("golden mismatch for %s\nWant:\n%s\n\nGot:\n%s", storagePath, expected, gotTrimmed)
	}
}

func TestGolden_SimplePage(t *testing.T) {
	runGoldenTest(
		t,
		filepath.Join("..", "..", "testdata", "simple_page.md"),
		filepath.Join("..", "..", "testdata", "simple_page.storage.xml"),
	)
}

func TestGolden_ComplexPage(t *testing.T) {
	runGoldenTest(
		t,
		filepath.Join("..", "..", "testdata", "complex_page.md"),
		filepath.Join("..", "..", "testdata", "complex_page.storage.xml"),
	)
}
