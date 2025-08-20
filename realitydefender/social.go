package realitydefender

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

const (
	socialMediaEndpoint = "/api/files/social"
)

// uploadSocialMediaLink upload social media link to Reality Defender
func uploadSocialMediaLink(ctx context.Context, client *httpClient, options UploadSocialMediaOptions) (*UploadResult, error) {

	// Validate file path
	if options.SocialLink == "" {
		return nil, &SDKError{
			Message: "Social media link is required",
			Code:    ErrorCodeInvalidRequest,
		}
	}

	// Validate that SocialLink is a valid URL
	parsedURL, err := url.Parse(options.SocialLink)
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("Invalid social media link: %v", err),
			Code:    ErrorCodeInvalidRequest,
		}
	}

	// Check if the URL has a valid scheme (http or https)
	if (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return nil, &SDKError{
			Message: fmt.Sprintf("Invalid social media link: %s", options.SocialLink),
			Code:    ErrorCodeInvalidRequest,
		}
	}

	payload := map[string]string{
		"socialLink": options.SocialLink,
	}

	responseData, err := client.post(ctx, socialMediaEndpoint, payload)
	if err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("Social media link upload failed:: %v", err),
			Code:    ErrorCodeUploadFailed,
		}
	}

	// Parse the response
	var response Response
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, &SDKError{
			Message: fmt.Sprintf("Invalid response from API: %v", err),
			Code:    ErrorCodeServerError,
		}
	}
	if response.RequestID == nil {
		return nil, &SDKError{
			Message: "Invalid response from API",
			Code:    ErrorCodeServerError,
		}
	}

	return &UploadResult{
		RequestID: *response.RequestID,
	}, nil
}
