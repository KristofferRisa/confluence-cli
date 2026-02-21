package converter

import (
	"bytes"
	"fmt"
	stdhtml "html"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	nethtml "golang.org/x/net/html"
)

// ---- Goldmark setup -------------------------------------------------------

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.Table,
		extension.Strikethrough,
		extension.TaskList,
	),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		gmhtml.WithXHTML(),
		gmhtml.WithUnsafe(), // allow raw HTML pass-through
	),
)

// ---- Regex patterns for post-processing -----------------------------------

// reCodeBlockWithLang matches <pre><code class="language-XXX">...</code></pre>
var reCodeBlockWithLang = regexp.MustCompile(
	`(?s)<pre><code class="language-([^"]+)">(.*?)</code></pre>`,
)

// reCodeBlockNoLang matches <pre><code>...</code></pre> (no language class)
var reCodeBlockNoLang = regexp.MustCompile(
	`(?s)<pre><code>(.*?)</code></pre>`,
)

// reImage matches <img ... />
var reImage = regexp.MustCompile(
	`<img([^>]*?)/?>`,
)

var reImgSrc = regexp.MustCompile(`src="([^"]*)"`)
var reImgAlt = regexp.MustCompile(`alt="([^"]*)"`)

// ---- MarkdownToStorage ----------------------------------------------------

// MarkdownToStorage converts markdown content to Confluence storage format XHTML.
func MarkdownToStorage(markdown string) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(markdown), &buf); err != nil {
		return "", fmt.Errorf("goldmark convert: %w", err)
	}

	result := buf.String()

	// Post-process: code blocks with language → ac:structured-macro.
	// Must run before the no-language variant to avoid double-matching.
	result = reCodeBlockWithLang.ReplaceAllStringFunc(result, func(match string) string {
		sub := reCodeBlockWithLang.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		lang := sub[1]
		code := stdhtml.UnescapeString(sub[2])
		return confluenceCodeMacro(lang, code)
	})

	// Post-process: code blocks without language → ac:structured-macro.
	result = reCodeBlockNoLang.ReplaceAllStringFunc(result, func(match string) string {
		sub := reCodeBlockNoLang.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		code := stdhtml.UnescapeString(sub[1])
		return confluenceCodeMacro("", code)
	})

	// Post-process: images → ac:image.
	result = reImage.ReplaceAllStringFunc(result, func(match string) string {
		srcMatch := reImgSrc.FindStringSubmatch(match)
		altMatch := reImgAlt.FindStringSubmatch(match)

		src := ""
		if len(srcMatch) >= 2 {
			src = srcMatch[1]
		}
		alt := ""
		if len(altMatch) >= 2 {
			alt = altMatch[1]
		}

		return confluenceImage(src, alt)
	})

	return result, nil
}

// confluenceCodeMacro builds an ac:structured-macro code block.
func confluenceCodeMacro(lang, code string) string {
	var sb strings.Builder
	sb.WriteString(`<ac:structured-macro ac:name="code">`)
	if lang != "" {
		sb.WriteString(`<ac:parameter ac:name="language">`)
		sb.WriteString(lang)
		sb.WriteString(`</ac:parameter>`)
	}
	sb.WriteString(`<ac:plain-text-body><![CDATA[`)
	sb.WriteString(code)
	sb.WriteString(`]]></ac:plain-text-body>`)
	sb.WriteString(`</ac:structured-macro>`)
	return sb.String()
}

// confluenceImage builds an ac:image element.
func confluenceImage(src, alt string) string {
	var sb strings.Builder
	sb.WriteString(`<ac:image`)
	if alt != "" {
		sb.WriteString(` ac:alt="`)
		sb.WriteString(stdhtml.EscapeString(alt))
		sb.WriteString(`"`)
	}
	sb.WriteString(`>`)
	sb.WriteString(`<ri:url ri:value="`)
	sb.WriteString(stdhtml.EscapeString(src))
	sb.WriteString(`" />`)
	sb.WriteString(`</ac:image>`)
	return sb.String()
}

// ---- StorageToMarkdown ----------------------------------------------------

// StorageToMarkdown converts Confluence storage format XHTML back to markdown.
func StorageToMarkdown(storage string) (string, error) {
	conv := &storageConverter{}
	if err := conv.convert(storage); err != nil {
		return "", err
	}
	return conv.result(), nil
}

// storageConverter is a state-machine HTML tokenizer that builds markdown output.
type storageConverter struct {
	buf strings.Builder

	// Block-level state
	inParagraph  bool
	inBlockquote bool
	listDepth    int
	listOrdered  []bool // stack: true = ordered

	// Inline state
	inStrong bool
	inEm     bool
	inCode   bool
	inLink   bool
	linkHref string
	linkText strings.Builder

	// Table state
	inTable     bool
	inThead     bool
	inTbody     bool
	inTR        bool
	inTH        bool
	inTD        bool
	tableRows   [][]string
	currentRow  []string
	currentCell strings.Builder

	// Heading state
	headingLevel int

	// Code macro state
	inCodeMacro   bool
	codeMacroLang string
	codeMacroBody strings.Builder
	codeParamName string
	inCodeParam   bool
	inCodeBody    bool

	// ac:image state
	inImage  bool
	imageAlt string
	imageSrc string

	// Deferred newline flushing to avoid trailing blanks.
	pendingNewlines int
}

func (c *storageConverter) convert(storage string) error {
	// Wrap content so the tokenizer sees a complete document fragment.
	wrapped := "<root>" + storage + "</root>"

	tokenizer := nethtml.NewTokenizer(strings.NewReader(wrapped))

	for {
		tt := tokenizer.Next()
		switch tt {
		case nethtml.ErrorToken:
			return nil // io.EOF
		case nethtml.StartTagToken, nethtml.SelfClosingTagToken:
			name, hasAttr := tokenizer.TagName()
			tagName := string(name)
			attrs := parseTokenAttrs(tokenizer, hasAttr)
			c.handleStartTag(tagName, attrs, tt == nethtml.SelfClosingTagToken)
		case nethtml.EndTagToken:
			name, _ := tokenizer.TagName()
			c.handleEndTag(string(name))
		case nethtml.TextToken:
			c.handleText(string(tokenizer.Text()))
		case nethtml.CommentToken:
			c.handleComment(string(tokenizer.Raw()))
		}
	}
}

func parseTokenAttrs(tokenizer *nethtml.Tokenizer, hasAttr bool) map[string]string {
	attrs := make(map[string]string)
	if !hasAttr {
		return attrs
	}
	for {
		key, val, more := tokenizer.TagAttr()
		attrs[string(key)] = string(val)
		if !more {
			break
		}
	}
	return attrs
}

// writeNewline queues a newline for deferred output.
func (c *storageConverter) writeNewline() {
	c.pendingNewlines++
}

// flushNewlines writes all queued newlines.
func (c *storageConverter) flushNewlines() {
	for i := 0; i < c.pendingNewlines; i++ {
		c.buf.WriteByte('\n')
	}
	c.pendingNewlines = 0
}

// write flushes any pending newlines, then writes s.
func (c *storageConverter) write(s string) {
	if s == "" {
		return
	}
	c.flushNewlines()
	c.buf.WriteString(s)
}

func (c *storageConverter) handleStartTag(tag string, attrs map[string]string, selfClosing bool) {
	switch tag {
	// ---- Headings ----
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(tag[1] - '0')
		c.headingLevel = level
		c.ensureBlankLine()
		c.write(strings.Repeat("#", level) + " ")

	// ---- Paragraphs ----
	case "p":
		if !c.inBlockquote {
			c.ensureBlankLine()
		}
		c.inParagraph = true

	// ---- Inline formatting ----
	case "strong", "b":
		c.inStrong = true
		c.write("**")

	case "em", "i":
		c.inEm = true
		c.write("*")

	case "code":
		if !c.inCodeMacro {
			c.inCode = true
			c.write("`")
		}

	// ---- Links ----
	case "a":
		c.inLink = true
		c.linkHref = attrs["href"]
		c.linkText.Reset()

	// ---- Lists ----
	case "ul":
		c.listOrdered = append(c.listOrdered, false)
		c.listDepth++
		if c.listDepth == 1 {
			c.ensureBlankLine()
		}

	case "ol":
		c.listOrdered = append(c.listOrdered, true)
		c.listDepth++
		if c.listDepth == 1 {
			c.ensureBlankLine()
		}

	case "li":
		c.flushNewlines()
		indent := strings.Repeat("  ", c.listDepth-1)
		ordered := len(c.listOrdered) > 0 && c.listOrdered[len(c.listOrdered)-1]
		if ordered {
			c.write(indent + "1. ")
		} else {
			c.write(indent + "- ")
		}

	// ---- Blockquote ----
	case "blockquote":
		c.inBlockquote = true
		c.ensureBlankLine()

	// ---- Horizontal rule ----
	case "hr":
		c.ensureBlankLine()
		c.write("---")
		c.writeNewline()
		c.writeNewline()

	// ---- Line break ----
	case "br":
		c.write("  \n")

	// ---- Tables ----
	case "table":
		c.inTable = true
		c.tableRows = nil
		c.ensureBlankLine()

	case "thead":
		c.inThead = true

	case "tbody":
		c.inTbody = true

	case "tr":
		c.inTR = true
		c.currentRow = nil

	case "th":
		c.inTH = true
		c.currentCell.Reset()

	case "td":
		c.inTD = true
		c.currentCell.Reset()

	// ---- Confluence: code macro ----
	case "ac:structured-macro":
		if attrs["ac:name"] == "code" {
			c.inCodeMacro = true
			c.codeMacroLang = ""
			c.codeMacroBody.Reset()
		}

	case "ac:parameter":
		if c.inCodeMacro && attrs["ac:name"] == "language" {
			c.inCodeParam = true
			c.codeParamName = "language"
		}

	case "ac:plain-text-body":
		if c.inCodeMacro {
			c.inCodeBody = true
		}

	// ---- Confluence: image ----
	case "ac:image":
		c.inImage = true
		c.imageAlt = attrs["ac:alt"]
		c.imageSrc = ""

	case "ri:url":
		if c.inImage {
			c.imageSrc = attrs["ri:value"]
		}

	// ---- Ignored structural elements ----
	case "root", "html", "head", "body", "pre":
		// no-op
	}
}

func (c *storageConverter) handleEndTag(tag string) {
	switch tag {
	// ---- Headings ----
	case "h1", "h2", "h3", "h4", "h5", "h6":
		c.headingLevel = 0
		c.writeNewline()
		c.writeNewline()

	// ---- Paragraphs ----
	case "p":
		c.inParagraph = false
		if !c.inBlockquote {
			c.writeNewline()
			c.writeNewline()
		} else {
			c.writeNewline()
		}

	// ---- Inline formatting ----
	case "strong", "b":
		c.inStrong = false
		c.write("**")

	case "em", "i":
		c.inEm = false
		c.write("*")

	case "code":
		if !c.inCodeMacro {
			c.inCode = false
			c.write("`")
		}

	// ---- Links ----
	case "a":
		c.inLink = false
		text := c.linkText.String()
		c.write("[" + text + "](" + c.linkHref + ")")
		c.linkText.Reset()
		c.linkHref = ""

	// ---- Lists ----
	case "ul", "ol":
		if len(c.listOrdered) > 0 {
			c.listOrdered = c.listOrdered[:len(c.listOrdered)-1]
		}
		c.listDepth--
		if c.listDepth == 0 {
			c.writeNewline()
			c.writeNewline()
		}

	case "li":
		c.writeNewline()

	// ---- Blockquote ----
	case "blockquote":
		c.inBlockquote = false
		c.writeNewline()
		c.writeNewline()

	// ---- Tables ----
	case "table":
		c.inTable = false
		c.renderTable()

	case "thead":
		c.inThead = false

	case "tbody":
		c.inTbody = false

	case "tr":
		c.inTR = false
		if len(c.currentRow) > 0 {
			c.tableRows = append(c.tableRows, c.currentRow)
		}
		c.currentRow = nil

	case "th":
		c.inTH = false
		c.currentRow = append(c.currentRow, c.currentCell.String())
		c.currentCell.Reset()

	case "td":
		c.inTD = false
		c.currentRow = append(c.currentRow, c.currentCell.String())
		c.currentCell.Reset()

	// ---- Confluence: code macro ----
	case "ac:structured-macro":
		if c.inCodeMacro {
			c.inCodeMacro = false
			c.inCodeBody = false
			c.ensureBlankLine()
			lang := c.codeMacroLang
			code := c.codeMacroBody.String()
			if lang != "" {
				c.write("```" + lang + "\n")
			} else {
				c.write("```\n")
			}
			c.write(code)
			if !strings.HasSuffix(code, "\n") {
				c.write("\n")
			}
			c.write("```")
			c.writeNewline()
			c.writeNewline()
		}

	case "ac:parameter":
		c.inCodeParam = false
		c.codeParamName = ""

	case "ac:plain-text-body":
		c.inCodeBody = false

	// ---- Confluence: image ----
	case "ac:image":
		if c.inImage {
			c.inImage = false
			c.write("![" + c.imageAlt + "](" + c.imageSrc + ")")
		}
	}
}

func (c *storageConverter) handleText(text string) {
	// Discard whitespace-only text between block-level elements.
	if strings.TrimSpace(text) == "" &&
		!c.inParagraph && !c.inTH && !c.inTD &&
		c.headingLevel == 0 && !c.inStrong && !c.inEm && !c.inCode &&
		!c.inLink && c.listDepth == 0 && !c.inBlockquote && !c.inCodeMacro {
		return
	}

	// Route text to the correct consumer.
	switch {
	case c.inTH || c.inTD:
		c.currentCell.WriteString(text)
	case c.inLink:
		c.linkText.WriteString(text)
	case c.inCodeParam && c.codeParamName == "language":
		c.codeMacroLang += text
	case c.inCodeBody:
		c.codeMacroBody.WriteString(text)
	case c.inBlockquote:
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if i == 0 {
				c.write("> " + line)
			} else {
				c.write("\n> " + line)
			}
		}
	default:
		c.write(text)
	}
}

// handleComment captures CDATA sections inside code macros.
//
// The golang.org/x/net/html tokenizer treats <![CDATA[...]]> as a bogus comment.
// Raw() returns the full raw bytes including the surrounding angle brackets, i.e.
// "<![CDATA[...]]>". We check for both forms just to be safe.
func (c *storageConverter) handleComment(raw string) {
	if !c.inCodeMacro {
		return
	}

	// Raw() includes the surrounding angle brackets: "<![CDATA[...]]>"
	const cdataFull = "<![CDATA["
	const cdataSuffix = "]]>"

	if strings.HasPrefix(raw, cdataFull) && strings.HasSuffix(raw, cdataSuffix) {
		content := raw[len(cdataFull) : len(raw)-len(cdataSuffix)]
		c.codeMacroBody.WriteString(content)
		return
	}

	// Fallback: some tokenizer versions may omit the leading <
	const cdataShort = "![CDATA["
	if strings.HasPrefix(raw, cdataShort) && strings.HasSuffix(raw, cdataSuffix) {
		content := raw[len(cdataShort) : len(raw)-len(cdataSuffix)]
		c.codeMacroBody.WriteString(content)
	}
}

// ensureBlankLine guarantees a blank line before the next output.
func (c *storageConverter) ensureBlankLine() {
	if c.buf.Len() == 0 && c.pendingNewlines == 0 {
		return
	}
	s := c.buf.String()
	alreadyBlank := strings.HasSuffix(s, "\n\n") ||
		(strings.HasSuffix(s, "\n") && c.pendingNewlines >= 1) ||
		c.pendingNewlines >= 2
	if !alreadyBlank {
		if strings.HasSuffix(s, "\n") {
			c.pendingNewlines = max(c.pendingNewlines, 1)
		} else {
			c.pendingNewlines = max(c.pendingNewlines, 2)
		}
	}
}

// renderTable emits the collected table rows as a markdown pipe table.
func (c *storageConverter) renderTable() {
	if len(c.tableRows) == 0 {
		return
	}

	// Count maximum columns.
	cols := 0
	for _, row := range c.tableRows {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if cols == 0 {
		return
	}

	// Calculate column widths (minimum 3 for the separator dashes).
	widths := make([]int, cols)
	for _, row := range c.tableRows {
		for j, cell := range row {
			trimmed := strings.TrimSpace(cell)
			if len(trimmed) > widths[j] {
				widths[j] = len(trimmed)
			}
		}
	}
	for i := range widths {
		if widths[i] < 3 {
			widths[i] = 3
		}
	}

	var sb strings.Builder

	for rowIdx, row := range c.tableRows {
		sb.WriteString("|")
		for colIdx := 0; colIdx < cols; colIdx++ {
			cell := ""
			if colIdx < len(row) {
				cell = strings.TrimSpace(row[colIdx])
			}
			padding := widths[colIdx] - len(cell)
			sb.WriteString(" " + cell + strings.Repeat(" ", padding) + " |")
		}
		sb.WriteString("\n")

		// Separator row after the header (first row).
		if rowIdx == 0 {
			sb.WriteString("|")
			for colIdx := 0; colIdx < cols; colIdx++ {
				sb.WriteString(" " + strings.Repeat("-", widths[colIdx]) + " |")
			}
			sb.WriteString("\n")
		}
	}

	c.write(sb.String())
	c.writeNewline()
}

// result returns the final markdown string.
func (c *storageConverter) result() string {
	c.flushNewlines()
	out := c.buf.String()
	out = strings.TrimRight(out, " \t\n")
	if out != "" {
		out += "\n"
	}
	return out
}
