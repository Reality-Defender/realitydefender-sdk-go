package realitydefender

import (
	"context"
	"strings"
)

const deleteMediaEndpoint = "/api/media"

// DeleteRequest deletes a scan request (DELETE /api/media/{requestId}).
func (c *Client) DeleteRequest(ctx context.Context, requestID string) error {
	return deleteRequest(ctx, c.httpClient, requestID)
}

func deleteRequest(ctx context.Context, client *httpClient, requestID string) error {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return &SDKError{
			Message: "requestId is required",
			Code:    ErrorCodeInvalidRequest,
		}
	}

	return client.delete(ctx, deleteMediaEndpoint+"/"+requestID)
}
