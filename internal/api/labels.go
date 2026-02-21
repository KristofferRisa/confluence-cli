package api

import (
	"context"
	"fmt"

	"github.com/kristofferrisa/confluence-cli/internal/models"
)

// labelRequest is the request body item for adding a label.
type labelRequest struct {
	Prefix string `json:"prefix"`
	Name   string `json:"name"`
}

// GetLabels returns all labels attached to a page.
// GET /wiki/rest/api/content/{id}/label
func (c *Client) GetLabels(ctx context.Context, pageID string) ([]models.Label, error) {
	u := c.v1Path("/content/" + pageID + "/label")

	var list models.LabelList
	if err := c.doJSON(ctx, "GET", u, nil, &list); err != nil {
		return nil, fmt.Errorf("get labels for page %s: %w", pageID, err)
	}
	return list.Results, nil
}

// AddLabels attaches one or more labels to a page.
// POST /wiki/rest/api/content/{id}/label
// Body: [{"prefix": "global", "name": "label1"}, ...]
func (c *Client) AddLabels(ctx context.Context, pageID string, labels []string) error {
	payload := make([]labelRequest, len(labels))
	for i, name := range labels {
		payload[i] = labelRequest{Prefix: "global", Name: name}
	}

	u := c.v1Path("/content/" + pageID + "/label")
	if err := c.doJSON(ctx, "POST", u, payload, nil); err != nil {
		return fmt.Errorf("add labels to page %s: %w", pageID, err)
	}
	return nil
}

// RemoveLabel removes a single label from a page.
// DELETE /wiki/rest/api/content/{id}/label/{label}
func (c *Client) RemoveLabel(ctx context.Context, pageID string, label string) error {
	u := c.v1Path("/content/" + pageID + "/label/" + label)
	if err := c.doJSON(ctx, "DELETE", u, nil, nil); err != nil {
		return fmt.Errorf("remove label %q from page %s: %w", label, pageID, err)
	}
	return nil
}
