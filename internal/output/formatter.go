package output

import "github.com/kristofferrisa/confluence-cli/internal/models"

// Formatter defines the interface for output formatting
type Formatter interface {
	FormatPage(page *models.Page) string
	FormatPages(pages []models.Page) string
	FormatSpace(space *models.Space) string
	FormatSpaces(spaces []models.Space) string
	FormatSearchResults(results *models.SearchResult) string
	FormatLabels(labels []models.Label) string
	FormatAttachments(attachments []models.Attachment) string
	FormatPageTree(tree *models.PageTree, baseURL string) string
}

// New creates a formatter based on the format name
func New(format string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	case "markdown", "md":
		return &MarkdownFormatter{}
	case "pretty", "":
		return &PrettyFormatter{}
	default:
		return &PrettyFormatter{}
	}
}
