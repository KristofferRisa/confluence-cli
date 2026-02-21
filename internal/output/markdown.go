package output

import (
	"fmt"
	"strings"

	"github.com/kristofferrisa/confluence-cli/internal/models"
)

// MarkdownFormatter outputs data as GitHub-flavoured Markdown tables and headings.
// It is intended for AI-readable output and pipe-friendly workflows.
type MarkdownFormatter struct{}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

// mdEscape escapes pipe characters inside cell values so they don't break tables.
func mdEscape(s string) string {
	return strings.ReplaceAll(s, "|", `\|`)
}

// mdTruncate truncates long strings so table cells stay readable.
func mdTruncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// mdRow formats a slice of cell values as a Markdown table row.
func mdRow(cells []string) string {
	return "| " + strings.Join(cells, " | ") + " |"
}

// mdSeparator returns a Markdown table separator row for n columns.
func mdSeparator(n int) string {
	cols := make([]string, n)
	for i := range cols {
		cols[i] = "---"
	}
	return mdRow(cols)
}

// --------------------------------------------------------------------------
// FormatPage
// --------------------------------------------------------------------------

// FormatPage renders a page as a Markdown heading followed by a key-value table.
func (f *MarkdownFormatter) FormatPage(page *models.Page) string {
	if page == nil {
		return "_No page._\n"
	}

	var sb strings.Builder

	sb.WriteString("# " + page.Title + "\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("| --- | --- |\n")
	sb.WriteString(fmt.Sprintf("| ID | %s |\n", mdEscape(page.ID)))
	sb.WriteString(fmt.Sprintf("| Status | %s |\n", mdEscape(page.Status)))
	sb.WriteString(fmt.Sprintf("| Space ID | %s |\n", mdEscape(page.SpaceID)))

	if page.ParentID != "" {
		sb.WriteString(fmt.Sprintf("| Parent ID | %s |\n", mdEscape(page.ParentID)))
	}
	if page.ParentType != "" {
		sb.WriteString(fmt.Sprintf("| Parent Type | %s |\n", mdEscape(page.ParentType)))
	}
	if page.AuthorID != "" {
		sb.WriteString(fmt.Sprintf("| Author ID | %s |\n", mdEscape(page.AuthorID)))
	}
	if page.CreatedAt != "" {
		sb.WriteString(fmt.Sprintf("| Created | %s |\n", mdEscape(page.CreatedAt)))
	}
	if page.Version != nil {
		sb.WriteString(fmt.Sprintf("| Version | %d |\n", page.Version.Number))
		if page.Version.CreatedAt != "" {
			sb.WriteString(fmt.Sprintf("| Version Date | %s |\n", mdEscape(page.Version.CreatedAt)))
		}
	}
	if page.Links != nil && page.Links.WebUI != "" {
		sb.WriteString(fmt.Sprintf("| Web URL | %s |\n", mdEscape(page.Links.WebUI)))
	}

	sb.WriteString("\n")
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatPages
// --------------------------------------------------------------------------

// FormatPages renders a Markdown table of pages.
func (f *MarkdownFormatter) FormatPages(pages []models.Page) string {
	if len(pages) == 0 {
		return "_No pages found._\n"
	}

	var sb strings.Builder

	sb.WriteString(mdRow([]string{"ID", "Title", "Status", "Space ID"}) + "\n")
	sb.WriteString(mdSeparator(4) + "\n")

	for _, p := range pages {
		sb.WriteString(mdRow([]string{
			mdEscape(p.ID),
			mdEscape(p.Title),
			mdEscape(p.Status),
			mdEscape(p.SpaceID),
		}) + "\n")
	}

	sb.WriteString(fmt.Sprintf("\n_%d page(s) total._\n", len(pages)))
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatSpace
// --------------------------------------------------------------------------

// FormatSpace renders a space as a Markdown heading followed by a key-value table.
func (f *MarkdownFormatter) FormatSpace(space *models.Space) string {
	if space == nil {
		return "_No space._\n"
	}

	var sb strings.Builder

	sb.WriteString("# " + space.Name + "\n\n")
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("| --- | --- |\n")
	sb.WriteString(fmt.Sprintf("| ID | %s |\n", mdEscape(space.ID)))
	sb.WriteString(fmt.Sprintf("| Key | %s |\n", mdEscape(space.Key)))
	sb.WriteString(fmt.Sprintf("| Type | %s |\n", mdEscape(space.Type)))
	sb.WriteString(fmt.Sprintf("| Status | %s |\n", mdEscape(space.Status)))

	if space.HomepageID != "" {
		sb.WriteString(fmt.Sprintf("| Homepage ID | %s |\n", mdEscape(space.HomepageID)))
	}
	if space.AuthorID != "" {
		sb.WriteString(fmt.Sprintf("| Author ID | %s |\n", mdEscape(space.AuthorID)))
	}
	if space.CreatedAt != "" {
		sb.WriteString(fmt.Sprintf("| Created | %s |\n", mdEscape(space.CreatedAt)))
	}
	if space.Description != nil && space.Description.Plain != nil && space.Description.Plain.Value != "" {
		sb.WriteString(fmt.Sprintf("| Description | %s |\n", mdEscape(mdTruncate(space.Description.Plain.Value, 120))))
	}
	if space.Links != nil && space.Links.WebUI != "" {
		sb.WriteString(fmt.Sprintf("| Web URL | %s |\n", mdEscape(space.Links.WebUI)))
	}

	sb.WriteString("\n")
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatSpaces
// --------------------------------------------------------------------------

// FormatSpaces renders a Markdown table of spaces.
func (f *MarkdownFormatter) FormatSpaces(spaces []models.Space) string {
	if len(spaces) == 0 {
		return "_No spaces found._\n"
	}

	var sb strings.Builder

	sb.WriteString(mdRow([]string{"Key", "Name", "Type", "Status"}) + "\n")
	sb.WriteString(mdSeparator(4) + "\n")

	for _, s := range spaces {
		sb.WriteString(mdRow([]string{
			mdEscape(s.Key),
			mdEscape(s.Name),
			mdEscape(s.Type),
			mdEscape(s.Status),
		}) + "\n")
	}

	sb.WriteString(fmt.Sprintf("\n_%d space(s) total._\n", len(spaces)))
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatSearchResults
// --------------------------------------------------------------------------

// FormatSearchResults renders search results as a Markdown table with linked titles.
func (f *MarkdownFormatter) FormatSearchResults(results *models.SearchResult) string {
	if results == nil {
		return "_No results._\n"
	}
	if len(results.Results) == 0 {
		return fmt.Sprintf("_No results found (total: %d)._\n", results.TotalSize)
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("**%d result(s) found**\n\n", results.TotalSize))

	sb.WriteString(mdRow([]string{"Title", "Space", "Excerpt"}) + "\n")
	sb.WriteString(mdSeparator(3) + "\n")

	for _, entry := range results.Results {
		title := entry.Title
		if title == "" {
			title = entry.Content.Title
		}

		// Linked title if URL is available.
		url := entry.URL
		if url == "" && entry.Content.Links != nil {
			url = entry.Content.Links.WebUI
		}
		var titleCell string
		if url != "" {
			titleCell = fmt.Sprintf("[%s](%s)", mdEscape(title), url)
		} else {
			titleCell = mdEscape(title)
		}

		spaceCell := ""
		if entry.Content.Space != nil {
			spaceCell = mdEscape(entry.Content.Space.Key)
			if entry.Content.Space.Name != "" {
				spaceCell += " — " + mdEscape(entry.Content.Space.Name)
			}
		}

		excerpt := mdTruncate(entry.Excerpt, 80)

		sb.WriteString(mdRow([]string{titleCell, spaceCell, mdEscape(excerpt)}) + "\n")
	}

	sb.WriteString("\n")
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatLabels
// --------------------------------------------------------------------------

// FormatLabels renders labels as a simple comma-separated list.
func (f *MarkdownFormatter) FormatLabels(labels []models.Label) string {
	if len(labels) == 0 {
		return "_No labels._\n"
	}

	names := make([]string, 0, len(labels))
	for _, l := range labels {
		names = append(names, "`"+l.Name+"`")
	}
	return strings.Join(names, ", ") + "\n"
}

// --------------------------------------------------------------------------
// FormatAttachments
// --------------------------------------------------------------------------

// FormatAttachments renders attachments as a Markdown table.
func (f *MarkdownFormatter) FormatAttachments(attachments []models.Attachment) string {
	if len(attachments) == 0 {
		return "_No attachments._\n"
	}

	var sb strings.Builder

	sb.WriteString(mdRow([]string{"ID", "Title", "Type", "Size"}) + "\n")
	sb.WriteString(mdSeparator(4) + "\n")

	for _, a := range attachments {
		mediaType := a.Extensions.MediaType
		if mediaType == "" {
			mediaType = a.Metadata.MediaType
		}
		size := humanFileSize(a.Extensions.FileSize)

		sb.WriteString(mdRow([]string{
			mdEscape(a.ID),
			mdEscape(a.Title),
			mdEscape(mediaType),
			size,
		}) + "\n")
	}

	sb.WriteString(fmt.Sprintf("\n_%d attachment(s) total._\n", len(attachments)))
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatPageTree
// --------------------------------------------------------------------------

// FormatPageTree renders a page tree as an indented Markdown list with links.
func (f *MarkdownFormatter) FormatPageTree(tree *models.PageTree, baseURL string) string {
	if tree == nil {
		return "_Empty tree._\n"
	}

	var sb strings.Builder
	renderMarkdownTree(&sb, tree, baseURL, 0)
	return sb.String()
}

// renderMarkdownTree is the recursive helper for FormatPageTree.
func renderMarkdownTree(sb *strings.Builder, node *models.PageTree, baseURL string, depth int) {
	if node == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	// Build URL.
	url := ""
	if node.Page.Links != nil && node.Page.Links.WebUI != "" {
		if strings.HasPrefix(node.Page.Links.WebUI, "http") {
			url = node.Page.Links.WebUI
		} else {
			url = baseURL + node.Page.Links.WebUI
		}
	}

	var itemStr string
	if url != "" {
		itemStr = fmt.Sprintf("%s- [%s](%s)", indent, mdEscape(node.Page.Title), url)
	} else {
		itemStr = fmt.Sprintf("%s- %s", indent, mdEscape(node.Page.Title))
	}
	sb.WriteString(itemStr + "\n")

	for i := range node.Children {
		renderMarkdownTree(sb, &node.Children[i], baseURL, depth+1)
	}
}
