package provider

import (
	"fmt"
	"time"
)

const (
	defaultMaxRetries         = 15
	defaultRetryInterval      = 90 * time.Second
	maxAllowedMaxRetries      = 100
	maxAllowedRetryInterval   = 300 * time.Second
	statusTooManyRequests     = 429
	statusInternalServerError = 500
	statusServiceUnavailable  = 503
)

// RetryConfig defines retry behavior configuration
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// RetryInterval is the duration to wait between retries
	RetryInterval time.Duration

	// RetryableStatusCodes are HTTP status codes that trigger retry
	RetryableStatusCodes []int

	// EnableLogging enables detailed retry logging
	EnableLogging bool

	// EnableReadinessCheck enables iDRAC readiness check before operations
	EnableReadinessCheck bool
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:           defaultMaxRetries,
		RetryInterval:        defaultRetryInterval,
		RetryableStatusCodes: []int{statusTooManyRequests, statusInternalServerError, statusServiceUnavailable},
		EnableLogging:        true,
		EnableReadinessCheck: true,
	}
}

// Validate checks if the retry configuration is valid
func (c *RetryConfig) Validate() error {
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative, got %d", c.MaxRetries)
	}
	if c.MaxRetries > maxAllowedMaxRetries {
		return fmt.Errorf("max_retries must be <= %d, got %d", maxAllowedMaxRetries, c.MaxRetries)
	}
	if c.RetryInterval < 0 {
		return fmt.Errorf("retry_interval must be non-negative, got %v", c.RetryInterval)
	}
	if c.RetryInterval > maxAllowedRetryInterval {
		return fmt.Errorf("retry_interval must be <= %v, got %v", maxAllowedRetryInterval, c.RetryInterval)
	}
	return nil
}

// TotalTimeout returns the maximum time that retries could take
func (c *RetryConfig) TotalTimeout() time.Duration {
	return time.Duration(c.MaxRetries) * c.RetryInterval
}
