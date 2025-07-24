package realitydefender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	signedURLEndpoint       = "/api/files/aws-presigned"
	mediaResultEndpoint     = "/api/media/users"
	allMediaResultsEndpoint = "/api/v2/media/users/pages"
	defaultMaxAttempts      = 30
)

// FileTypeConfig represents configuration for a supported file type
type FileTypeConfig struct {
	// Extensions is the list of supported file extensions
	Extensions []string `json:"extensions"`
	// SizeLimit is the maximum file size in bytes
	SizeLimit int64 `json:"size_limit"`
}

// SupportedFileTypes defines the supported file types and their size limits
var SupportedFileTypes = []FileTypeConfig{
	{
		Extensions: []string{".mp4", ".mov"},
		SizeLimit:  262144000, // ~250 MB
	},
	{
		Extensions: []string{".jpg", ".png", ".jpeg", ".gif", ".webp"},
		SizeLimit:  52428800, // ~50 MB
	},
	{
		Extensions: []string{".flac", ".wav", ".mp3", ".m4a", ".aac", ".alac", ".ogg"},
		SizeLimit:  20971520, // ~20 MB
	},
	{
		Extensions: []string{".txt"},
		SizeLimit:  5242880, // ~5 MB
	},
}

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

// AllMediaResponse represents a paginated response containing a list of media and related metadata.
type AllMediaResponse struct {
	TotalItems            int             `json:"totalItems"`
	TotalPages            int             `json:"totalPages"`
	CurrentPage           int             `json:"currentPage"`
	CurrentPageItemsCount int             `json:"currentPageItemsCount"`
	MediaList             []MediaResponse `json:"mediaList"`
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

	// Get file size
	fileInfo, _ := os.Stat(options.FilePath)
	fileSize := fileInfo.Size()

	// Get file extension
	fileExtension := strings.ToLower(filepath.Ext(options.FilePath))

	// Find size limit for the file extension
	var fileSizeLimit int64 = 0
	for _, fileType := range SupportedFileTypes {
		for _, ext := range fileType.Extensions {
			if ext == fileExtension {
				fileSizeLimit = fileType.SizeLimit
				break
			}
		}
		if fileSizeLimit != 0 {
			break
		}
	}

	// Check if file type is supported
	if fileSizeLimit == 0 {
		return nil, &SDKError{
			Message: fmt.Sprintf("Unsupported file type: %s", fileExtension),
			Code:    ErrorCodeInvalidFile,
		}
	}

	// Check file size
	if fileSize > fileSizeLimit {
		return nil, &SDKError{
			Message: fmt.Sprintf("File too large to upload: %s", options.FilePath),
			Code:    ErrorCodeFileTooLarge,
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
	requestID := response.RequestID

	status := response.ResultsSummary.Status

	// Replace FAKE with MANIPULATED
	if status == "FAKE" {
		status = "MANIPULATED"
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

		// Replace FAKE with MANIPULATED in model status
		modelStatus := model.Status
		if modelStatus == "FAKE" {
			modelStatus = "MANIPULATED"
		}

		models = append(models, ModelResult{
			Name:   model.Name,
			Status: modelStatus,
			Score:  modelScore,
		})
	}

	return &DetectionResult{
		RequestID: requestID,
		Status:    status,
		Score:     score,
		Models:    models,
	}
}

// formatResults converts an API response into a paginated list of formatted detection results.
func formatResults(response *AllMediaResponse) *DetectionResultList {
	var detectionResults []DetectionResult
	for _, media := range response.MediaList {
		detectionResults = append(detectionResults, *FormatResult(&media))
	}

	return &DetectionResultList{
		TotalItems:            response.TotalItems,
		TotalPages:            response.TotalPages,
		CurrentPage:           response.CurrentPage,
		CurrentPageItemsCount: response.CurrentPageItemsCount,
		Items:                 detectionResults,
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
		responseData, err := client.get(ctx, fmt.Sprintf("%s/%s", mediaResultEndpoint, requestID), nil)

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

// getDetectionResults gets the detection result stored in the platform
func getDetectionResults(ctx context.Context, client *httpClient, pageNumber *int, size *int, name *string, startDate *time.Time, endDate *time.Time, options GetResultOptions) (*DetectionResultList, error) {
	// Set default values if not provided
	if pageNumber == nil {
		defaultPageNumber := 0
		pageNumber = &defaultPageNumber
	}

	var parameters = make(map[string]string)

	if size == nil {
		defaultSize := 10
		size = &defaultSize
	}
	parameters["size"] = fmt.Sprintf("%d", *size)

	if name != nil {
		parameters["name"] = *name
	}

	if startDate != nil {
		parameters["startDate"] = startDate.Format("2006-01-02")
	}

	if endDate != nil {
		parameters["endDate"] = endDate.Format("2006-01-02")
	}

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
		responseData, err := client.get(ctx, fmt.Sprintf("%s/%d", allMediaResultsEndpoint, *pageNumber), parameters)

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
		var allMediaResponse AllMediaResponse
		if err := json.Unmarshal(responseData, &allMediaResponse); err != nil {
			return nil, &SDKError{
				Message: fmt.Sprintf("failed to parse result response: %v", err),
				Code:    ErrorCodeUnknownError,
			}
		}

		// Format the response into a DetectionResult
		result := formatResults(&allMediaResponse)

		return result, nil
	}

	// If we've reached this point, we've exceeded the maximum number of attempts
	return nil, &SDKError{
		Message: "exceeded maximum number of polling attempts",
		Code:    ErrorCodeTimeout,
	}
}
