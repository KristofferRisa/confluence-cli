package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/kristofferrisa/confluence-cli/internal/models"
)

// Search executes a CQL query against the Confluence search API.
// GET /wiki/rest/api/search?cql={cql}&limit=N&start=N
func (c *Client) Search(ctx context.Context, cql string, opts *models.ListOptions) (*models.SearchResult, error) {
	params := url.Values{}
	params.Set("cql", cql)

	if opts != nil {
		if opts.Limit > 0 {
			params.Set("limit", strconv.Itoa(opts.Limit))
		}
		// v1 search uses "start" (integer offset), not a cursor.
		// Map Cursor to start when it is a numeric string; otherwise ignore.
		if opts.Cursor != "" {
			params.Set("start", opts.Cursor)
		}
	}

	u := c.v1Path("/search") + "?" + params.Encode()

	var result models.SearchResult
	if err := c.doJSON(ctx, "GET", u, nil, &result); err != nil {
		return nil, fmt.Errorf("search (cql=%q): %w", cql, err)
	}
	return &result, nil
}
