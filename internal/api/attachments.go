package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/kristofferrisa/confluence-cli/internal/models"
)

// ListAttachments returns all attachments for a page.
// GET /wiki/rest/api/content/{id}/child/attachment
func (c *Client) ListAttachments(ctx context.Context, pageID string) ([]models.Attachment, error) {
	u := c.v1Path("/content/" + pageID + "/child/attachment")

	var list models.AttachmentList
	if err := c.doJSON(ctx, "GET", u, nil, &list); err != nil {
		return nil, fmt.Errorf("list attachments for page %s: %w", pageID, err)
	}
	return list.Results, nil
}

// UploadAttachment uploads a file as an attachment to a page.
// POST /wiki/rest/api/content/{id}/child/attachment
// The request is a multipart/form-data upload.
// Confluence requires the X-Atlassian-Token: nocheck header to bypass XSRF.
func (c *Client) UploadAttachment(ctx context.Context, pageID, filename string, reader io.Reader) (*models.Attachment, error) {
	// Build multipart body.
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}

	if _, err := io.Copy(fw, reader); err != nil {
		return nil, fmt.Errorf("write file to form: %w", err)
	}

	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	u := c.v1Path("/content/" + pageID + "/child/attachment")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &buf)
	if err != nil {
		return nil, fmt.Errorf("build upload request: %w", err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	// Confluence requires this header for multipart uploads to bypass XSRF protection.
	req.Header.Set("X-Atlassian-Token", "nocheck")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute upload request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, fmt.Errorf("upload attachment to page %s: %w", pageID, err)
	}

	// The v1 API wraps results in a list even for a single attachment upload.
	var list models.AttachmentList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode upload response: %w", err)
	}

	if len(list.Results) == 0 {
		return nil, fmt.Errorf("upload attachment: empty result set in response")
	}

	return &list.Results[0], nil
}

// DownloadAttachment streams the attachment content to writer.
// downloadPath is the value from AttachmentLinks.Download (e.g. "/wiki/download/attachments/...").
// The full URL is constructed as baseURL + downloadPath.
func (c *Client) DownloadAttachment(ctx context.Context, downloadPath string, writer io.Writer) error {
	u := c.baseURL + downloadPath

	resp, err := c.doRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("download attachment: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return fmt.Errorf("download attachment: %w", err)
	}

	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("write attachment data: %w", err)
	}

	return nil
}
