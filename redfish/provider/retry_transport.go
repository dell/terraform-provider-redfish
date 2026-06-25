package provider

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// RetryableTransport wraps http.RoundTripper with retry logic
type RetryableTransport struct {
	transport http.RoundTripper
	config    RetryConfig
}

// NewRetryableTransport creates a new retryable transport
func NewRetryableTransport(transport http.RoundTripper, config RetryConfig) *RetryableTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &RetryableTransport{
		transport: transport,
		config:    config,
	}
}

// RoundTrip implements http.RoundTripper with retry logic
func (t *RetryableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	var lastErr error
	var resp *http.Response

	// Read and store request body for retries
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, lastErr = io.ReadAll(req.Body)
		if lastErr != nil {
			return nil, fmt.Errorf("failed to read request body: %w", lastErr)
		}
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	for attempt := 0; attempt <= t.config.MaxRetries; attempt++ {
		// Clone request for retry (restore body)
		reqClone := req.Clone(ctx)
		if bodyBytes != nil {
			reqClone.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Execute request
		resp, lastErr = t.transport.RoundTrip(reqClone)

		// Check if successful
		if lastErr == nil && !t.shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		// Log the error
		if lastErr != nil {
			tflog.Warn(ctx, "HTTP request failed", map[string]any{
				"attempt": attempt + 1,
				"error":   lastErr.Error(),
				"method":  req.Method,
				"url":     req.URL.String(),
			})
		} else if t.shouldRetry(resp.StatusCode) {
			tflog.Warn(ctx, "HTTP request returned retryable error", map[string]any{
				"attempt":     attempt + 1,
				"status_code": resp.StatusCode,
				"method":      req.Method,
				"url":         req.URL.String(),
			})
			// Drain and close response body to reuse connection
			if resp.Body != nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
		}

		// Don't retry if we've exhausted attempts
		if attempt >= t.config.MaxRetries {
			break
		}

		// Log retry attempt
		t.logRetryAttempt(ctx, attempt+1, lastErr)

		// Wait before retry
		select {
		case <-time.After(t.config.RetryInterval):
			// Continue to next retry
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d retries: %w", t.config.MaxRetries, lastErr)
	}

	return resp, nil
}

// shouldRetry determines if a status code should trigger a retry
func (t *RetryableTransport) shouldRetry(statusCode int) bool {
	for _, code := range t.config.RetryableStatusCodes {
		if statusCode == code {
			return true
		}
	}
	return false
}

// logRetryAttempt logs information about a retry attempt
func (t *RetryableTransport) logRetryAttempt(ctx context.Context, attempt int, err error) {
	if !t.config.EnableLogging {
		return
	}

	retriesRemaining := t.config.MaxRetries - attempt
	timeRemaining := time.Duration(retriesRemaining) * t.config.RetryInterval

	tflog.Info(ctx, "Retrying request", map[string]any{
		"attempt":           attempt,
		"max_retries":       t.config.MaxRetries,
		"retries_remaining": retriesRemaining,
		"retry_interval":    t.config.RetryInterval.String(),
		"time_remaining":    timeRemaining.String(),
		"error":             err.Error(),
	})
}
