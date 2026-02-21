package output

import (
	"encoding/json"

	"github.com/kristofferrisa/confluence-cli/internal/models"
)

// JSONFormatter outputs all data as indented JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) marshal(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return `{"error": "failed to marshal to JSON"}`
	}
	return string(b)
}

// FormatPage returns a page as indented JSON.
func (f *JSONFormatter) FormatPage(page *models.Page) string {
	return f.marshal(page)
}

// FormatPages returns a slice of pages as an indented JSON array.
func (f *JSONFormatter) FormatPages(pages []models.Page) string {
	return f.marshal(pages)
}

// FormatSpace returns a space as indented JSON.
func (f *JSONFormatter) FormatSpace(space *models.Space) string {
	return f.marshal(space)
}

// FormatSpaces returns a slice of spaces as an indented JSON array.
func (f *JSONFormatter) FormatSpaces(spaces []models.Space) string {
	return f.marshal(spaces)
}

// FormatSearchResults returns search results as indented JSON.
func (f *JSONFormatter) FormatSearchResults(results *models.SearchResult) string {
	return f.marshal(results)
}

// FormatLabels returns a slice of labels as an indented JSON array.
func (f *JSONFormatter) FormatLabels(labels []models.Label) string {
	return f.marshal(labels)
}

// FormatAttachments returns a slice of attachments as an indented JSON array.
func (f *JSONFormatter) FormatAttachments(attachments []models.Attachment) string {
	return f.marshal(attachments)
}

// FormatPageTree returns a page tree as indented JSON.
// The baseURL parameter is accepted to satisfy the Formatter interface but is not used
// since the full struct is serialised as-is.
func (f *JSONFormatter) FormatPageTree(tree *models.PageTree, _ string) string {
	return f.marshal(tree)
}
