package realitydefender

import (
	"context"
	"encoding/json"
	"fmt"
)

const userFeedbackEndpoint = "/api/v2/user-feedback"

type userFeedbackPayload struct {
	RequestID        string  `json:"requestId"`
	Label            string  `json:"label"`
	FeedbackCategory string  `json:"feedbackCategory"`
	Comment          *string `json:"comment,omitempty"`
}

func createUserFeedback(ctx context.Context, client *httpClient, opts CreateUserFeedbackOptions) (*UserFeedback, error) {
	if opts.RequestID == "" || opts.Label == "" || opts.FeedbackCategory == "" {
		return nil, &SDKError{
			Message: "requestId, label, and feedbackCategory are required",
			Code:    ErrorCodeInvalidRequest,
		}
	}

	payload := userFeedbackPayload{
		RequestID:        opts.RequestID,
		Label:            opts.Label,
		FeedbackCategory: opts.FeedbackCategory,
		Comment:          opts.Comment,
	}

	responseData, err := client.post(ctx, userFeedbackEndpoint, payload)
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("user feedback submission failed: %v", err),
			Code:    ErrorCodeUploadFailed,
		}
	}

	var out UserFeedback
	if err := json.Unmarshal(responseData, &out); err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("invalid response from user feedback API: %v", err),
			Code:    ErrorCodeServerError,
		}
	}

	return &out, nil
}
