package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Response wraps a successful API response.
type Response struct {
	StatusCode int
	Body       []byte
}

// Options configures the Client behavior per-session.
type Options struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
	NoRetry bool
	Version string
}

// Client is the shared HTTP client for all Nightwatch API calls. It handles
// authentication, retries with jittered exponential backoff, and structured
// error translation.
type Client struct {
	httpClient *http.Client
	apiKey     string
	baseURL    string
	noRetry    bool
	version    string
	timeout    time.Duration
	maxRetries int

	// retrySleepFn waits for the given delay or returns early if the context
	// is cancelled. Defaults to contextSleep but can be overridden in tests
	// to skip actual delays.
	retrySleepFn func(context.Context, time.Duration) error
}

// New creates a Client from the given options. The returned client is safe
// for concurrent use across multiple commands.
func New(opts Options) *Client {
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	version := opts.Version
	if version == "" {
		version = "dev"
	}

	return &Client{
		httpClient:   &http.Client{},
		apiKey:       opts.APIKey,
		baseURL:      strings.TrimRight(opts.BaseURL, "/"),
		noRetry:      opts.NoRetry,
		version:      version,
		timeout:      timeout,
		maxRetries:   3,
		retrySleepFn: contextSleep,
	}
}

// resolve returns path unchanged if it is already an absolute URL,
// otherwise it prepends baseURL.
func (c *Client) resolve(path string) string {
	if strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://") {
		return path
	}
	return c.baseURL + path
}

// Get performs an authenticated GET request. path can be relative (prepends
// baseURL) or an absolute URL for endpoints on a different host.
func (c *Client) Get(ctx context.Context, path string, query url.Values) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	u := c.resolve(path)
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	return c.doWithRetry(ctx, http.MethodGet, u, nil)
}

// Post performs an authenticated POST request with a JSON body. path can be
// relative or absolute.
func (c *Client) Post(ctx context.Context, path string, body []byte) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.doWithRetry(ctx, http.MethodPost, c.resolve(path), body)
}

// Put performs an authenticated PUT request with a JSON body. path can be
// relative or absolute.
func (c *Client) Put(ctx context.Context, path string, body []byte) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.doWithRetry(ctx, http.MethodPut, c.resolve(path), body)
}

// Delete performs an authenticated DELETE request. path can be relative or
// absolute.
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	return c.doWithRetry(ctx, http.MethodDelete, c.resolve(path), nil)
}

// setHeaders applies common headers to all API requests.
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "nightwatch-cli/"+c.version)
}

// injectAccessToken appends the API key as an access_token query parameter
// to the request URL. The Nightwatch API uses query-param auth.
func (c *Client) injectAccessToken(req *http.Request) {
	q := req.URL.Query()
	q.Set("access_token", c.apiKey)
	req.URL.RawQuery = q.Encode()
}

// doWithRetry builds a fresh request on each attempt and retries on 429
// responses with jittered exponential backoff unless --no-retry is set.
// A fresh request is needed because the body reader is consumed on each attempt.
func (c *Client) doWithRetry(ctx context.Context, method, u string, reqBody []byte) (*Response, error) {
	var lastErr *APIError

	attempts := 1 + c.maxRetries
	if c.noRetry {
		attempts = 1
	}

	for attempt := range attempts {
		var bodyReader io.Reader
		if reqBody != nil {
			bodyReader = bytes.NewReader(reqBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
		if err != nil {
			return nil, networkError(err)
		}
		c.setHeaders(req)
		c.injectAccessToken(req)
		if reqBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, networkError(err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, networkError(readErr)
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return &Response{
				StatusCode: resp.StatusCode,
				Body:       body,
			}, nil
		}

		apiErr := statusToError(resp.StatusCode, extractErrorMessage(body))

		if resp.StatusCode != 429 {
			return nil, apiErr
		}

		// 429: retry unless disabled or max attempts reached.
		lastErr = apiErr
		if attempt < attempts-1 {
			delay := retryDelay(attempt, resp.Header)
			if err := c.retrySleep(ctx, delay); err != nil {
				return nil, networkError(err)
			}
		}
	}

	// All retries exhausted for 429.
	if c.noRetry {
		return nil, lastErr
	}
	return nil, &APIError{
		Message:   fmt.Sprintf("Rate limited after %d retries", c.maxRetries),
		ExitCode:  ExitRateLimit,
		ErrorType: ErrorTypeRateLimit,
	}
}

// retrySleep waits for the given delay or until the context is cancelled,
// whichever comes first.
func (c *Client) retrySleep(ctx context.Context, delay time.Duration) error {
	return c.retrySleepFn(ctx, delay)
}

// contextSleep waits for the given delay or until the context is cancelled.
func contextSleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// retryDelay computes the delay for the given retry attempt using jittered
// exponential backoff: base * 2^attempt +/- 25%. If the Retry-After header
// contains a valid integer, that value (in seconds) is used instead.
func retryDelay(attempt int, header http.Header) time.Duration {
	if ra := header.Get("Retry-After"); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
	}

	baseSeconds := math.Pow(2, float64(attempt))
	jitter := 0.75 + rand.Float64()*0.5
	return time.Duration(baseSeconds*jitter*1000) * time.Millisecond
}

// extractErrorMessage attempts to extract a human-readable error message
// from a JSON response body.
func extractErrorMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var structured struct {
		Detail any    `json:"detail"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(body, &structured); err == nil {
		if structured.Error != "" {
			return structured.Error
		}
		if structured.Detail != nil {
			switch v := structured.Detail.(type) {
			case string:
				return v
			default:
				detailBytes, _ := json.Marshal(v)
				return string(detailBytes)
			}
		}
	}

	return strings.TrimSpace(string(body))
}
