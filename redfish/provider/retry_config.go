package provider

import (
	"fmt"
	"time"
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
		MaxRetries:           15,
		RetryInterval:        90 * time.Second,
		RetryableStatusCodes: []int{429, 500, 503},
		EnableLogging:        true,
		EnableReadinessCheck: true,
	}
}

// Validate checks if the retry configuration is valid
func (c *RetryConfig) Validate() error {
	if c.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative, got %d", c.MaxRetries)
	}
	if c.MaxRetries > 100 {
		return fmt.Errorf("max_retries must be <= 100, got %d", c.MaxRetries)
	}
	if c.RetryInterval < 0 {
		return fmt.Errorf("retry_interval must be non-negative, got %v", c.RetryInterval)
	}
	if c.RetryInterval > 300*time.Second {
		return fmt.Errorf("retry_interval must be <= 300 seconds, got %v", c.RetryInterval)
	}
	return nil
}

// TotalTimeout returns the maximum time that retries could take
func (c *RetryConfig) TotalTimeout() time.Duration {
	return time.Duration(c.MaxRetries) * c.RetryInterval
}
