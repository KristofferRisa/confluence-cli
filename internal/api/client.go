package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

// Client is a REST client for the Confluence Cloud API.
type Client struct {
	baseURL    string
	email      string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Confluence API client.
func NewClient(baseURL, email, token string) *Client {
	return &Client{
		baseURL: baseURL,
		email:   email,
		token:   token,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// v2Path returns the full URL for a v2 API endpoint.
func (c *Client) v2Path(path string) string {
	return c.baseURL + "/wiki/api/v2" + path
}

// v1Path returns the full URL for a v1 REST API endpoint.
func (c *Client) v1Path(path string) string {
	return c.baseURL + "/wiki/rest/api" + path
}

// authHeader returns the Basic Auth header value for the configured credentials.
func (c *Client) authHeader() string {
	creds := c.email + ":" + c.token
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))
}

// doRequest builds and executes an HTTP request, returning the raw response.
// The caller is responsible for closing the response body.
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("build request %s %s: %w", method, url, err)
	}

	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request %s %s: %w", method, url, err)
	}

	return resp, nil
}

// doJSON sends a JSON request and decodes the response body into result.
// body may be nil for requests without a payload.
// result may be nil when the response body is not needed (e.g. DELETE 204).
func (c *Client) doJSON(ctx context.Context, method, url string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	resp, err := c.doRequest(ctx, method, url, bodyReader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return err
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}

// apiError represents an error returned by the Confluence API.
type apiError struct {
	statusCode int
	message    string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("confluence API error %d: %s", e.statusCode, e.message)
}

// checkResponse inspects an HTTP response and returns an error if the status
// code indicates failure. It attempts to extract the error message from the
// JSON response body.
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	// Try to extract a human-readable message from the response body.
	message := extractErrorMessage(body, resp.StatusCode)

	return &apiError{
		statusCode: resp.StatusCode,
		message:    message,
	}
}

// extractErrorMessage tries several common Confluence error response shapes
// and falls back to the raw body when none match.
func extractErrorMessage(body []byte, statusCode int) string {
	if len(body) == 0 {
		return http.StatusText(statusCode)
	}

	// v2 API error: {"message":"..."}
	var v2Err struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &v2Err); err == nil && v2Err.Message != "" {
		return v2Err.Message
	}

	// v1 API error: {"statusCode":...,"message":"...","data":{...}}
	var v1Err struct {
		Message string `json:"message"`
		Data    struct {
			Authorized bool     `json:"authorized"`
			Valid      bool     `json:"valid"`
			Errors     []string `json:"errors"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &v1Err); err == nil && v1Err.Message != "" {
		return v1Err.Message
	}

	// Fall back to raw body (trimmed).
	raw := string(body)
	if len(raw) > 200 {
		raw = raw[:200] + "..."
	}
	return raw
}
