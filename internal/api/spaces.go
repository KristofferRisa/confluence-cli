package api

import (
	"context"
	"fmt"

	"github.com/kristofferrisa/confluence-cli/internal/models"
)

// ListSpaces returns a paginated list of all spaces.
// GET /wiki/api/v2/spaces?limit=N&cursor=X
func (c *Client) ListSpaces(ctx context.Context, opts *models.ListOptions) (*models.SpaceList, error) {
	u := applyListOptions(c.v2Path("/spaces"), opts)

	var list models.SpaceList
	if err := c.doJSON(ctx, "GET", u, nil, &list); err != nil {
		return nil, fmt.Errorf("list spaces: %w", err)
	}
	return &list, nil
}

// GetSpace retrieves a single space by its ID.
// GET /wiki/api/v2/spaces/{id}
func (c *Client) GetSpace(ctx context.Context, spaceID string) (*models.Space, error) {
	var space models.Space
	if err := c.doJSON(ctx, "GET", c.v2Path("/spaces/"+spaceID), nil, &space); err != nil {
		return nil, fmt.Errorf("get space %s: %w", spaceID, err)
	}
	return &space, nil
}

// GetSpaceByKey looks up a space by its key and returns the first match.
// GET /wiki/api/v2/spaces?keys={key}
func (c *Client) GetSpaceByKey(ctx context.Context, key string) (*models.Space, error) {
	u := c.v2Path("/spaces") + "?keys=" + key

	var list models.SpaceList
	if err := c.doJSON(ctx, "GET", u, nil, &list); err != nil {
		return nil, fmt.Errorf("get space by key %q: %w", key, err)
	}

	if len(list.Results) == 0 {
		return nil, fmt.Errorf("space with key %q not found", key)
	}

	return &list.Results[0], nil
}
