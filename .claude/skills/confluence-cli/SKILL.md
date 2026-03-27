---
name: cfluence-cli
description: >-
  Assists with using the cfluence CLI tool for managing Confluence Cloud pages
  via markdown files. Use when the user asks about Confluence, cfluence commands,
  markdown frontmatter for Confluence, CQL search queries, pushing/pulling pages,
  or docs-as-code workflows with Confluence.
user-invocable: true
argument-hint: "[command or question]"
allowed-tools: Read Grep Glob Bash(cfluence *)
metadata:
  author: kristofferrisa
  version: "1.0"
  repository: https://github.com/kristofferrisa/cfluence-cli
compatibility: Requires cfluence binary installed (Go CLI). Works on Linux, macOS, Windows.
---

# Confluence CLI (cfluence) Skill

You are an expert on the `cfluence` CLI tool for managing Confluence Cloud pages using markdown files.

## Core Concepts

**cfluence** is a markdown-first CLI for Confluence Cloud. The key workflow is:

1. Author pages as `.md` files with YAML frontmatter
2. Push to Confluence with `cfluence page push`
3. Pull pages back as markdown with `cfluence page pull`
4. Search, manage labels, and handle attachments from the terminal

## Markdown File Format

Every markdown file uses YAML frontmatter between `---` delimiters:

```markdown
---
title: "Page Title"
space: "ENG"
parent_id: "123456"
page_id: "789012"
labels:
  - architecture
  - adr
---

# Your content here

Standard markdown content...
```

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `title` | Yes | Page title in Confluence |
| `space` | Yes (or set default) | Space key (e.g., `ENG`, `DOCS`) |
| `parent_id` | No | Parent page ID for nesting under another page |
| `page_id` | No | Auto-set after first push. Presence means "update", absence means "create" |
| `labels` | No | List of labels to apply to the page |

**Important behaviors:**
- On first `push`, `page_id` is written back to the file automatically
- If `page_id` is present, push updates the existing page
- If `page_id` is absent, push creates a new page
- `space` can be omitted if a default is configured via `cfluence config set space ENG`

## Command Reference

### Configuration

```bash
cfluence config init          # Interactive setup wizard
cfluence config show          # Show current configuration
cfluence config set space ENG # Set a default value
cfluence config path          # Show config file path
```

**Config priority:** CLI flags > environment variables > config file (`~/.confluence/config.yaml`)

**Environment variables:**
- `CONFLUENCE_BASE_URL` - Your Confluence instance URL (e.g., `https://mycompany.atlassian.net`)
- `CONFLUENCE_EMAIL` - Your Atlassian account email
- `CONFLUENCE_TOKEN` - Your API token from https://id.atlassian.com/manage/api-tokens

### Pages

```bash
cfluence page push <file.md> [file2.md ...]   # Push markdown to Confluence
cfluence page pull <page-id> [-o output.md]    # Pull page as markdown
cfluence page get <page-id>                    # Get page info
cfluence page list --space ENG [--limit 25]    # List pages in a space
cfluence page tree --space ENG [--depth 3]     # Show page hierarchy
cfluence page delete <page-id> [--force]       # Delete a page
```

### Search

```bash
# Simple text search
cfluence search "deployment guide"

# Text search scoped to a space
cfluence search "deployment guide" --space ENG

# Raw CQL query (auto-detected when query contains CQL operators)
cfluence search 'type=page AND space=ENG AND title~"API"'
```

CQL operators that trigger raw mode: `=`, `~`, `AND`, `OR`, `IN`, `NOT`, `ORDER BY`

### Labels

```bash
cfluence label add <page-id> <label1> [label2 ...]  # Add labels
cfluence label list <page-id>                         # List labels
cfluence label remove <page-id> <label>               # Remove a label
```

### Attachments

```bash
cfluence attachment upload <page-id> <file>                     # Upload file
cfluence attachment list <page-id>                               # List attachments
cfluence attachment download <page-id> <attachment-id> [-o out]  # Download
```

### Output Formats

All commands support `--format` flag: `pretty` (default, colored), `json`, `markdown`

```bash
cfluence page list --space ENG --format json | jq '.[].title'
```

## Common Workflows

### New page from scratch

```bash
# 1. Create a markdown file
cat > my-page.md << 'EOF'
---
title: "My New Page"
space: "ENG"
labels:
  - draft
---

# My New Page

Content goes here...
EOF

# 2. Push to Confluence (page_id gets written back automatically)
cfluence page push my-page.md
```

### Pull, edit, push cycle

```bash
cfluence page pull 123456 -o design.md
# Edit design.md in your editor
cfluence page push design.md
```

### Docs-as-code in CI

```bash
export CONFLUENCE_BASE_URL="https://mycompany.atlassian.net"
export CONFLUENCE_EMAIL="$CI_EMAIL"
export CONFLUENCE_TOKEN="$CI_TOKEN"

for f in docs/*.md; do
  cfluence page push "$f"
done
```

### Bulk export

```bash
cfluence page list --space ENG --format json | \
  jq -r '.[].id' | \
  xargs -I{} cfluence page pull {} -o "pages/{}.md"
```

## CQL Query Examples

When helping users write CQL search queries:

```
# Pages modified in the last week
type=page AND lastModified > now("-1w")

# Pages by a specific author in a space
type=page AND space="ENG" AND creator="user@company.com"

# Pages with a specific label
type=page AND label="architecture"

# Pages with title containing a term
type=page AND title~"design"

# Combined filters
type=page AND space="ENG" AND label="adr" AND title~"authentication"
```

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| "No API token found" | Missing credentials | Set `CONFLUENCE_TOKEN` env var or run `cfluence config init` |
| "401 Unauthorized" | Wrong credentials | Verify email + token at id.atlassian.com/manage/api-tokens |
| "404 Page not found" | Wrong page ID or no access | Check page ID exists and your account has access |
| "429 Too Many Requests" | Rate limited | Wait for the retry-after period (quota resets hourly) |
| "space is required" | No space specified | Add `space` to frontmatter or run `cfluence config set space ENG` |

## Guidelines for Assisting Users

When a user asks for help with Confluence CLI tasks:

1. **Creating markdown files**: Always include proper YAML frontmatter with at least `title` and `space`
2. **Updating pages**: Ensure `page_id` is present in frontmatter (pull first if needed)
3. **CQL queries**: Default to simple text search; only use raw CQL when the user needs advanced filtering
4. **CI/CD**: Recommend environment variables for credentials, never hardcode tokens
5. **Batch operations**: Use `--format json` with `jq` for scripting
6. **Credentials**: Never log, print, or expose tokens in output

For detailed architecture information, see [references/architecture.md](references/architecture.md).
