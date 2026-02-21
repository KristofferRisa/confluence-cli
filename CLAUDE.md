# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build          # Build binary to ./confluence
make test           # Run all tests
make fmt            # Format code
make lint           # Run golangci-lint
make tidy           # Tidy go.mod
make build-all      # Cross-compile for linux/darwin/windows
make clean          # Remove build artifacts
```

Run a single test:
```bash
go test -v ./internal/config -run TestLoad_EnvVarTakesPriority
```

## Architecture

This is a Go CLI for Confluence Cloud using Cobra for commands and markdown files as the primary input format.

**Key flow:** Commands (`internal/commands/`) → API client (`internal/api/`) → Converter (`internal/converter/`) → Output formatter (`internal/output/`)

**Configuration priority:** CLI flags > env vars (`CONFLUENCE_BASE_URL`, `CONFLUENCE_EMAIL`, `CONFLUENCE_TOKEN`) > `~/.confluence/config.yaml`

**Output formats:** `pretty` (default, colored), `json`, `markdown` — selected via `--format` flag

**API:** REST v2 at `{base_url}/wiki/api/v2` for pages/spaces, v1 at `{base_url}/wiki/rest/api` for labels/attachments/search

**Markdown files use YAML frontmatter** for metadata (title, space, page_id, parent_id, labels). The `page_id` field is written back after first push to enable updates.

## Key Patterns

- Commands use shared `cfg` and `formatter` from `root.go`
- All formatters implement the `Formatter` interface in `output/formatter.go`
- `internal/converter/` handles Markdown ↔ Confluence storage format (XHTML) conversion
- `internal/converter/frontmatter.go` parses/serializes YAML frontmatter in markdown files
- API client uses Basic Auth: `email:token` Base64-encoded in Authorization header
- Hybrid API: v2 for pages/spaces CRUD, v1 for labels/attachments/search (v2 is read-only for those)
- Exit code 1 for errors, clear error messages to stderr
- Golden file tests in `testdata/` for converter verification
- Token and credentials are never logged or printed in output
