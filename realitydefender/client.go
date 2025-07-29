package realitydefender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// httpClientConfig represents configuration for the HTTP client
type httpClientConfig struct {
	apiKey  string
	baseURL string
}

// httpClient manages HTTP communication with the Reality Defender API
type httpClient struct {
	config     *httpClientConfig
	httpClient *http.Client
}

// newHTTPClient creates a new HTTP client for the Reality Defender API
func newHTTPClient(config *httpClientConfig) *httpClient {
	return &httpClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// get performs a GET request to the specified endpoint
func (c *httpClient) get(ctx context.Context, endpoint string, parameters map[string]string) ([]byte, error) {
	var queryString = ""
	if parameters != nil {
		queryValues := url.Values{}

		for key, value := range parameters {
			queryValues.Add(key, value)
		}
		queryString += "?" + queryValues.Encode()
	}

	getURL := fmt.Sprintf("%s%s%s", c.config.baseURL, endpoint, queryString)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getURL, nil)
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("failed to create request: %v", err),
			Code:    ErrorCodeUnknownError,
		}
	}

	req.Header.Set("X-API-KEY", c.config.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("request failed: %v", err),
			Code:    ErrorCodeServerError,
		}
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// post performs a POST request to the specified endpoint with JSON data
func (c *httpClient) post(ctx context.Context, endpoint string, data interface{}) ([]byte, error) {
	postURL := c.config.baseURL + endpoint

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("failed to marshal JSON: %v", err),
			Code:    ErrorCodeUnknownError,
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("failed to create request: %v", err),
			Code:    ErrorCodeUnknownError,
		}
	}

	req.Header.Set("X-API-KEY", c.config.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("request failed: %v", err),
			Code:    ErrorCodeServerError,
		}
	}
	defer resp.Body.Close()

	return handleResponse(resp)
}

// put performs a PUT request to upload data to the specified URL
func (c *httpClient) put(ctx context.Context, url string, data []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return &SDKError{
			Message: fmt.Sprintf("failed to create request: %v", err),
			Code:    ErrorCodeUnknownError,
		}
	}

	// Set content type to application/octet-stream for binary data
	req.Header.Set("Content-Type", "application/octet-stream")

	// Send the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &SDKError{
			Message: fmt.Sprintf("request failed: %v", err),
			Code:    ErrorCodeUploadFailed,
		}
	}
	defer resp.Body.Close()

	// Check for success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &SDKError{
			Message: fmt.Sprintf("upload failed with status code %d", resp.StatusCode),
			Code:    ErrorCodeUploadFailed,
		}
	}

	return nil
}

// handleResponse processes HTTP responses and handles errors
func handleResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("failed to read response body: %v", err),
			Code:    ErrorCodeUnknownError,
		}
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return body, nil
	}

	// Handle error responses
	var errorCode ErrorCode
	var errorMessage string

	// Try to extract error message from response body
	var errorResp Response

	err = json.Unmarshal(body, &errorResp)

	switch resp.StatusCode {
	case http.StatusBadRequest:
		if errorResp != (Response{}) {
			if errorResp.Code == "free-tier-not-allowed" || errorResp.Code == "upload-limit-reached" {
				errorCode = ErrorCodeUnauthorized
				errorMessage = errorResp.Response
			} else {
				errorCode = ErrorCodeInvalidRequest
				errorMessage = fmt.Sprintf("Invalid request: %s", errorResp.Response)
			}
		} else {
			errorCode = ErrorCodeInvalidRequest
			errorMessage = "Invalid request"
		}
	case http.StatusUnauthorized:
		errorCode = ErrorCodeUnauthorized
		errorMessage = "Invalid API key"
	case http.StatusNotFound:
		errorCode = ErrorCodeNotFound
		errorMessage = "Resource not found"
	case http.StatusInternalServerError:
		errorCode = ErrorCodeServerError
		errorMessage = "Server error"
	default:
		if errorResp != (Response{}) {
			errorCode = ErrorCodeServerError
			errorMessage = fmt.Sprintf("API error: %s", errorResp.Response)
		} else {
			errorCode = ErrorCodeServerError
			errorMessage = "API error: Unknown error"
		}
	}

	return nil, &SDKError{
		Message: errorMessage,
		Code:    errorCode,
	}
}
