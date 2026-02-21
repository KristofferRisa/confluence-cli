# Confluence CLI

A command-line tool for managing Confluence Cloud pages using markdown files.

<p align="center">
  <img src="https://img.shields.io/github/v/release/kristofferrisa/confluence-cli" alt="Release">
  <img src="https://img.shields.io/github/actions/workflow/status/kristofferrisa/confluence-cli/test.yml" alt="Tests">
  <img src="https://img.shields.io/github/license/kristofferrisa/confluence-cli" alt="License">
</p>

## Features

- **Markdown-first workflow** - Author pages as `.md` files with YAML frontmatter, push to Confluence
- **Bidirectional sync** - Push markdown to Confluence, pull Confluence pages back as markdown
- **Full CRUD** - Create, read, update, and delete pages, spaces, labels, and attachments
- **CQL search** - Search Confluence content from the command line
- **Multiple output formats** - Colored CLI (default), JSON for scripting, Markdown for docs
- **Cross-platform** - Works on Linux, macOS, and Windows

## Installation

### Download Binary

Download the latest release for your platform from [Releases](https://github.com/kristofferrisa/confluence-cli/releases).

### Build from Source

```bash
git clone https://github.com/kristofferrisa/confluence-cli.git
cd confluence-cli
make build
./confluence --help
```

## Quick Start

1. **Get your API token** from [id.atlassian.com/manage/api-tokens](https://id.atlassian.com/manage/api-tokens)

2. **Run setup wizard:**
   ```bash
   confluence config init
   ```

3. **Pull a page as markdown:**
   ```bash
   confluence page pull 123456 -o my-page.md
   ```

4. **Edit the markdown file, then push:**
   ```bash
   confluence page push my-page.md
   ```

## Markdown File Format

Pages are represented as standard markdown files with YAML frontmatter:

```markdown
---
title: "Architecture Decision Record"
space: "ENG"
parent_id: "123456"
page_id: "789012"
labels:
  - architecture
  - adr
---

# Architecture Decision Record

## Status

Accepted

## Context

We need a way to manage Confluence pages as code...
```

| Field | Required | Description |
|-------|----------|-------------|
| `title` | Yes | Page title in Confluence |
| `space` | Yes (or set default) | Space key (e.g., `ENG`, `DOCS`) |
| `parent_id` | No | Parent page ID for nested pages |
| `page_id` | No | Set automatically after first push (used for updates) |
| `labels` | No | List of labels to apply to the page |

## Usage

### Configuration

**Option 1: Environment variables (recommended for CI)**
```bash
export CONFLUENCE_BASE_URL="https://mycompany.atlassian.net"
export CONFLUENCE_EMAIL="user@company.com"
export CONFLUENCE_TOKEN="your-api-token"
confluence page list --space ENG
```

**Option 2: Config file**
```bash
confluence config init  # Interactive setup
# or manually edit ~/.confluence/config.yaml
```

**Option 3: Command flags**
```bash
confluence --config /path/to/config.yaml page list
```

### Commands

#### Pages

```bash
# Push a markdown file to Confluence (creates or updates)
confluence page push my-page.md

# Push multiple files
confluence page push docs/*.md

# Pull a page as markdown
confluence page pull 123456 -o my-page.md

# Get page info
confluence page get 123456

# List pages in a space
confluence page list --space ENG

# List pages with filtering
confluence page list --space ENG --status current --limit 25

# Show page hierarchy as a tree
confluence page tree --space ENG

# Delete a page
confluence page delete 123456
```

#### Spaces

```bash
# List all accessible spaces
confluence space list

# Get space details
confluence space get ENG
```

#### Search

```bash
# Search with CQL
confluence search "type=page AND space=ENG AND title~\"architecture\""

# Simple text search
confluence search "deployment guide" --space ENG
```

#### Labels

```bash
# Add labels to a page
confluence label add 123456 architecture design-doc

# List labels on a page
confluence label list 123456

# Remove a label
confluence label remove 123456 draft
```

#### Attachments

```bash
# Upload an attachment to a page
confluence attachment upload 123456 diagram.png

# List attachments on a page
confluence attachment list 123456

# Download an attachment
confluence attachment download att789012 -o diagram.png
```

#### Configuration

```bash
# Interactive setup wizard
confluence config init

# Show current configuration
confluence config show

# Set a single value
confluence config set space ENG
confluence config set format json

# Show config file path
confluence config path
```

#### Version

```bash
confluence version
```

### Output Formats

Default output is colored CLI. Change format with `--format`:

**JSON** (for scripting/piping):
```bash
confluence page list --space ENG --format json | jq '.[].title'
```

**Markdown** (for documentation):
```bash
confluence page get 123456 --format markdown
```

## Configuration File

Location: `~/.confluence/config.yaml`

```yaml
base_url: "https://mycompany.atlassian.net"
email: "user@company.com"
token: "your-api-token"
space: "ENG"              # Default space key
format: "pretty"          # Options: pretty, json, markdown
```

View current config:
```bash
confluence config show
```

## Typical Workflows

### Documentation as Code

Keep your Confluence docs in a git repo and push on merge:

```bash
# In your CI pipeline
for f in docs/*.md; do
  confluence page push "$f"
done
```

### Pull, Edit, Push

```bash
confluence page pull 123456 -o design.md
# Edit design.md in your favorite editor
confluence page push design.md
```

### Bulk Export

```bash
# Export all pages in a space
confluence page list --space ENG --format json | \
  jq -r '.[].id' | \
  xargs -I{} confluence page pull {} -o "pages/{}.md"
```

## API Information

- **REST API v2:** `{base_url}/wiki/api/v2` (pages, spaces)
- **REST API v1:** `{base_url}/wiki/rest/api` (labels, attachments, search)
- **Authentication:** Basic auth with email + API token
- **Rate limits:** Points-based system (65,000+ points/hour depending on plan)
- **Documentation:** [developer.atlassian.com/cloud/confluence](https://developer.atlassian.com/cloud/confluence/rest/v2/intro/)

## Troubleshooting

**"No API token found"**
- Set `CONFLUENCE_TOKEN` environment variable or run `confluence config init`

**"401 Unauthorized"**
- Verify your email and API token are correct
- Ensure the token was generated at [id.atlassian.com/manage/api-tokens](https://id.atlassian.com/manage/api-tokens)

**"404 Page not found"**
- Verify the page ID exists and you have access to it
- Check that `base_url` points to the correct Confluence instance

**"429 Too Many Requests"**
- You've hit the rate limit; the CLI will show the retry-after time
- Distribute requests or wait for the quota to reset (hourly)

## Development

### Build
```bash
make build          # Build ./confluence
make build-all      # Cross-compile all platforms
```

### Test
```bash
make test           # Run all tests
go test ./internal/converter -run TestMarkdownToStorage  # Run specific test
```

### Lint & Format
```bash
make fmt            # Format code
make lint           # Run linter (requires golangci-lint)
```

## Contributing

Contributions welcome! Please read [ARCHITECTURE.md](ARCHITECTURE.md) for code structure details.

1. Fork the repository
2. Create a feature branch (`git checkout -b feat-amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feat-amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Made by [Kristoffer Risa](https://github.com/kristofferrisa)
