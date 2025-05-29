package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Reality-Defender/eng-sdk/go/src/realitydefender"
)

func formatScore(score *float64) string {
	if score == nil {
		return "None"
	}
	return fmt.Sprintf("%.4f (%.1f%%)", *score, *score*100)
}

func main() {
	fmt.Println("Reality Defender SDK Event-Based Example")

	// Get API key from environment variable
	apiKey := os.Getenv("REALITY_DEFENDER_API_KEY")
	if apiKey == "" {
		fmt.Println("ERROR: Please set REALITY_DEFENDER_API_KEY environment variable")
		os.Exit(1)
	}

	// Initialize the SDK
	client, err := realitydefender.New(realitydefender.Config{
		APIKey: apiKey,
	})
	if err != nil {
		fmt.Printf("Failed to initialize client: %v\n", err)
		os.Exit(1)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Create a WaitGroup to wait for the event to be received
	var wg sync.WaitGroup
	wg.Add(1)

	// Set up event handlers
	client.On("result", func(data interface{}) {
		result := data.(*realitydefender.DetectionResult)

		fmt.Println("\nResult received from event:")
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

		wg.Done()
	})

	client.On("error", func(data interface{}) {
		err := data.(error)
		fmt.Printf("\nError received from event: %v\n", err)
		wg.Done()
	})

	// Upload a file for analysis
	filePath := "../images/test_image.jpg"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("File not found: %s\n", filePath)
		fmt.Println("Please add a test image named 'test_image.jpg' to the examples/images directory")
		os.Exit(1)
	}

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

	// Start polling using the event-based approach
	fmt.Println("\nStarting event-based polling...")
	err = client.PollForResults(ctx, uploadResult.RequestID, &realitydefender.PollOptions{
		PollingInterval: 2000,  // 2 seconds between polls
		Timeout:         60000, // 60 second timeout
	})
	if err != nil {
		fmt.Printf("Polling failed: %v\n", err)
		os.Exit(1)
	}

	// Wait for the result event
	wg.Wait()
	fmt.Println("\nPolling complete!")
	fmt.Println("\nExample complete!")
}
