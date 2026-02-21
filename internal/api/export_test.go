// export_test.go exposes internal helpers for white-box testing.
// This file is only compiled during `go test`.
package api

import (
	"context"
)

// DoJSONExported wraps Client.doJSON so test files in package api_test can
// exercise the unexported method without leaking it into production code.
func DoJSONExported(c *Client, ctx context.Context, method, url string, body, result interface{}) error {
	return c.doJSON(ctx, method, url, body, result)
}
