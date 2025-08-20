// Package realitydefender provides a client for the Reality Defender API
// for detecting deepfakes and manipulated media.
package realitydefender

import (
	"context"
	"errors"
	"sync"
	"time"
)

// ErrorCode represents error codes returned by the SDK
type ErrorCode string

// Error codes returned by the SDK
const (
	ErrorCodeUnauthorized   ErrorCode = "unauthorized"    // Invalid or missing API key
	ErrorCodeInvalidRequest ErrorCode = "invalid_request" // Request format error
	ErrorCodeServerError    ErrorCode = "server_error"    // Server-side error occurred
	ErrorCodeTimeout        ErrorCode = "timeout"         // Operation timed out
	ErrorCodeInvalidFile    ErrorCode = "invalid_file"    // File not found or invalid format
	ErrorCodeFileTooLarge   ErrorCode = "file_too_large"  // File is too large
	ErrorCodeUploadFailed   ErrorCode = "upload_failed"   // Failed to upload the file
	ErrorCodeNotFound       ErrorCode = "not_found"       // Requested resource not found
	ErrorCodeUnknownError   ErrorCode = "unknown_error"   // Unexpected error
)

// Default configuration values
const (
	DefaultPollingInterval = 2000  // Default polling interval in milliseconds
	DefaultTimeout         = 60000 // Default timeout in milliseconds
	DefaultBaseURL         = "https://api.prd.realitydefender.xyz"
)

// SDKError represents an error returned by the SDK
type SDKError struct {
	Message string
	Code    ErrorCode
}

// Error implements the error interface
func (e *SDKError) Error() string {
	return e.Message + " (Code: " + string(e.Code) + ")"
}

// Config represents configuration options for the Reality Defender SDK
type Config struct {
	// APIKey is the authentication key for the API (required)
	APIKey string
	// BaseURL is the optional custom base URL for the API (defaults to production)
	BaseURL string
}

// Client is the main SDK client for interacting with the Reality Defender API
type Client struct {
	apiKey      string
	baseURL     string
	httpClient  *httpClient
	eventsMutex sync.RWMutex
	handlers    map[string][]EventHandler
}

// EventHandler is a function that handles SDK events
type EventHandler func(interface{})

// New creates a new Reality Defender SDK client
func New(config Config) (*Client, error) {
	if config.APIKey == "" {
		return nil, &SDKError{
			Message: "API key is required",
			Code:    ErrorCodeUnauthorized,
		}
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	client := &Client{
		apiKey:   config.APIKey,
		baseURL:  baseURL,
		handlers: make(map[string][]EventHandler),
	}

	client.httpClient = newHTTPClient(&httpClientConfig{
		apiKey:  config.APIKey,
		baseURL: baseURL,
	})

	return client, nil
}

// On registers an event handler for the specified event
func (c *Client) On(event string, handler EventHandler) {
	c.eventsMutex.Lock()
	defer c.eventsMutex.Unlock()

	if _, exists := c.handlers[event]; !exists {
		c.handlers[event] = []EventHandler{}
	}

	c.handlers[event] = append(c.handlers[event], handler)
}

// emit triggers handlers for the specified event
func (c *Client) emit(event string, data interface{}) {
	c.eventsMutex.RLock()
	defer c.eventsMutex.RUnlock()

	handlers, exists := c.handlers[event]
	if !exists {
		return
	}

	for _, handler := range handlers {
		handler(data)
	}
}

// Upload uploads a file to Reality Defender for analysis
func (c *Client) Upload(ctx context.Context, options UploadOptions) (*UploadResult, error) {
	result, err := uploadFile(ctx, c.httpClient, options)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UploadSocialMedia uploads a social media link to Reality Defender for analysis
func (c *Client) UploadSocialMedia(ctx context.Context, options UploadSocialMediaOptions) (*UploadResult, error) {
	result, err := uploadSocialMediaLink(ctx, c.httpClient, options)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetResult gets the detection result for a specific request ID
func (c *Client) GetResult(ctx context.Context, requestID string, options *GetResultOptions) (*DetectionResult, error) {
	if options == nil {
		options = &GetResultOptions{}
	}

	return getDetectionResult(ctx, c.httpClient, requestID, *options)
}

// GetResults queries the detection results stored in the platform
func (c *Client) GetResults(ctx context.Context, pageNumber *int, size *int, name *string, startDate *time.Time, endDate *time.Time, options *GetResultOptions) (*DetectionResultList, error) {
	if options == nil {
		options = &GetResultOptions{}
	}

	return getDetectionResults(ctx, c.httpClient, pageNumber, size, name, startDate, endDate, *options)
}

// PollForResults starts polling for results with event-based callback
func (c *Client) PollForResults(ctx context.Context, requestID string, options *PollOptions) error {
	pollingInterval := DefaultPollingInterval
	timeout := DefaultTimeout

	if options != nil {
		if options.PollingInterval > 0 {
			pollingInterval = options.PollingInterval
		}
		if options.Timeout > 0 {
			timeout = options.Timeout
		}
	}

	return c.pollForResults(ctx, requestID, pollingInterval, timeout)
}

// pollForResults is the internal implementation of polling for results
func (c *Client) pollForResults(ctx context.Context, requestID string, pollingInterval, timeout int) error {
	elapsed := 0
	maxWaitTime := timeout
	isCompleted := false

	// Check if timeout is already zero/expired before starting
	if timeout <= 0 {
		c.emit("error", &SDKError{
			Message: "Polling timeout exceeded",
			Code:    ErrorCodeTimeout,
		})
		return errors.New("polling timeout exceeded")
	}

	for !isCompleted && elapsed < maxWaitTime {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			result, err := c.GetResult(ctx, requestID, nil)

			if err != nil {
				var sdkErr *SDKError
				if errors.As(err, &sdkErr) && sdkErr.Code == ErrorCodeNotFound {
					// Result not ready yet, continue polling if we haven't exceeded the timeout
					elapsed += pollingInterval
					time.Sleep(time.Duration(pollingInterval) * time.Millisecond)
					continue
				}

				// Any other error is emitted and polling stops
				isCompleted = true
				c.emit("error", err)
				return err
			}

			// If the status is still ANALYZING and we haven't exceeded the timeout,
			// continue polling after a delay
			if result.Status == "ANALYZING" {
				elapsed += pollingInterval
				time.Sleep(time.Duration(pollingInterval) * time.Millisecond)
			} else {
				// We have a final result
				isCompleted = true
				c.emit("result", result)
			}
		}
	}

	// Check if we timed out
	if !isCompleted && elapsed >= maxWaitTime {
		timeoutErr := &SDKError{
			Message: "Polling timeout exceeded",
			Code:    ErrorCodeTimeout,
		}
		c.emit("error", timeoutErr)
		return timeoutErr
	}

	return nil
}

// DetectFile is a convenience method to upload and detect a file in one step
func (c *Client) DetectFile(ctx context.Context, filePath string) (*DetectionResult, error) {
	uploadResult, err := c.Upload(ctx, UploadOptions{
		FilePath: filePath,
	})
	if err != nil {
		return nil, err
	}

	return c.GetResult(ctx, uploadResult.RequestID, nil)
}
