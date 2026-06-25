package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// IDRACReadinessChecker checks if iDRAC is ready for operations
type IDRACReadinessChecker struct {
	endpoint      string
	username      string
	password      string
	insecure      bool
	maxRetries    int
	retryInterval time.Duration
	httpClient    *http.Client
}

// IDRACStatus represents the status response from GetRemoteServicesAPIStatus
type IDRACStatus struct {
	Status                 string `json:"Status"`
	LCStatus               string `json:"LCStatus"`
	RTStatus               string `json:"RTStatus"`
	RedfishStatus          string `json:"RedfishStatus"`
	SEKMServiceStatus      string `json:"SEKMServiceStatus"`
	ServerStatus           string `json:"ServerStatus"`
	TelemetryServiceStatus string `json:"TelemetryServiceStatus"`
}

// NewIDRACReadinessChecker creates a new readiness checker
func NewIDRACReadinessChecker(endpoint, username, password string, insecure bool, config RetryConfig) *IDRACReadinessChecker {
	return &IDRACReadinessChecker{
		endpoint:      endpoint,
		username:      username,
		password:      password,
		insecure:      insecure,
		maxRetries:    config.MaxRetries,
		retryInterval: config.RetryInterval,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WaitForReady waits for iDRAC to be ready
func (c *IDRACReadinessChecker) WaitForReady(ctx context.Context) error {
	tflog.Info(ctx, "Checking iDRAC readiness", map[string]any{
		"endpoint":       c.endpoint,
		"max_retries":    c.maxRetries,
		"retry_interval": c.retryInterval.String(),
	})

	// First check if Redfish service is available
	if err := c.checkRedfishServiceAvailability(ctx); err != nil {
		return fmt.Errorf("Redfish service not available: %w", err)
	}

	// Then check iDRAC readiness status
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		status, err := c.CheckStatus(ctx)
		if err != nil {
			// If GetRemoteServicesAPIStatus is not available, log warning and continue
			if isNotFoundError(err) {
				tflog.Warn(ctx, "GetRemoteServicesAPIStatus API not available, skipping readiness check", map[string]any{
					"error": err.Error(),
				})
				return nil
			}

			tflog.Warn(ctx, "Failed to check iDRAC status", map[string]any{
				"attempt": attempt + 1,
				"error":   err.Error(),
			})
		} else if c.IsReady(status) {
			tflog.Info(ctx, "iDRAC is ready", map[string]any{
				"status":      status.Status,
				"lc_status":   status.LCStatus,
				"redfish_status": status.RedfishStatus,
			})
			return nil
		} else {
			tflog.Info(ctx, "iDRAC not ready yet", map[string]any{
				"attempt":     attempt + 1,
				"status":      status.Status,
				"lc_status":   status.LCStatus,
				"server_status": status.ServerStatus,
			})
		}

		// Don't wait after last attempt
		if attempt >= c.maxRetries {
			break
		}

		// Log progress
		retriesRemaining := c.maxRetries - attempt
		timeRemaining := time.Duration(retriesRemaining) * c.retryInterval
		tflog.Info(ctx, "Waiting for iDRAC to be ready", map[string]any{
			"retry_in":       c.retryInterval.String(),
			"time_remaining": timeRemaining.String(),
		})

		// Wait before next check
		select {
		case <-time.After(c.retryInterval):
			// Continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("iDRAC not ready after %d attempts (~%v)", c.maxRetries+1, c.retryInterval*time.Duration(c.maxRetries))
}

// checkRedfishServiceAvailability checks if Redfish service is available
func (c *IDRACReadinessChecker) checkRedfishServiceAvailability(ctx context.Context) error {
	url := fmt.Sprintf("%s/redfish/v1", c.endpoint)

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.SetBasicAuth(c.username, c.password)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			tflog.Warn(ctx, "Redfish service check failed", map[string]any{
				"attempt": attempt + 1,
				"error":   err.Error(),
			})
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				tflog.Info(ctx, "Redfish service is available")
				return nil
			}
			tflog.Warn(ctx, "Redfish service returned non-OK status", map[string]any{
				"attempt":     attempt + 1,
				"status_code": resp.StatusCode,
			})
		}

		// Don't wait after last attempt
		if attempt >= c.maxRetries {
			break
		}

		// Wait before retry
		select {
		case <-time.After(c.retryInterval):
			// Continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("Redfish service not available after %d attempts", c.maxRetries+1)
}

// CheckStatus checks the current iDRAC status
func (c *IDRACReadinessChecker) CheckStatus(ctx context.Context) (*IDRACStatus, error) {
	url := fmt.Sprintf("%s/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/DellLCService/Actions/DellLCService.GetRemoteServicesAPIStatus", c.endpoint)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &NotFoundError{Message: "GetRemoteServicesAPIStatus API not found"}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var status IDRACStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &status, nil
}

// IsReady checks if the iDRAC status indicates readiness
func (c *IDRACReadinessChecker) IsReady(status *IDRACStatus) bool {
	if status == nil {
		return false
	}
	return status.Status == "Ready"
}

// NotFoundError represents a 404 error
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// isNotFoundError checks if an error is a NotFoundError
func isNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
