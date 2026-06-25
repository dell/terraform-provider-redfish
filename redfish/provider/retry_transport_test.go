package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRetryableTransport_Success(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries:           3,
		RetryInterval:        100 * time.Millisecond,
		RetryableStatusCodes: []int{500, 503, 429},
	}

	transport := NewRetryableTransport(http.DefaultTransport, config)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if attempts != 1 {
		t.Fatalf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetryableTransport_RetryOn500(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries:           5,
		RetryInterval:        100 * time.Millisecond,
		RetryableStatusCodes: []int{500, 503, 429},
	}

	transport := NewRetryableTransport(http.DefaultTransport, config)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if attempts != 3 {
		t.Fatalf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableTransport_RetryOn503(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries:           3,
		RetryInterval:        100 * time.Millisecond,
		RetryableStatusCodes: []int{500, 503, 429},
	}

	transport := NewRetryableTransport(http.DefaultTransport, config)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if attempts != 2 {
		t.Fatalf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryableTransport_RetryOn429(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries:           3,
		RetryInterval:        100 * time.Millisecond,
		RetryableStatusCodes: []int{500, 503, 429},
	}

	transport := NewRetryableTransport(http.DefaultTransport, config)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	if attempts != 2 {
		t.Fatalf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryableTransport_MaxRetriesExhausted(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries:           3,
		RetryInterval:        100 * time.Millisecond,
		RetryableStatusCodes: []int{500, 503, 429},
	}

	transport := NewRetryableTransport(http.DefaultTransport, config)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", server.URL, nil)
	_, err := client.Do(req)

	if err == nil {
		t.Fatal("Expected error after max retries, got nil")
	}
	if attempts != 4 { // Initial + 3 retries
		t.Fatalf("Expected 4 attempts, got %d", attempts)
	}
}

func TestRetryableTransport_NoRetryOn404(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries:           3,
		RetryInterval:        100 * time.Millisecond,
		RetryableStatusCodes: []int{500, 503, 429},
	}

	transport := NewRetryableTransport(http.DefaultTransport, config)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected 404, got %d", resp.StatusCode)
	}
	if attempts != 1 {
		t.Fatalf("Expected 1 attempt (no retry), got %d", attempts)
	}
}

func TestRetryableTransport_NoRetryOn400(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries:           3,
		RetryInterval:        100 * time.Millisecond,
		RetryableStatusCodes: []int{500, 503, 429},
	}

	transport := NewRetryableTransport(http.DefaultTransport, config)
	client := &http.Client{Transport: transport}

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400, got %d", resp.StatusCode)
	}
	if attempts != 1 {
		t.Fatalf("Expected 1 attempt (no retry), got %d", attempts)
	}
}

func TestRetryableTransport_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries:           10,
		RetryInterval:        1 * time.Second,
		RetryableStatusCodes: []int{500, 503, 429},
	}

	transport := NewRetryableTransport(http.DefaultTransport, config)
	client := &http.Client{Transport: transport}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
	_, err := client.Do(req)

	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}
	// Check if error contains context deadline exceeded
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("Expected context deadline exceeded error, got %v", err)
	}
}

func TestRetryConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  RetryConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: RetryConfig{
				MaxRetries:    15,
				RetryInterval: 90 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "negative max retries",
			config: RetryConfig{
				MaxRetries:    -1,
				RetryInterval: 90 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "too many retries",
			config: RetryConfig{
				MaxRetries:    101,
				RetryInterval: 90 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative interval",
			config: RetryConfig{
				MaxRetries:    15,
				RetryInterval: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "interval too long",
			config: RetryConfig{
				MaxRetries:    15,
				RetryInterval: 301 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRetryConfig_TotalTimeout(t *testing.T) {
	config := RetryConfig{
		MaxRetries:    15,
		RetryInterval: 90 * time.Second,
	}

	expected := 15 * 90 * time.Second
	if config.TotalTimeout() != expected {
		t.Errorf("TotalTimeout() = %v, want %v", config.TotalTimeout(), expected)
	}
}
