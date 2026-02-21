package output

import (
	"fmt"
	"strings"

	"github.com/kristofferrisa/confluence-cli/internal/models"
)

// ANSI colour / style constants.
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// PrettyFormatter renders human-friendly, coloured CLI output.
type PrettyFormatter struct{}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

func bold(s string) string    { return colorBold + s + colorReset }
func cyan(s string) string    { return colorCyan + s + colorReset }
func blue(s string) string    { return colorBlue + s + colorReset }
func yellow(s string) string  { return colorYellow + s + colorReset }
func green(s string) string   { return colorGreen + s + colorReset }
func red(s string) string     { return colorRed + s + colorReset }
func dim(s string) string     { return colorDim + s + colorReset }

// statusBadge returns a coloured status string.
func statusBadge(status string) string {
	switch strings.ToLower(status) {
	case "current":
		return green("[" + status + "]")
	case "trashed", "deleted":
		return red("[" + status + "]")
	case "draft":
		return yellow("[" + status + "]")
	default:
		return dim("[" + status + "]")
	}
}

// humanFileSize converts bytes to a human-readable string (B, KB, MB, GB).
func humanFileSize(bytes int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// truncate shortens s to at most n runes, appending "…" when truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

// field renders a labelled key-value row.
func field(label, value string) string {
	return fmt.Sprintf("  %s %s\n", bold(cyan(label+":")), value)
}

// --------------------------------------------------------------------------
// FormatPage
// --------------------------------------------------------------------------

// FormatPage renders a detailed view of a single page.
func (f *PrettyFormatter) FormatPage(page *models.Page) string {
	if page == nil {
		return red("(no page)")
	}
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(bold("  \U0001F4C4 " + page.Title) + "\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	sb.WriteString(field("ID", page.ID))
	sb.WriteString(field("Status", statusBadge(page.Status)))
	sb.WriteString(field("Space ID", page.SpaceID))

	if page.ParentID != "" {
		sb.WriteString(field("Parent ID", page.ParentID))
	}

	if page.Version != nil {
		sb.WriteString(field("Version", fmt.Sprintf("%d", page.Version.Number)))
		if page.Version.CreatedAt != "" {
			sb.WriteString(field("Version Date", page.Version.CreatedAt))
		}
	}

	if page.CreatedAt != "" {
		sb.WriteString(field("Created", page.CreatedAt))
	}

	if page.AuthorID != "" {
		sb.WriteString(field("Author ID", page.AuthorID))
	}

	if page.Links != nil && page.Links.WebUI != "" {
		sb.WriteString(field("Web URL", blue(page.Links.WebUI)))
	}

	sb.WriteString("\n")
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatPages
// --------------------------------------------------------------------------

// FormatPages renders a table-like listing of pages.
func (f *PrettyFormatter) FormatPages(pages []models.Page) string {
	if len(pages) == 0 {
		return dim("  (no pages)\n")
	}

	var sb strings.Builder

	sb.WriteString("\n")
	// Header
	sb.WriteString(bold(fmt.Sprintf("  %-20s  %-40s  %-10s\n", "ID", "Title", "Status")))
	sb.WriteString("  " + strings.Repeat("─", 74) + "\n")

	for _, p := range pages {
		id := truncate(p.ID, 20)
		title := truncate(p.Title, 40)
		status := p.Status
		sb.WriteString(fmt.Sprintf("  %-20s  %-40s  %s\n",
			cyan(id),
			title,
			statusBadge(status),
		))
	}

	sb.WriteString(fmt.Sprintf("\n  %s\n\n", dim(fmt.Sprintf("%d page(s)", len(pages)))))
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatSpace
// --------------------------------------------------------------------------

// FormatSpace renders a detailed view of a single space.
func (f *PrettyFormatter) FormatSpace(space *models.Space) string {
	if space == nil {
		return red("(no space)")
	}
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(bold("  \U0001F4DA " + space.Name) + "\n")
	sb.WriteString(strings.Repeat("─", 60) + "\n")

	sb.WriteString(field("ID", space.ID))
	sb.WriteString(field("Key", yellow(space.Key)))
	sb.WriteString(field("Type", space.Type))
	sb.WriteString(field("Status", statusBadge(space.Status)))

	if space.HomepageID != "" {
		sb.WriteString(field("Homepage ID", space.HomepageID))
	}

	if space.CreatedAt != "" {
		sb.WriteString(field("Created", space.CreatedAt))
	}

	if space.AuthorID != "" {
		sb.WriteString(field("Author ID", space.AuthorID))
	}

	if space.Description != nil && space.Description.Plain != nil && space.Description.Plain.Value != "" {
		sb.WriteString(field("Description", truncate(space.Description.Plain.Value, 120)))
	}

	if space.Links != nil && space.Links.WebUI != "" {
		sb.WriteString(field("Web URL", blue(space.Links.WebUI)))
	}

	sb.WriteString("\n")
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatSpaces
// --------------------------------------------------------------------------

// FormatSpaces renders a table-like listing of spaces.
func (f *PrettyFormatter) FormatSpaces(spaces []models.Space) string {
	if len(spaces) == 0 {
		return dim("  (no spaces)\n")
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(bold(fmt.Sprintf("  %-12s  %-36s  %-12s\n", "Key", "Name", "Type")))
	sb.WriteString("  " + strings.Repeat("─", 64) + "\n")

	for _, s := range spaces {
		key := truncate(s.Key, 12)
		name := truncate(s.Name, 36)
		typ := truncate(s.Type, 12)
		sb.WriteString(fmt.Sprintf("  %-12s  %-36s  %-12s\n",
			yellow(key),
			name,
			dim(typ),
		))
	}

	sb.WriteString(fmt.Sprintf("\n  %s\n\n", dim(fmt.Sprintf("%d space(s)", len(spaces)))))
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatSearchResults
// --------------------------------------------------------------------------

// FormatSearchResults renders a list of search hits with excerpts.
func (f *PrettyFormatter) FormatSearchResults(results *models.SearchResult) string {
	if results == nil {
		return red("(no results)")
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(bold(fmt.Sprintf("  \U0001F50D %d result(s) found", results.TotalSize)))
	if results.TotalSize != results.Size && results.Size > 0 {
		sb.WriteString(dim(fmt.Sprintf(" (showing %d)", results.Size)))
	}
	sb.WriteString("\n")
	sb.WriteString("  " + strings.Repeat("─", 70) + "\n\n")

	for i, entry := range results.Results {
		// Title line
		title := entry.Title
		if title == "" {
			title = entry.Content.Title
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", bold(cyan(fmt.Sprintf("[%d]", i+1))), bold(title)))

		// Excerpt
		if entry.Excerpt != "" {
			sb.WriteString(fmt.Sprintf("      %s\n", dim(truncate(entry.Excerpt, 100))))
		}

		// Space
		if entry.Content.Space != nil && entry.Content.Space.Key != "" {
			sb.WriteString(fmt.Sprintf("      Space: %s\n", yellow(entry.Content.Space.Key+" — "+entry.Content.Space.Name)))
		}

		// URL
		url := entry.URL
		if url == "" && entry.Content.Links != nil {
			url = entry.Content.Links.WebUI
		}
		if url != "" {
			sb.WriteString(fmt.Sprintf("      URL:   %s\n", blue(url)))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// --------------------------------------------------------------------------
// FormatLabels
// --------------------------------------------------------------------------

// FormatLabels renders labels as a comma-separated tag list.
func (f *PrettyFormatter) FormatLabels(labels []models.Label) string {
	if len(labels) == 0 {
		return dim("  (no labels)\n")
	}

	var sb strings.Builder
	sb.WriteString("\n  ")

	parts := make([]string, 0, len(labels))
	for _, l := range labels {
		tag := cyan("[") + colorBold + colorGreen + l.Name + colorReset + cyan("]")
		parts = append(parts, tag)
	}
	sb.WriteString(strings.Join(parts, "  "))
	sb.WriteString("\n\n")
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatAttachments
// --------------------------------------------------------------------------

// FormatAttachments renders attachments with file size in human-readable form.
func (f *PrettyFormatter) FormatAttachments(attachments []models.Attachment) string {
	if len(attachments) == 0 {
		return dim("  (no attachments)\n")
	}

	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString(bold(fmt.Sprintf("  %-10s  %-36s  %-24s  %-10s\n", "ID", "Title", "Media Type", "Size")))
	sb.WriteString("  " + strings.Repeat("─", 84) + "\n")

	for _, a := range attachments {
		id := truncate(a.ID, 10)
		title := truncate(a.Title, 36)
		mediaType := a.Extensions.MediaType
		if mediaType == "" {
			mediaType = a.Metadata.MediaType
		}
		mediaType = truncate(mediaType, 24)
		size := humanFileSize(a.Extensions.FileSize)

		sb.WriteString(fmt.Sprintf("  %-10s  %-36s  %-24s  %-10s\n",
			dim(id),
			title,
			dim(mediaType),
			yellow(size),
		))
	}

	sb.WriteString(fmt.Sprintf("\n  %s\n\n", dim(fmt.Sprintf("%d attachment(s)", len(attachments)))))
	return sb.String()
}

// --------------------------------------------------------------------------
// FormatPageTree
// --------------------------------------------------------------------------

// FormatPageTree renders a recursive indented tree using box-drawing characters.
func (f *PrettyFormatter) FormatPageTree(tree *models.PageTree, baseURL string) string {
	if tree == nil {
		return dim("  (empty tree)\n")
	}

	var sb strings.Builder
	sb.WriteString("\n")
	renderPrettyTree(&sb, tree, baseURL, "", true)
	sb.WriteString("\n")
	return sb.String()
}

// renderPrettyTree is the recursive helper for FormatPageTree.
func renderPrettyTree(sb *strings.Builder, node *models.PageTree, baseURL, prefix string, isLast bool) {
	if node == nil {
		return
	}

	// Choose the connector glyph for this node.
	var connector string
	if prefix == "" {
		// Root node — no connector.
		connector = ""
	} else if isLast {
		connector = "└── "
	} else {
		connector = "├── "
	}

	// Build web URL from Links or baseURL fallback.
	url := ""
	if node.Page.Links != nil && node.Page.Links.WebUI != "" {
		if strings.HasPrefix(node.Page.Links.WebUI, "http") {
			url = node.Page.Links.WebUI
		} else {
			url = baseURL + node.Page.Links.WebUI
		}
	}

	titleStr := bold(node.Page.Title)
	idStr := dim("(" + node.Page.ID + ")")
	if url != "" {
		sb.WriteString(fmt.Sprintf("%s%s%s %s  %s\n", prefix, connector, titleStr, idStr, blue(url)))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s%s %s\n", prefix, connector, titleStr, idStr))
	}

	// Determine the prefix extension for children.
	var childPrefix string
	if prefix == "" {
		childPrefix = "  "
	} else if isLast {
		childPrefix = prefix + "    "
	} else {
		childPrefix = prefix + "│   "
	}

	for i, child := range node.Children {
		last := i == len(node.Children)-1
		renderPrettyTree(sb, &child, baseURL, childPrefix, last)
	}
}
