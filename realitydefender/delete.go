package realitydefender

import (
	"context"
	"strings"
)

const deleteMediaEndpoint = "/api/media"

// DeleteAnalysis deletes a scan analysis (DELETE /api/media/{requestId}).
func (c *Client) DeleteAnalysis(ctx context.Context, requestID string) error {
	return deleteAnalysis(ctx, c.httpClient, requestID)
}

func deleteAnalysis(ctx context.Context, client *httpClient, requestID string) error {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return &SDKError{
			Message: "requestId is required",
			Code:    ErrorCodeInvalidRequest,
		}
	}

	return client.delete(ctx, deleteMediaEndpoint+"/"+requestID)
}
