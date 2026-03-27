# cfluence Architecture

A cross-platform CLI tool for managing Confluence Cloud pages using markdown files, built following Unix philosophy.

## Design Principles

1. **Do one thing well** - Each command has a single responsibility
2. **Markdown-first** - Pages are authored as `.md` files with YAML frontmatter
3. **Composable output** - JSON for piping, Markdown for humans/AI
4. **Fail fast, fail loud** - Clear error messages, non-zero exit codes
5. **Zero configuration to start** - Works with just environment variables

## Directory Structure

```
cfluence-cli/
├── cmd/
│   └── cfluence/
│       └── main.go              # Entry point, command registration
├── internal/
│   ├── api/
│   │   ├── client.go            # REST API HTTP client (shared)
│   │   ├── client_test.go       # Client tests with HTTP mocks
│   │   ├── pages.go             # Page CRUD operations (v2 API)
│   │   ├── pages_test.go        # Page endpoint tests
│   │   ├── spaces.go            # Space operations (v2 API)
│   │   ├── spaces_test.go       # Space endpoint tests
│   │   ├── search.go            # CQL search operations (v1 API)
│   │   ├── search_test.go       # Search endpoint tests
│   │   ├── labels.go            # Label operations (v1 API)
│   │   ├── labels_test.go       # Label endpoint tests
│   │   ├── attachments.go       # Attachment operations (v1 API)
│   │   └── attachments_test.go  # Attachment endpoint tests
│   ├── commands/
│   │   ├── root.go              # Root command, global flags
│   │   ├── version.go           # `cfluence version`
│   │   ├── config.go            # `cfluence config` - init/show/set/path
│   │   ├── page.go              # `cfluence page` - push/pull/get/list/delete/tree
│   │   ├── space.go             # `cfluence space` - list/get
│   │   ├── search.go            # `cfluence search`
│   │   ├── label.go             # `cfluence label` - add/list/remove
│   │   └── attachment.go        # `cfluence attachment` - upload/list/download
│   ├── config/
│   │   ├── config.go            # Configuration loading
│   │   └── config_test.go       # Config priority tests
│   ├── converter/
│   │   ├── markdown.go          # Markdown → Confluence storage format (XHTML)
│   │   ├── markdown_test.go     # Converter golden file tests
│   │   ├── frontmatter.go       # YAML frontmatter parse/serialize
│   │   └── frontmatter_test.go  # Frontmatter tests
│   ├── models/
│   │   └── types.go             # Data structures
│   └── output/
│       ├── formatter.go         # Formatter interface
│       ├── pretty.go            # Beautiful CLI output (default)
│       ├── json.go              # JSON formatter
│       ├── markdown.go          # Markdown formatter
│       └── formatter_test.go    # Formatter tests
├── testdata/
│   ├── simple_page.md           # Simple markdown input
│   ├── simple_page.storage.xml  # Expected storage format output
│   ├── complex_page.md          # Complex markdown (tables, code, images)
│   └── complex_page.storage.xml # Expected storage format output
├── go.mod
├── go.sum
├── Makefile
├── .goreleaser.yml
├── .gitignore
├── CLAUDE.md
├── ARCHITECTURE.md
├── README.md
└── LICENSE
```

## Component Overview

### Entry Point (`cmd/cfluence/main.go`)

Minimal entry point following Go conventions:

```go
func main() {
    if err := commands.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Configuration (`internal/config/`)

Configuration resolution order (first wins):

1. Command-line flags
2. Environment variables (`CONFLUENCE_BASE_URL`, `CONFLUENCE_EMAIL`, `CONFLUENCE_TOKEN`, `CONFLUENCE_SPACE`)
3. Config file (`~/.confluence/config.yaml`)
4. Defaults

```yaml
# ~/.confluence/config.yaml
base_url: "https://mycompany.atlassian.net"
email: "user@company.com"
token: "your-api-token"
space: "ENG"                # Default space key
format: "pretty"            # Default output format
```

### API Layer (`internal/api/`)

#### REST Client (`client.go`)

- Single HTTP client instance (connection pooling)
- Timeout: 30 seconds
- Retry: None (fail fast)
- Auth: Basic authentication (`email:token` Base64-encoded)
- Supports both v1 and v2 API base paths

```go
type Client struct {
    baseURL    string
    email      string
    token      string
    httpClient *http.Client
}
```

#### Hybrid API Strategy

The Confluence REST API v2 does not cover all operations. The client uses both versions:

| Operation | API Version | Base Path |
|-----------|-------------|-----------|
| Page CRUD | v2 | `/wiki/api/v2/pages` |
| Space operations | v2 | `/wiki/api/v2/spaces` |
| Search (CQL) | v1 | `/wiki/rest/api/search` |
| Label management | v1 | `/wiki/rest/api/content/{id}/label` |
| Attachment upload | v1 | `/wiki/rest/api/content/{id}/child/attachment` |
| Attachment download | v2 | `/wiki/api/v2/attachments/{id}` |

#### Page Operations (`pages.go`)

```go
func (c *Client) CreatePage(ctx context.Context, page *models.CreatePageRequest) (*models.Page, error)
func (c *Client) GetPage(ctx context.Context, pageID string) (*models.Page, error)
func (c *Client) UpdatePage(ctx context.Context, pageID string, page *models.UpdatePageRequest) (*models.Page, error)
func (c *Client) DeletePage(ctx context.Context, pageID string) error
func (c *Client) ListPages(ctx context.Context, spaceID string, opts *models.ListOptions) (*models.PageList, error)
func (c *Client) GetPageByTitle(ctx context.Context, spaceKey, title string) (*models.Page, error)
```

#### Space Operations (`spaces.go`)

```go
func (c *Client) GetSpace(ctx context.Context, spaceID string) (*models.Space, error)
func (c *Client) ListSpaces(ctx context.Context, opts *models.ListOptions) (*models.SpaceList, error)
func (c *Client) GetSpaceByKey(ctx context.Context, spaceKey string) (*models.Space, error)
```

#### Search (`search.go`)

```go
func (c *Client) Search(ctx context.Context, cql string, opts *models.ListOptions) (*models.SearchResult, error)
```

#### Labels (`labels.go`)

```go
func (c *Client) AddLabels(ctx context.Context, pageID string, labels []string) error
func (c *Client) GetLabels(ctx context.Context, pageID string) ([]models.Label, error)
func (c *Client) RemoveLabel(ctx context.Context, pageID string, label string) error
```

#### Attachments (`attachments.go`)

```go
func (c *Client) UploadAttachment(ctx context.Context, pageID string, filename string, reader io.Reader) (*models.Attachment, error)
func (c *Client) ListAttachments(ctx context.Context, pageID string) ([]models.Attachment, error)
func (c *Client) DownloadAttachment(ctx context.Context, attachmentID string, writer io.Writer) error
```

### Converter (`internal/converter/`)

The converter is the core differentiator — it transforms between markdown files and Confluence storage format (XHTML).

#### Frontmatter (`frontmatter.go`)

Parses and serializes YAML frontmatter from markdown files:

```go
type Frontmatter struct {
    Title    string   `yaml:"title"`
    Space    string   `yaml:"space"`
    PageID   string   `yaml:"page_id,omitempty"`
    ParentID string   `yaml:"parent_id,omitempty"`
    Labels   []string `yaml:"labels,omitempty"`
}

func ParseFile(path string) (*Frontmatter, string, error)       // Returns frontmatter + body
func WriteFile(path string, fm *Frontmatter, body string) error  // Writes frontmatter + body
```

#### Markdown Converter (`markdown.go`)

Converts between markdown and Confluence storage format:

```go
func MarkdownToStorage(markdown string) (string, error)  // MD → XHTML storage format
func StorageToMarkdown(storage string) (string, error)    // XHTML storage format → MD
```

**Conversion mappings:**

| Markdown | Confluence Storage Format |
|----------|--------------------------|
| `# Heading` | `<h1>Heading</h1>` |
| `**bold**` | `<strong>bold</strong>` |
| `*italic*` | `<em>italic</em>` |
| `[text](url)` | `<a href="url">text</a>` |
| `` `code` `` | `<code>code</code>` |
| ```` ```lang ```` | `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">lang</ac:parameter><ac:plain-text-body>...</ac:plain-text-body></ac:structured-macro>` |
| `![alt](url)` | `<ac:image><ri:url ri:value="url" /><ac:parameter ac:name="alt">alt</ac:parameter></ac:image>` |
| `\| table \|` | `<table><tbody><tr><td>...</td></tr></tbody></table>` |
| `> blockquote` | `<blockquote><p>...</p></blockquote>` |
| `- list item` | `<ul><li>list item</li></ul>` |
| `---` | `<hr />` |

### Commands (`internal/commands/`)

| Command | Subcommand | Input | Output | Exit Codes |
|---------|------------|-------|--------|------------|
| `config` | `init` | interactive | Setup wizard | 0=OK, 1=Error |
| `config` | `show` | - | Current config | 0=OK |
| `config` | `set` | key value | Confirmation | 0=OK, 1=Error |
| `config` | `path` | - | Config file path | 0=OK |
| `page` | `push` | `.md` file(s) | Created/updated info | 0=OK, 1=Error |
| `page` | `pull` | page ID | `.md` file written | 0=OK, 1=Error |
| `page` | `get` | page ID | Page details | 0=OK, 1=Error |
| `page` | `list` | `--space` | Page list | 0=OK, 1=Error |
| `page` | `delete` | page ID | Confirmation | 0=OK, 1=Error |
| `page` | `tree` | `--space` | Hierarchy tree | 0=OK, 1=Error |
| `space` | `list` | - | Space list | 0=OK, 1=Error |
| `space` | `get` | space key | Space details | 0=OK, 1=Error |
| `search` | - | CQL query | Search results | 0=OK, 1=Error |
| `label` | `add` | page ID, labels | Confirmation | 0=OK, 1=Error |
| `label` | `list` | page ID | Label list | 0=OK, 1=Error |
| `label` | `remove` | page ID, label | Confirmation | 0=OK, 1=Error |
| `attachment` | `upload` | page ID, file | Upload info | 0=OK, 1=Error |
| `attachment` | `list` | page ID | Attachment list | 0=OK, 1=Error |
| `attachment` | `download` | attachment ID | File written | 0=OK, 1=Error |
| `version` | - | - | Version info | 0=OK |

### Output Formatters (`internal/output/`)

```go
type Formatter interface {
    FormatPage(page *models.Page) string
    FormatPages(pages []models.Page) string
    FormatSpace(space *models.Space) string
    FormatSpaces(spaces []models.Space) string
    FormatSearchResults(results *models.SearchResult) string
    FormatLabels(labels []models.Label) string
    FormatAttachments(attachments []models.Attachment) string
    FormatPageTree(tree *models.PageTree) string
}
```

Three implementations:
- `PrettyFormatter` - Beautiful CLI output with colors (default)
- `JSONFormatter` - Compact JSON, one object per line for streaming
- `MarkdownFormatter` - Tables and headers, AI-readable

### Models (`internal/models/`)

```go
// Page represents a Confluence page
type Page struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    SpaceID   string    `json:"spaceId"`
    Status    string    `json:"status"`
    ParentID  string    `json:"parentId,omitempty"`
    ParentType string   `json:"parentType,omitempty"`
    Body      *Body     `json:"body,omitempty"`
    Version   *Version  `json:"version,omitempty"`
    Labels    []Label   `json:"labels,omitempty"`
    CreatedAt time.Time `json:"createdAt"`
    AuthorID  string    `json:"authorId"`
}

// Body contains page content in various formats
type Body struct {
    Storage        *BodyContent `json:"storage,omitempty"`
    AtlasDocFormat *BodyContent `json:"atlas_doc_format,omitempty"`
}

type BodyContent struct {
    Value          string `json:"value"`
    Representation string `json:"representation"`
}

// Version tracks page version for optimistic locking
type Version struct {
    Number    int       `json:"number"`
    Message   string    `json:"message,omitempty"`
    CreatedAt time.Time `json:"createdAt"`
    AuthorID  string    `json:"authorId"`
}

// Space represents a Confluence space
type Space struct {
    ID          string `json:"id"`
    Key         string `json:"key"`
    Name        string `json:"name"`
    Type        string `json:"type"`
    Status      string `json:"status"`
    Description string `json:"description,omitempty"`
    HomepageID  string `json:"homepageId,omitempty"`
}

// Label represents a page label/tag
type Label struct {
    ID     string `json:"id"`
    Prefix string `json:"prefix"`
    Name   string `json:"name"`
}

// Attachment represents a file attached to a page
type Attachment struct {
    ID        string `json:"id"`
    Title     string `json:"title"`
    MediaType string `json:"mediaType"`
    FileSize  int64  `json:"fileSize"`
    Comment   string `json:"comment,omitempty"`
    PageID    string `json:"pageId"`
}

// SearchResult holds CQL search results
type SearchResult struct {
    Results []SearchEntry `json:"results"`
    Start   int           `json:"start"`
    Limit   int           `json:"limit"`
    Size    int           `json:"size"`
}

type SearchEntry struct {
    Content     Page   `json:"content"`
    Title       string `json:"title"`
    Excerpt     string `json:"excerpt"`
    URL         string `json:"url"`
    LastModified string `json:"lastModified"`
}

// CreatePageRequest is the payload for creating a page
type CreatePageRequest struct {
    SpaceID  string `json:"spaceId"`
    Title    string `json:"title"`
    ParentID string `json:"parentId,omitempty"`
    Status   string `json:"status"`
    Body     Body   `json:"body"`
}

// UpdatePageRequest is the payload for updating a page
type UpdatePageRequest struct {
    ID      string  `json:"id"`
    Title   string  `json:"title"`
    Status  string  `json:"status"`
    Body    Body    `json:"body"`
    Version Version `json:"version"`
}

// ListOptions provides pagination and filtering
type ListOptions struct {
    Limit  int    `json:"limit,omitempty"`
    Cursor string `json:"cursor,omitempty"`
    Status string `json:"status,omitempty"`
}

// PageList is a paginated list of pages
type PageList struct {
    Results []Page `json:"results"`
    Links   Links  `json:"_links"`
}

type SpaceList struct {
    Results []Space `json:"results"`
    Links   Links   `json:"_links"`
}

type Links struct {
    Next string `json:"next,omitempty"`
}

// PageTree represents the page hierarchy for tree display
type PageTree struct {
    Page     Page       `json:"page"`
    Children []PageTree `json:"children,omitempty"`
}
```

## Data Flow

### Push (Markdown → Confluence)

```
┌──────────────┐     ┌───────────────┐     ┌────────────────┐
│  .md file    │────▶│  Frontmatter  │────▶│  Markdown →    │
│  (on disk)   │     │  Parser       │     │  Storage XHTML │
└──────────────┘     └───────────────┘     └────────────────┘
                            │                       │
                            ▼                       ▼
                     ┌───────────────┐     ┌────────────────┐
                     │  Resolve      │     │  REST API      │
                     │  Space/Page   │────▶│  Create/Update │
                     └───────────────┘     └────────────────┘
                                                    │
                                                    ▼
                                           ┌────────────────┐
                                           │  Write page_id │
                                           │  to frontmatter│
                                           └────────────────┘
```

### Pull (Confluence → Markdown)

```
┌──────────────┐     ┌───────────────┐     ┌────────────────┐
│  REST API    │────▶│  Page + Meta  │────▶│  Storage XHTML │
│  GET page    │     │  Response     │     │  → Markdown    │
└──────────────┘     └───────────────┘     └────────────────┘
                                                    │
                                                    ▼
                                           ┌────────────────┐
                                           │  Build front-  │
                                           │  matter + body │
                                           └────────────────┘
                                                    │
                                                    ▼
                                           ┌────────────────┐
                                           │  Write .md     │
                                           │  file to disk  │
                                           └────────────────┘
```

### General Command Flow

```
┌─────────────┐     ┌──────────────┐     ┌───────────────┐
│   CLI       │────▶│   Command    │────▶│   API Client  │
│   Input     │     │   Handler    │     │   (REST)      │
└─────────────┘     └──────────────┘     └───────────────┘
                           │                     │
                           ▼                     ▼
                    ┌──────────────┐     ┌───────────────┐
                    │  Formatter   │◀────│   Response    │
                    │  (JSON/MD)   │     │   Parser      │
                    └──────────────┘     └───────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │   stdout     │
                    └──────────────┘
```

## API Integration

### Endpoints

| Type | URL Pattern |
|------|-------------|
| REST v2 (pages/spaces) | `{base_url}/wiki/api/v2/...` |
| REST v1 (labels/attachments/search) | `{base_url}/wiki/rest/api/...` |

### Authentication

All requests include:
```
Authorization: Basic <base64(email:token)>
Content-Type: application/json
```

Attachment uploads use:
```
Content-Type: multipart/form-data
X-Atlassian-Token: nocheck
```

### Rate Limits

- Points-based system (not simple request count)
- Default: 65,000 points per hour (shared global pool)
- Higher tiers: up to 500,000 points/hour based on plan
- HTTP 429 response with `Retry-After` header when exceeded
- Quotas reset at the start of each UTC hour

### Pagination

The v2 API uses cursor-based pagination:

```go
// First request
GET /wiki/api/v2/pages?space-id=123&limit=25

// Subsequent requests use cursor from _links.next
GET /wiki/api/v2/pages?space-id=123&limit=25&cursor=eyJ...
```

## Error Handling

| Category | Strategy |
|----------|----------|
| Network errors | Log and exit with code 1 |
| 401 Unauthorized | Clear message: "Invalid credentials. Check email and API token" |
| 403 Forbidden | "Permission denied for this resource" |
| 404 Not found | "Page/space not found. Verify the ID and your access" |
| 409 Conflict | "Version conflict. Pull latest version first" (stale version on update) |
| 429 Rate limited | "Rate limited. Retry after {n} seconds" |
| Parse errors | Log raw response, exit 1 |
| Converter errors | Clear message with line number if possible |

## Cross-Platform Build

```makefile
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

build-all:
    @for platform in $(PLATFORMS); do \
        GOOS=$${platform%/*} GOARCH=$${platform#*/} \
        go build -o dist/cfluence-$${platform%/*}-$${platform#*/} ./cmd/cfluence; \
    done
```

## Dependencies

| Package | Purpose | Why |
|---------|---------|-----|
| `spf13/cobra` | CLI framework | Industry standard (kubectl, hugo) |
| `spf13/viper` | Config loading | Handles env + file + flags |
| `yuin/goldmark` | Markdown parsing | Extensible, CommonMark-compliant |
| `gopkg.in/yaml.v3` | YAML frontmatter | Standard YAML library |

## Testing Strategy

```
internal/
├── api/
│   ├── client_test.go       # Auth header, error handling
│   ├── pages_test.go        # Mock HTTP: create, get, update, delete, list
│   ├── spaces_test.go       # Mock HTTP: list, get
│   ├── search_test.go       # Mock HTTP: CQL queries
│   ├── labels_test.go       # Mock HTTP: add, list, remove
│   └── attachments_test.go  # Mock HTTP: upload, list, download
├── config/
│   └── config_test.go       # Env priority, file loading, validation
├── converter/
│   ├── markdown_test.go     # Golden file tests (testdata/*.md ↔ *.storage.xml)
│   └── frontmatter_test.go  # Parse, serialize, round-trip
└── output/
    └── formatter_test.go    # Output format verification

testdata/
├── simple_page.md           # Basic headings, paragraphs, lists
├── simple_page.storage.xml  # Expected XHTML output
├── complex_page.md          # Tables, code blocks, images, nested lists
└── complex_page.storage.xml # Expected XHTML output
```

### Test patterns:

- **API tests:** `httptest.NewServer` with canned JSON responses. Verify request method, path, headers, body. Test error status codes.
- **Config tests:** Use `t.Setenv()` for env var tests. Temp files for config file tests. Verify priority chain.
- **Converter tests:** Golden file comparison. Read `testdata/*.md`, convert, compare against `testdata/*.storage.xml`. Round-trip tests (MD → XHTML → MD).
- **Frontmatter tests:** Parse known YAML, verify struct fields. Serialize and verify output. Edge cases (empty fields, no frontmatter).
- **Formatter tests:** Snapshot/golden output comparison for each format type.

## Security Considerations

1. Token and email never logged or printed (masked in `config show`)
2. Config file permissions checked (warn if world-readable)
3. No shell expansion in any path handling
4. TLS verification enabled for all HTTPS requests
5. `X-Atlassian-Token: nocheck` only sent on attachment uploads (CSRF protection)
6. Credentials stored in `~/.confluence/config.yaml` with 0600 permissions
