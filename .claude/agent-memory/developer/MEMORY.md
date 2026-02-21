# confluence-cli Developer Memory

## Config Package Interface
- `config.Load(cfgFile string) (*Config, error)` — takes explicit path or "" for default
- `config.DefaultConfigPath() string` — returns ~/.confluence/config.yaml
- `config.EnsureConfigDir() error` — creates ~/.confluence/ with 0700, returns error only
- Config fields: BaseURL, Email, Token, Space, Format (all exported)
- Env vars: CONFLUENCE_BASE_URL, CONFLUENCE_EMAIL, CONFLUENCE_TOKEN, CONFLUENCE_SPACE
- Priority: env vars > config file > defaults; default format = "pretty"
- Use `viper.New()` (not global viper) to avoid test pollution across parallel tests

## API Client Patterns
- v2 API base: `{baseURL}/wiki/api/v2`
- v1 API base: `{baseURL}/wiki/rest/api`
- Auth: Basic `base64(email:token)`
- `doJSON` handles marshal/unmarshal; `doRequest` gives raw response for streaming
- Attachments upload needs `X-Atlassian-Token: nocheck` header + multipart/form-data
- `checkResponse` reads body on error; caller must not double-read after error
- `export_test.go` pattern (package api, compiled only during test) exposes unexported methods

## Test Patterns
- `containsString` helper defined in `client_test.go` is shared across all `api_test` files
- Config tests: use `os.Unsetenv` not `t.Setenv("VAR", "")` to clear env vars
- API tests: use `httptest.NewServer`, verify request method/path/query/headers/body

## Converter Package (`internal/converter/`)

### Import aliases — avoid naming conflicts
```go
stdhtml "html"                                   // stdlib html.UnescapeString, html.EscapeString
gmhtml "github.com/yuin/goldmark/renderer/html"  // goldmark WithXHTML(), WithUnsafe()
nethtml "golang.org/x/net/html"                  // tokenizer
```

### Goldmark setup
- `parser.WithAutoHeadingID()` is a valid `parser.Option` for `goldmark.WithParserOptions()`
- Produces `id` attributes on headings (e.g., `<h1 id="my-page">My Page</h1>`)
- Go 1.23 has builtin `max`/`min` — do NOT define custom versions (compile error)

### CDATA in x/net/html tokenizer
- Tokenizer treats `<![CDATA[...]]>` as a bogus CommentToken
- `tokenizer.Raw()` returns full raw bytes INCLUDING angle brackets: `"<![CDATA[content]]>"`
- Check prefix `"<![CDATA["` (9 chars) and suffix `"]]>"` (3 chars)

### Golden test pattern
- Golden tests use `UPDATE_GOLDEN=1` env var to regenerate `.storage.xml` files
- Tests skip (not fail) if golden file doesn't exist
- Run: `UPDATE_GOLDEN=1 go test ./internal/converter/...`
