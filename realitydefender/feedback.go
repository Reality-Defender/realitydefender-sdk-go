package realitydefender

import (
	"context"
	"encoding/json"
	"fmt"
)

const userFeedbackV2Endpoint = "/api/v2/user-feedback"

type userFeedbackV2Payload struct {
	RequestID        string  `json:"requestId"`
	Label            string  `json:"label"`
	FeedbackCategory string  `json:"feedbackCategory"`
	Comment          *string `json:"comment,omitempty"`
}

func createUserFeedbackV2(ctx context.Context, client *httpClient, opts CreateUserFeedbackV2Options) (*UserFeedbackV2, error) {
	if opts.RequestID == "" || opts.Label == "" || opts.FeedbackCategory == "" {
		return nil, &SDKError{
			Message: "requestId, label, and feedbackCategory are required",
			Code:    ErrorCodeInvalidRequest,
		}
	}

	payload := userFeedbackV2Payload{
		RequestID:        opts.RequestID,
		Label:            opts.Label,
		FeedbackCategory: opts.FeedbackCategory,
		Comment:          opts.Comment,
	}

	responseData, err := client.post(ctx, userFeedbackV2Endpoint, payload)
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("user feedback submission failed: %v", err),
			Code:    ErrorCodeUploadFailed,
		}
	}

	var out UserFeedbackV2
	if err := json.Unmarshal(responseData, &out); err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("invalid response from user feedback API: %v", err),
			Code:    ErrorCodeServerError,
		}
	}

	return &out, nil
}
