package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/kristofferrisa/confluence-cli/internal/models"
)

// CreatePage creates a new page in Confluence.
// POST /wiki/api/v2/pages
func (c *Client) CreatePage(ctx context.Context, req *models.CreatePageRequest) (*models.Page, error) {
	var page models.Page
	if err := c.doJSON(ctx, "POST", c.v2Path("/pages"), req, &page); err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}
	return &page, nil
}

// GetPage retrieves a page by ID with its body in storage format.
// GET /wiki/api/v2/pages/{id}?body-format=storage
func (c *Client) GetPage(ctx context.Context, pageID string) (*models.Page, error) {
	u := c.v2Path("/pages/"+pageID) + "?body-format=storage"
	var page models.Page
	if err := c.doJSON(ctx, "GET", u, nil, &page); err != nil {
		return nil, fmt.Errorf("get page %s: %w", pageID, err)
	}
	return &page, nil
}

// UpdatePage updates an existing page.
// PUT /wiki/api/v2/pages/{id}
func (c *Client) UpdatePage(ctx context.Context, pageID string, req *models.UpdatePageRequest) (*models.Page, error) {
	var page models.Page
	if err := c.doJSON(ctx, "PUT", c.v2Path("/pages/"+pageID), req, &page); err != nil {
		return nil, fmt.Errorf("update page %s: %w", pageID, err)
	}
	return &page, nil
}

// DeletePage deletes a page by ID.
// DELETE /wiki/api/v2/pages/{id}
func (c *Client) DeletePage(ctx context.Context, pageID string) error {
	if err := c.doJSON(ctx, "DELETE", c.v2Path("/pages/"+pageID), nil, nil); err != nil {
		return fmt.Errorf("delete page %s: %w", pageID, err)
	}
	return nil
}

// ListPages returns a paginated list of pages in a space.
// GET /wiki/api/v2/spaces/{spaceID}/pages?limit=N&cursor=X&status=Y
func (c *Client) ListPages(ctx context.Context, spaceID string, opts *models.ListOptions) (*models.PageList, error) {
	u := c.v2Path("/spaces/" + spaceID + "/pages")
	u = applyListOptions(u, opts)

	var list models.PageList
	if err := c.doJSON(ctx, "GET", u, nil, &list); err != nil {
		return nil, fmt.Errorf("list pages in space %s: %w", spaceID, err)
	}
	return &list, nil
}

// GetPageChildren returns a paginated list of a page's direct children.
// GET /wiki/api/v2/pages/{id}/children?limit=N
func (c *Client) GetPageChildren(ctx context.Context, pageID string, opts *models.ListOptions) (*models.PageList, error) {
	u := c.v2Path("/pages/" + pageID + "/children")
	u = applyListOptions(u, opts)

	var list models.PageList
	if err := c.doJSON(ctx, "GET", u, nil, &list); err != nil {
		return nil, fmt.Errorf("get children for page %s: %w", pageID, err)
	}
	return &list, nil
}

// applyListOptions appends pagination/filter query parameters to a URL string.
func applyListOptions(rawURL string, opts *models.ListOptions) string {
	if opts == nil {
		return rawURL
	}

	params := url.Values{}

	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		params.Set("cursor", opts.Cursor)
	}
	if opts.Status != "" {
		params.Set("status", opts.Status)
	}

	if len(params) == 0 {
		return rawURL
	}

	return rawURL + "?" + params.Encode()
}
