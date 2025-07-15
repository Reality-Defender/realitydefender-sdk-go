package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Reality-Defender/realitydefender-sdk-go/realitydefender"
)

func formatScore(score *float64) string {
	if score == nil {
		return "None"
	}
	return fmt.Sprintf("%.4f (%.1f%%)", *score, *score*100)
}

// analyzeWithChannels demonstrates using channels to handle asynchronous detection results
func analyzeWithChannels(ctx context.Context, filePath string, apiKey string) {
	// Initialize the SDK
	client, err := realitydefender.New(realitydefender.Config{
		APIKey: apiKey,
	})
	if err != nil {
		fmt.Printf("Failed to initialize client: %v\n", err)
		os.Exit(1)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File not found: %s\n", filePath)
		os.Exit(1)
	}

	// Create channels for results and errors
	resultCh := make(chan *realitydefender.DetectionResult, 1)
	errCh := make(chan error, 1)

	// Upload the file
	fmt.Printf("Uploading file: %s\n", filePath)
	uploadResult, err := client.Upload(ctx, realitydefender.UploadOptions{
		FilePath: filePath,
	})
	if err != nil {
		fmt.Printf("Upload failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Upload successful!")
	fmt.Printf("Request ID: %s\n", uploadResult.RequestID)

	// Start a goroutine to poll for results
	go func() {
		// Poll for results with exponential backoff
		attempt := 0
		maxAttempts := 30
		baseWait := 1000 // milliseconds
		requestID := uploadResult.RequestID

		for attempt < maxAttempts {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				// Get the result
				result, err := client.GetResult(ctx, requestID, nil)
				if err != nil {
					// If not found, wait and try again
					var sdkErr *realitydefender.SDKError
					if errors.As(err, &sdkErr) && sdkErr.Code == realitydefender.ErrorCodeNotFound {
						waitTime := time.Duration(baseWait*(attempt+1)) * time.Millisecond
						time.Sleep(waitTime)
						attempt++
						continue
					}

					// Any other error, return immediately
					errCh <- err
					return
				}

				// If still analyzing, wait and try again
				if result.Status == "ANALYZING" {
					waitTime := time.Duration(baseWait*(attempt+1)) * time.Millisecond
					time.Sleep(waitTime)
					attempt++
					continue
				}

				// We have a final result
				resultCh <- result
				return
			}
		}

		// If we reached max attempts
		errCh <- fmt.Errorf("exceeded maximum number of polling attempts")
	}()

	// Wait for the result or error
	fmt.Println("\nWaiting for analysis results...")
	select {
	case result := <-resultCh:
		fmt.Println("\nDetection Results:")
		fmt.Printf("Status: %s\n", result.Status)
		fmt.Printf("Score: %s\n", formatScore(result.Score))

		fmt.Println("\nModel Results:")
		for _, model := range result.Models {
			fmt.Printf("  - %s: %s (Score: %s)\n",
				model.Name,
				model.Status,
				formatScore(model.Score),
			)
		}

	case err := <-errCh:
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)

	case <-ctx.Done():
		fmt.Printf("\nOperation canceled: %v\n", ctx.Err())
		os.Exit(1)
	}
}

func main() {
	fmt.Println("Reality Defender SDK Channel-Based Example")

	// Get API key from environment variable
	apiKey := os.Getenv("REALITY_DEFENDER_API_KEY")
	if apiKey == "" {
		fmt.Println("ERROR: Please set REALITY_DEFENDER_API_KEY environment variable")
		os.Exit(1)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run the channel-based example
	filePath := "../images/test_image.jpg"
	analyzeWithChannels(ctx, filePath, apiKey)

	fmt.Println("\nExample complete!")
}
