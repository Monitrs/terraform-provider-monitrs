// Package api is the internal HTTP client used by the Terraform provider.
// It is NOT a general-purpose Go SDK — that would live in its own module.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://monit.rs"
	defaultTimeout = 30 * time.Second
	defaultUA      = "terraform-provider-monitrs"
	defaultRetries = 3
)

type Client struct {
	baseURL    string
	apiKey     string
	http       *http.Client
	userAgent  string
	maxRetries int
}

type Option func(*Client)

func WithBaseURL(u string) Option         { return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") } }
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }
func WithUserAgent(ua string) Option      { return func(c *Client) { c.userAgent = ua } }
func WithMaxRetries(n int) Option         { return func(c *Client) { c.maxRetries = n } }

func New(apiKey string, opts ...Option) (*Client, error) {
	if !strings.HasPrefix(apiKey, "mrs_") {
		return nil, errors.New("api_key must start with 'mrs_' (monit.rs API key). Get one at https://monit.rs/settings/api-keys")
	}
	c := &Client{
		baseURL:    defaultBaseURL,
		apiKey:     apiKey,
		http:       &http.Client{Timeout: defaultTimeout},
		userAgent:  defaultUA,
		maxRetries: defaultRetries,
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	url := c.baseURL + path

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		var reqBody io.Reader
		if body != nil {
			buf, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("marshal request body: %w", err)
			}
			reqBody = bytes.NewReader(buf)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("User-Agent", c.userAgent)
		req.Header.Set("Accept", "application/json")
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.http.Do(req)
		if err != nil {
			// Transient network error — backoff + retry
			if attempt < c.maxRetries {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
				continue
			}
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNoContent {
			return nil
		}

		if resp.StatusCode >= 400 {
			apiErr := parseErrorResponse(resp)
			if apiErr.IsRateLimited() && attempt < c.maxRetries {
				wait := time.Duration(apiErr.RetryAfterSeconds) * time.Second
				if wait <= 0 {
					wait = time.Second
				}
				time.Sleep(wait)
				continue
			}
			return apiErr
		}

		if out != nil {
			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("read response body: %w", err)
			}
			if err := json.Unmarshal(raw, out); err != nil {
				return fmt.Errorf("unmarshal response: %w", err)
			}
		}
		return nil
	}
	return errors.New("exhausted retries")
}

// ── Endpoint methods ─────────────────────────────────────────────────

func (c *Client) CreateEndpoint(ctx context.Context, body EndpointCreate) (*Endpoint, error) {
	var out Endpoint
	if err := c.do(ctx, "POST", "/api/v1/endpoints", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetEndpoint(ctx context.Context, id string) (*Endpoint, error) {
	var out Endpoint
	if err := c.do(ctx, "GET", "/api/v1/endpoints/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateEndpoint(ctx context.Context, id string, body EndpointUpdate) (*Endpoint, error) {
	var out Endpoint
	if err := c.do(ctx, "PATCH", "/api/v1/endpoints/"+id, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteEndpoint(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", "/api/v1/endpoints/"+id, nil, nil)
}

func (c *Client) ListEndpoints(ctx context.Context, limit int) ([]Endpoint, error) {
	if limit <= 0 {
		limit = 50
	}
	var wrap struct {
		Data []Endpoint `json:"data"`
	}
	if err := c.do(ctx, "GET", fmt.Sprintf("/api/v1/endpoints?limit=%d", limit), nil, &wrap); err != nil {
		return nil, err
	}
	return wrap.Data, nil
}

// ── Status page methods ──────────────────────────────────────────────

func (c *Client) CreateStatusPage(ctx context.Context, body StatusPageCreate) (*StatusPage, error) {
	var out StatusPage
	if err := c.do(ctx, "POST", "/api/v1/status-pages", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetStatusPage(ctx context.Context, id string) (*StatusPage, error) {
	var out StatusPage
	if err := c.do(ctx, "GET", "/api/v1/status-pages/"+id, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateStatusPage(ctx context.Context, id string, body map[string]any) (*StatusPage, error) {
	var out StatusPage
	if err := c.do(ctx, "PATCH", "/api/v1/status-pages/"+id, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteStatusPage(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", "/api/v1/status-pages/"+id, nil, nil)
}

func (c *Client) AttachEndpoint(ctx context.Context, pageID string, req struct {
	EndpointID   string `json:"endpoint_id"`
	DisplayName  string `json:"display_name"`
	DisplayOrder *int   `json:"display_order,omitempty"`
}) (*StatusPageAttachment, error) {
	var out StatusPageAttachment
	if err := c.do(ctx, "POST", "/api/v1/status-pages/"+pageID+"/endpoints", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DetachEndpoint(ctx context.Context, pageID, attachmentID string) error {
	return c.do(ctx, "DELETE", "/api/v1/status-pages/"+pageID+"/endpoints/"+attachmentID, nil, nil)
}

// ── Alert channel methods ────────────────────────────────────────────

func (c *Client) CreateAlertChannel(ctx context.Context, body AlertChannelCreate) (*AlertChannel, error) {
	var out AlertChannel
	if err := c.do(ctx, "POST", "/api/v1/alerts/channels", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteAlertChannel(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", "/api/v1/alerts/channels/"+id, nil, nil)
}

func (c *Client) ListAlertChannels(ctx context.Context) ([]AlertChannel, error) {
	var out []AlertChannel
	if err := c.do(ctx, "GET", "/api/v1/alerts/channels", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ── /me ──────────────────────────────────────────────────────────────

func (c *Client) Me(ctx context.Context) (*Me, error) {
	var out Me
	if err := c.do(ctx, "GET", "/api/v1/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
