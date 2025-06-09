package realitydefender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	signedURLEndpoint   = "/api/files/aws-presigned"
	mediaResultEndpoint = "/api/media/users"
	defaultMaxAttempts  = 30
)

// SignedURLResponse represents the response from the signed URL request
type SignedURLResponse struct {
	Code     string `json:"code"`
	Response struct {
		SignedURL string `json:"signedUrl"`
	} `json:"response"`
	Errno     int    `json:"errno"`
	MediaID   string `json:"mediaId"`
	RequestID string `json:"requestId"`
}

// MediaResponse represents the raw API response for media results
type MediaResponse struct {
	Name             string `json:"name"`
	Filename         string `json:"filename"`
	OriginalFileName string `json:"originalFileName"`
	RequestID        string `json:"requestId"`
	UploadedDate     string `json:"uploadedDate"`
	MediaType        string `json:"mediaType"`
	OverallStatus    string `json:"overallStatus"`
	ResultsSummary   struct {
		Status   string `json:"status"`
		Metadata struct {
			FinalScore *float64 `json:"finalScore"`
		} `json:"metadata"`
	} `json:"resultsSummary"`
	Models []struct {
		Name       string      `json:"name"`
		Status     string      `json:"status"`
		FinalScore *float64    `json:"finalScore"`
		Data       interface{} `json:"data"`
		Code       string      `json:"code,omitempty"`
	} `json:"models"`
}

// getSignedURL gets a signed URL for uploading a file
func getSignedURL(ctx context.Context, client *httpClient, fileName string) (*SignedURLResponse, error) {
	// Create request payload
	payload := map[string]string{
		"fileName": fileName,
	}

	// Get signed URL
	responseData, err := client.post(ctx, signedURLEndpoint, payload)

	if err != nil {

		return nil, err



	}

	// Parse the response
	var response SignedURLResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("failed to parse signed URL response: %v", err),
			Code:    ErrorCodeUnknownError,
		}
	}

	return &response, nil
}

// uploadToSignedURL uploads file content to a signed URL
func uploadToSignedURL(ctx context.Context, client *httpClient, signedURL string, filePath string) error {
	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return &SDKError{
			Message: fmt.Sprintf("failed to read file: %v", err),
			Code:    ErrorCodeInvalidFile,
		}
	}

	// Upload to signed URL
	err = client.put(ctx, signedURL, fileContent)
	if err != nil {
		return &SDKError{
			Message: fmt.Sprintf("failed to upload to signed URL: %v", err),
			Code:    ErrorCodeUploadFailed,
		}
	}

	return nil
}

// uploadFile uploads a file to Reality Defender for analysis
func uploadFile(ctx context.Context, client *httpClient, options UploadOptions) (*UploadResult, error) {
	// Validate file path
	if options.FilePath == "" {
		return nil, &SDKError{
			Message: "file path is required",
			Code:    ErrorCodeInvalidFile,
		}
	}

	// Check if file exists
	if _, err := os.Stat(options.FilePath); os.IsNotExist(err) {
		return nil, &SDKError{
			Message: fmt.Sprintf("file not found: %s", options.FilePath),
			Code:    ErrorCodeInvalidFile,
		}
	}

	// Get the filename
	fileName := filepath.Base(options.FilePath)

	// Get signed URL
	signedURLResponse, err := getSignedURL(ctx, client, fileName)
	if err != nil {
		return nil, err
	}

	// Upload to signed URL
	err = uploadToSignedURL(ctx, client, signedURLResponse.Response.SignedURL, options.FilePath)
	if err != nil {
		return nil, err
	}

	// Return result for tracking
	return &UploadResult{
		RequestID: signedURLResponse.RequestID,
		MediaID:   signedURLResponse.MediaID,
	}, nil
}

// FormatResult formats the raw API response into a user-friendly result
func FormatResult(response *MediaResponse) *DetectionResult {
	// Extract the overall status and score
	status := response.ResultsSummary.Status

	// Replace FAKE with ARTIFICIAL
	if status == "FAKE" {
		status = "ARTIFICIAL"
	}

	// Normalize score from 0-100 to 0-1 if needed
	var score *float64
	if response.ResultsSummary.Metadata.FinalScore != nil {
		normalizedScore := *response.ResultsSummary.Metadata.FinalScore
		// If the score is greater than 1, assume it's on a 0-100 scale and normalize
		if normalizedScore > 1 {
			normalizedScore = normalizedScore / 100.0
		}
		score = &normalizedScore
	}

	// Format models
	var models []ModelResult
	for _, model := range response.Models {
		// Normalize model score from 0-100 to 0-1 if needed
		var modelScore *float64
		if model.FinalScore != nil {
			normalizedModelScore := *model.FinalScore
			// If the score is greater than 1, assume it's on a 0-100 scale and normalize
			if normalizedModelScore > 1 {
				normalizedModelScore = normalizedModelScore / 100.0
			}
			modelScore = &normalizedModelScore
		}

		// Replace FAKE with ARTIFICIAL in model status
		modelStatus := model.Status
		if modelStatus == "FAKE" {
			modelStatus = "ARTIFICIAL"
		}

		models = append(models, ModelResult{
			Name:   model.Name,
			Status: modelStatus,
			Score:  modelScore,
		})
	}

	return &DetectionResult{
		Status: status,
		Score:  score,
		Models: models,
	}
}

// getDetectionResult gets the detection result for a specific request ID
func getDetectionResult(ctx context.Context, client *httpClient, requestID string, options GetResultOptions) (*DetectionResult, error) {
	// Set default values if not provided
	maxAttempts := options.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = defaultMaxAttempts
	}

	pollingInterval := options.PollingInterval
	if pollingInterval <= 0 {
		pollingInterval = DefaultPollingInterval
	}

	// Keep track of attempts
	attempt := 0

	// Loop until we get a result or reach max attempts
	for attempt < maxAttempts {
		// Get the result
		responseData, err := client.get(ctx, fmt.Sprintf("%s/%s", mediaResultEndpoint, requestID))

		// Handle specific error types
		if err != nil {
			var sdkErr *SDKError
			if errors.As(err, &sdkErr) && sdkErr.Code == ErrorCodeNotFound {
				// If resource not found, wait and try again
				attempt++

				// Check if we've reached max attempts
				if attempt >= maxAttempts {
					return nil, err
				}

				// Wait for the polling interval before trying again
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(pollingInterval) * time.Millisecond):
					continue
				}
			}

			// Any other error, return immediately
			return nil, err
		}

		// Parse the response
		var mediaResponse MediaResponse
		if err := json.Unmarshal(responseData, &mediaResponse); err != nil {
			return nil, &SDKError{
				Message: fmt.Sprintf("failed to parse result response: %v", err),
				Code:    ErrorCodeUnknownError,
			}
		}

		// Format the response into a DetectionResult
		result := FormatResult(&mediaResponse)

		// Check if we have a final result or still analyzing
		isAnalyzing := result.Status == "ANALYZING"

		// Also check if all models are still analyzing or not applicable
		allModelsAnalyzingOrNA := true
		hasAnalyzingModels := false

		for _, model := range result.Models {
			if model.Status != "ANALYZING" && model.Status != "NOT_APPLICABLE" {
				allModelsAnalyzingOrNA = false
				break
			}
			if model.Status == "ANALYZING" {
				hasAnalyzingModels = true
			}
		}

		// Continue polling if overall status is ANALYZING or
		// if all models are either ANALYZING or NOT_APPLICABLE and at least one is ANALYZING
		if isAnalyzing || (allModelsAnalyzingOrNA && hasAnalyzingModels) {
			attempt++

			// Check if we've reached max attempts
			if attempt >= maxAttempts {
				return result, nil
			}

			// Wait for the polling interval before trying again
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(pollingInterval) * time.Millisecond):
				continue
			}
		}

		// We have a final result
		return result, nil
	}

	// If we've reached this point, we've exceeded the maximum number of attempts
	return nil, &SDKError{
		Message: "exceeded maximum number of polling attempts",
		Code:    ErrorCodeTimeout,
	}
}
