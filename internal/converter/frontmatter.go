package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter holds metadata from markdown file YAML frontmatter.
type Frontmatter struct {
	Title    string   `yaml:"title"`
	Space    string   `yaml:"space"`
	PageID   string   `yaml:"page_id,omitempty"`
	ParentID string   `yaml:"parent_id,omitempty"`
	Labels   []string `yaml:"labels,omitempty"`
}

// ParseFile reads a markdown file and returns frontmatter + body content.
func ParseFile(path string) (*Frontmatter, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("read file %s: %w", path, err)
	}
	fm, body, err := ParseFrontmatter(string(data))
	if err != nil {
		return nil, "", fmt.Errorf("parse frontmatter in %s: %w", path, err)
	}
	return fm, body, nil
}

// WriteFile writes frontmatter + body content back to a markdown file.
// It creates parent directories if they do not exist.
func WriteFile(path string, fm *Frontmatter, body string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create parent directories for %s: %w", path, err)
	}

	rendered, err := RenderFrontmatter(fm)
	if err != nil {
		return fmt.Errorf("render frontmatter: %w", err)
	}

	content := rendered + body
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}
	return nil
}

// ParseFrontmatter parses YAML frontmatter from a raw markdown string.
// It returns the parsed Frontmatter struct, the body content (without the
// frontmatter block), and any error encountered.
//
// If no frontmatter is present (content does not start with "---"), it returns
// nil for the Frontmatter and the full content as the body.
func ParseFrontmatter(content string) (*Frontmatter, string, error) {
	// Frontmatter must start at the very beginning of the file.
	if !strings.HasPrefix(content, "---") {
		return nil, content, nil
	}

	// Split into lines to find the closing --- delimiter.
	// The spec: opening "---" on its own line, then YAML, then closing "---" on its own line.
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		// Just "---" with nothing after — treat as empty frontmatter.
		return &Frontmatter{}, "", nil
	}

	// lines[0] should be "---" (possibly with trailing \r)
	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		trimmed := strings.TrimRight(lines[i], "\r")
		if trimmed == "---" {
			closingIdx = i
			break
		}
	}

	if closingIdx == -1 {
		// No closing delimiter found — treat entire content as body (no frontmatter).
		return nil, content, nil
	}

	// Extract YAML block between the two delimiters.
	yamlLines := lines[1:closingIdx]
	yamlBlock := strings.Join(yamlLines, "\n")

	// Body is everything after the closing delimiter line.
	bodyLines := lines[closingIdx+1:]
	body := strings.Join(bodyLines, "\n")

	// Trim a single leading newline from body (common formatting).
	body = strings.TrimPrefix(body, "\n")

	// Handle empty frontmatter block.
	if strings.TrimSpace(yamlBlock) == "" {
		return &Frontmatter{}, body, nil
	}

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(yamlBlock), &fm); err != nil {
		return nil, "", fmt.Errorf("unmarshal YAML frontmatter: %w", err)
	}

	return &fm, body, nil
}

// RenderFrontmatter serializes a Frontmatter struct to YAML enclosed in ---
// delimiters, ready to be prepended to a markdown body.
func RenderFrontmatter(fm *Frontmatter) (string, error) {
	if fm == nil {
		return "", nil
	}

	data, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("marshal frontmatter to YAML: %w", err)
	}

	return "---\n" + string(data) + "---\n", nil
}
