package transport

import (
	"context"
	"io"
	"net/http"
	"time"
)

// HTTPClientConfig holds optional HTTP client settings for Pulse scenarios.
type HTTPClientConfig struct {
	Timeout time.Duration
	Headers map[string]string
}

// HTTPClient is the minimal HTTP transport for Pulse scenarios.
type HTTPClient struct {
	client  *http.Client
	headers map[string]string
}

// NewHTTPClient creates an HTTP client backed by the default net/http client.
func NewHTTPClient() *HTTPClient {
	return NewHTTPClientWith(HTTPClientConfig{})
}

// NewHTTPClientWith builds a client using the given config. A zero config
// matches NewHTTPClient: default client and no extra headers.
func NewHTTPClientWith(cfg HTTPClientConfig) *HTTPClient {
	if cfg.Timeout == 0 && len(cfg.Headers) == 0 {
		return &HTTPClient{client: http.DefaultClient}
	}
	return &HTTPClient{
		client:  &http.Client{Timeout: cfg.Timeout},
		headers: cloneHeaderMap(cfg.Headers),
	}
}

func cloneHeaderMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// Get performs an HTTP GET request. On success it returns the response status
// code and a nil error. If the request fails before a response is received,
// the status code is 0.
func (c *HTTPClient) Get(ctx context.Context, url string) (int, error) {
	return c.do(ctx, http.MethodGet, url, nil)
}

// Post performs an HTTP POST request with the provided body. On success it
// returns the response status code and a nil error. If the request fails
// before a response is received, the status code is 0.
func (c *HTTPClient) Post(ctx context.Context, url string, body io.Reader) (int, error) {
	return c.do(ctx, http.MethodPost, url, body)
}

func (c *HTTPClient) do(ctx context.Context, method, url string, body io.Reader) (int, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, err
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return 0, err
	}
	defer resp.Body.Close()

	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return resp.StatusCode, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return resp.StatusCode, &HTTPStatusError{StatusCode: resp.StatusCode}
	}

	return resp.StatusCode, nil
}
