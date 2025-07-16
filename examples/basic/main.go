package main

import (
	"context"
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

func main() {
	fmt.Println("Reality Defender SDK Basic Example")

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

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
	fmt.Printf("Media ID: %s\n", uploadResult.MediaID)

	// Get the result
	fmt.Println("\nPolling for results...")
	fmt.Println("This may take some time as models analyze the content...")

	// Create a ticker to show progress
	ticker := time.NewTicker(5 * time.Second)
	done := make(chan bool)

	// Start a goroutine to show progress
	go func() {
		dots := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				dots = (dots % 3) + 1
				progress := "Waiting for analysis to complete" + "...."[:dots]
				fmt.Print("\r" + progress + "                    ")
			}
		}
	}()

	// Use GetResult with a longer polling interval for better performance
	result, err := client.GetResult(ctx, uploadResult.RequestID, &realitydefender.GetResultOptions{
		PollingInterval: 3000, // 3 seconds between polls
	})

	// Stop the progress indicator
	ticker.Stop()
	done <- true
	fmt.Println("\r                                              ")

	if err != nil {
		fmt.Printf("Failed to get result: %v\n", err)
		os.Exit(1)
	}

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

	fmt.Println("\nExample complete!")
}
