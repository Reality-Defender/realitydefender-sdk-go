package realitydefender

// UploadOptions represents options for uploading media
type UploadOptions struct {
	// FilePath is the path to the file to be analyzed
	FilePath string
}

// UploadSocialMediaOptions represents options for uploading social media
type UploadSocialMediaOptions struct {
	// SocialMediaLink is the URL of the social media to be analyzed
	SocialLink string
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
	// RequestID is the request ID that initiated the detection process
	RequestID string `json:"requestId"`
	// Status is the overall status determination (e.g., "MANIPULATED", "AUTHENTIC")
	Status string `json:"status"`
	// Score is the confidence score (0-1, nil if processing)
	Score *float64 `json:"score"`
	// Models contains results from individual detection models
	Models []ModelResult `json:"models"`
}

// DetectionResultList represents a paginated list of detection results.
type DetectionResultList struct {
	// TotalItems is the total number of items available.
	TotalItems int `json:"total_items"`
	// CurrentPageItemsCount is the number of items on the current page.
	CurrentPageItemsCount int `json:"current_page_items_count"`
	// TotalPages is the total number of pages available.
	TotalPages int `json:"total_pages"`
	// CurrentPage is the current page number in the result set.
	CurrentPage int `json:"current_page"`
	// Items is a slice containing the detection results for the current page.
	Items []DetectionResult `json:"items"`
}

// Response represents a standard structure for API responses, including status codes, messages, and error details.
type Response struct {
	Code      string  `json:"code"`
	Response  string  `json:"response"`
	ErrNo     int     `json:"errno"`
	RequestID *string `json:"requestId"`
}

// CreateUserFeedbackOptions configures POST /api/v2/user-feedback.
type CreateUserFeedbackOptions struct {
	// RequestID is the media / detection result ID (required).
	RequestID string
	// Label is the content judgment: REAL, SYNTHETIC, MANIPULATED, UNKNOWN (required).
	Label string
	// FeedbackCategory is one of FALSE_POSITIVE, FALSE_NEGATIVE, CONFIRMATION, OTHER (required).
	FeedbackCategory string
	// Comment is optional free text.
	Comment *string
}

// UserFeedback is the JSON body returned on successful feedback creation (201).
type UserFeedback struct {
	ID            string `json:"id,omitempty"`
	UserID        string `json:"userId,omitempty"`
	RequestID     string `json:"requestId,omitempty"`
	InstitutionID string `json:"institutionId,omitempty"`
	Text          string `json:"text,omitempty"`
	Category      string `json:"category,omitempty"`
	UserName      string `json:"userName,omitempty"`
	UserEmail     string `json:"userEmail,omitempty"`
	OrgName       string `json:"orgName,omitempty"`
	MediaType     string `json:"mediaType,omitempty"`
	MediaViewURL  string `json:"mediaViewUrl,omitempty"`
	MediaSource   string `json:"mediaSource,omitempty"`
	Label         string `json:"label,omitempty"`
	CreatedAt     string `json:"createdAt,omitempty"`
}
