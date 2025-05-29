package realitydefender

// UploadOptions represents options for uploading media
type UploadOptions struct {
	// FilePath is the path to the file to be analyzed
	FilePath string
}

// UploadResult represents the result of a successful upload
type UploadResult struct {
	// RequestID is the ID used to retrieve results
	RequestID string `json:"request_id"`
	// MediaID is the ID assigned by the system
	MediaID string `json:"media_id"`
}

// GetResultOptions represents options for retrieving results
type GetResultOptions struct {
	// MaxAttempts is the maximum number of polling attempts before returning even if still analyzing
	MaxAttempts int
	// PollingInterval is the interval in milliseconds between polling attempts
	PollingInterval int
}

// PollOptions represents options for polling for results
type PollOptions struct {
	// PollingInterval is the interval in milliseconds between polling attempts
	PollingInterval int
	// Timeout is the maximum time to poll in milliseconds
	Timeout int
}

// ModelResult represents results from an individual detection model
type ModelResult struct {
	// Name is the model name
	Name string `json:"name"`
	// Status is the model status determination
	Status string `json:"status"`
	// Score is the model confidence score (0-1, nil if not available)
	Score *float64 `json:"score"`
}

// DetectionResult represents the simplified detection result returned to the user
type DetectionResult struct {
	// Status is the overall status determination (e.g., "ARTIFICIAL", "AUTHENTIC")
	Status string `json:"status"`
	// Score is the confidence score (0-1, nil if processing)
	Score *float64 `json:"score"`
	// Models contains results from individual detection models
	Models []ModelResult `json:"models"`
}
