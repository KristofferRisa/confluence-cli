# Confluence CLI Architecture Reference

## Project Structure

```
cmd/cfluence/main.go          # Entry point
internal/
  commands/                    # Cobra command definitions
    root.go                    # Root command, config loading, formatter setup
    page.go                    # push, pull, get, list, delete, tree
    search.go                  # CQL and text search
    label.go                   # add, list, remove labels
    attachment.go              # upload, list, download attachments
    space.go                   # list, get spaces
    config.go                  # init, show, set, path
    version.go                 # Version info
  api/                         # Confluence REST API client
    client.go                  # HTTP client with basic auth
    pages.go                   # Page CRUD (v2 API)
    spaces.go                  # Space operations (v2 API)
    labels.go                  # Label management (v1 API)
    attachments.go             # Attachment operations (v1 API)
    search.go                  # CQL search (v1 API)
  converter/                   # Format conversion
    markdown.go                # Markdown <-> Confluence storage format (XHTML)
    frontmatter.go             # YAML frontmatter parsing/serialization
  models/                      # Shared data types
    types.go                   # Page, Space, Label, Attachment, etc.
  output/                      # Output formatting
    formatter.go               # Formatter interface
    pretty.go                  # Colored CLI output
    json.go                    # JSON output
    markdown.go                # Markdown output
```

## Key Flow

```
User command → commands/ → api/ (HTTP) → Confluence Cloud
                         → converter/ (markdown ↔ storage format)
                         → output/ (format response)
```

## API Details

- **REST v2** (`/wiki/api/v2`): Pages and spaces (full CRUD)
- **REST v1** (`/wiki/rest/api`): Labels, attachments, search (v2 is read-only for these)
- **Auth**: Basic auth with `email:token` Base64-encoded in Authorization header
- **Space creation**: v2 requires `spaceId` (not `spaceKey`), so pages always resolve key→ID first

## Converter

The converter uses goldmark to parse markdown to HTML, then post-processes for Confluence:
- Code blocks become `ac:structured-macro` elements with `ac:parameter` for language
- Images become `ac:image` elements with `ri:url` references
- Standard HTML elements map to Confluence storage format equivalents

## Config Priority

1. CLI flags (`--config`, `--format`)
2. Environment variables (`CONFLUENCE_BASE_URL`, `CONFLUENCE_EMAIL`, `CONFLUENCE_TOKEN`)
3. Config file (`~/.confluence/config.yaml`)

## Development Commands

```bash
make build          # Build binary to ./cfluence
make test           # Run all tests
make fmt            # Format code
make lint           # Run golangci-lint
make build-all      # Cross-compile for linux/darwin/windows
go test -v ./internal/config -run TestLoad_EnvVarTakesPriority  # Single test
```
