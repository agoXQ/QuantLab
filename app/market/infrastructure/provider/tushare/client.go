// Package tushare implements a domain.DataProvider backed by Tushare Pro.
//
// Reference: https://tushare.pro/document/2 (POST /api/v1, JSON body)
//
// The client is intentionally minimal: it knows how to call the Tushare HTTP
// gateway, decode the column-major response shape, and surface API-level
// errors. Higher-level mapping into domain aggregates lives in provider.go.
package tushare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	domainErr "github.com/agoXQ/QuantLab/app/market/domain/errors"
)

const (
	defaultEndpoint = "https://api.tushare.pro"
	defaultTimeout  = 15 * time.Second
)

// Client is a thin HTTP client for the Tushare Pro JSON gateway.
type Client struct {
	endpoint string
	token    string
	http     *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithEndpoint overrides the default Tushare endpoint.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) { c.endpoint = endpoint }
}

// WithHTTPClient injects a custom *http.Client (used in tests).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.http = h }
}

// NewClient creates a new Tushare client. token is required.
func NewClient(token string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("tushare token is required")
	}
	c := &Client{
		endpoint: defaultEndpoint,
		token:    token,
		http:     &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// Token returns the configured access token.
func (c *Client) Token() string { return c.token }

// Response is the canonical Tushare envelope.
type Response struct {
	RequestID string       `json:"request_id"`
	Code      int          `json:"code"`
	Msg       string       `json:"msg"`
	Data      ResponseData `json:"data"`
}

// ResponseData holds the column-major payload returned by the Tushare API.
type ResponseData struct {
	Fields []string        `json:"fields"`
	Items  [][]any         `json:"items"`
	HasMore bool           `json:"has_more"`
	Count   int            `json:"count"`
}

// Request is the JSON body sent to the gateway.
type Request struct {
	APIName string         `json:"api_name"`
	Token   string         `json:"token"`
	Params  map[string]any `json:"params"`
	Fields  string         `json:"fields,omitempty"`
}

// Call invokes a single Tushare endpoint and returns the parsed response.
func (c *Client) Call(ctx context.Context, apiName string, params map[string]any, fields string) (*Response, error) {
	if params == nil {
		params = map[string]any{}
	}
	body, err := json.Marshal(Request{APIName: apiName, Token: c.token, Params: params, Fields: fields})
	if err != nil {
		return nil, fmt.Errorf("marshal tushare request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build tushare request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call tushare %s: %w", apiName, domainErr.ErrProviderUnavailable)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read tushare response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tushare http %d: %s", resp.StatusCode, string(raw))
	}

	var parsed Response
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode tushare response: %w", domainErr.ErrProviderResponse)
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("tushare api error %s: %s (code=%d)", apiName, parsed.Msg, parsed.Code)
	}
	return &parsed, nil
}

// Index returns the column index for a named field, or -1 if missing.
func (d ResponseData) Index(name string) int {
	for i, f := range d.Fields {
		if f == name {
			return i
		}
	}
	return -1
}

// String returns the cell as a string.
func cellString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", x)
	}
}

// Float returns the cell as a float, defaulting to 0.
func cellFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case float32:
		return float64(x)
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case json.Number:
		f, _ := x.Float64()
		return f
	case string:
		var f float64
		_, _ = fmt.Sscanf(x, "%f", &f)
		return f
	}
	return 0
}

// Int returns the cell as an int64, defaulting to 0.
func cellInt(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int:
		return int64(x)
	case int64:
		return x
	case string:
		var n int64
		_, _ = fmt.Sscanf(x, "%d", &n)
		return n
	}
	return 0
}
