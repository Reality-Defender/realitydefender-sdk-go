package main

import (
	"context"
	"fmt"
	"os"
	"strings"
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
	fmt.Println("Reality Defender SDK Social Media Analysis Example")
	fmt.Println("==================================================")

	// Get API key from environment variable
	apiKey := os.Getenv("REALITY_DEFENDER_API_KEY")
	if apiKey == "" {
		fmt.Println("ERROR: Please set REALITY_DEFENDER_API_KEY environment variable")
		fmt.Println("Example: export REALITY_DEFENDER_API_KEY=your-api-key-here")
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

	// Get social media URL from command line argument or use a default
	var socialURL string
	if len(os.Args) > 1 {
		socialURL = os.Args[1]
	} else {
		// Example URLs for demonstration (you can replace these with real URLs)
		fmt.Println("No URL provided. Please provide a social media URL as a command line argument.")
		fmt.Println("Usage: go run main.go <social-media-url>")
		fmt.Println("\nExample URLs you could use:")
		fmt.Println("  go run main.go https://twitter.com/username/status/123456789")
		fmt.Println("  go run main.go https://facebook.com/username/posts/123456789")
		fmt.Println("  go run main.go https://instagram.com/p/ABC123/")
		fmt.Println("  go run main.go https://youtube.com/watch?v=dQw4w9WgXcQ")
		os.Exit(1)
	}

	fmt.Printf("Analyzing social media content from: %s\n\n", socialURL)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Upload social media link for analysis
	fmt.Println("Submitting social media link for analysis...")
	uploadResult, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
		SocialLink: socialURL,
	})
	if err != nil {
		fmt.Printf("Upload failed: %v\n", err)

		// Provide helpful error messages for common issues
		if strings.Contains(err.Error(), "invalid_request") {
			fmt.Println("\nPossible reasons:")
			fmt.Println("- The URL format is invalid")
			fmt.Println("- The social media platform may not be supported")
			fmt.Println("- The URL may not be publicly accessible")
		} else if strings.Contains(err.Error(), "unauthorized") {
			fmt.Println("\nPlease check your API key is valid and has the necessary permissions.")
		}
		os.Exit(1)
	}

	fmt.Println("✓ Upload successful!")
	fmt.Printf("Request ID: %s\n", uploadResult.RequestID)

	// Get the result with polling
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Polling for analysis results...")
	fmt.Println("This may take some time as our AI models analyze the content...")

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
				progress := "Analyzing content" + strings.Repeat(".", dots) + strings.Repeat(" ", 3-dots)
				fmt.Print("\r" + progress)
			}
		}
	}()

	// Poll for results with custom options
	result, err := client.GetResult(ctx, uploadResult.RequestID, &realitydefender.GetResultOptions{
		PollingInterval: 3000, // 3 seconds between polls
		MaxAttempts:     60,   // Maximum 60 attempts (3 minutes)
	})

	// Stop the progress indicator
	ticker.Stop()
	done <- true
	fmt.Print("\r" + strings.Repeat(" ", 30) + "\r")

	if err != nil {
		fmt.Printf("Failed to get result: %v\n", err)

		// Provide helpful information for timeout errors
		if strings.Contains(err.Error(), "timeout") {
			fmt.Println("\nThe analysis is taking longer than expected.")
			fmt.Println("You can try checking the result later using the Request ID:", uploadResult.RequestID)
		}
		os.Exit(1)
	}

	// Display results
	fmt.Println("✓ Analysis complete!")
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("DETECTION RESULTS")
	fmt.Println(strings.Repeat("=", 50))

	// Overall result
	fmt.Printf("Overall Status: %s\n", result.Status)
	fmt.Printf("Confidence Score: %s\n", formatScore(result.Score))

	// Interpret the results
	fmt.Println("\nInterpretation:")
	switch result.Status {
	case "MANIPULATED":
		fmt.Println("⚠️  This content appears to be artificially generated or manipulated.")
	case "AUTHENTIC":
		fmt.Println("✅ This content appears to be authentic (not artificially generated).")
	case "ANALYZING":
		fmt.Println("⏳ Analysis is still in progress. Some models may still be processing.")
	default:
		fmt.Printf("🔍 Status: %s\n", result.Status)
	}

	// Model-specific results
	if len(result.Models) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 50))
		fmt.Println("INDIVIDUAL MODEL RESULTS")
		fmt.Println(strings.Repeat("-", 50))

		for i, model := range result.Models {
			fmt.Printf("%d. Model: %s\n", i+1, model.Name)
			fmt.Printf("   Status: %s\n", model.Status)
			fmt.Printf("   Score: %s\n", formatScore(model.Score))

			// Add interpretation for each model
			switch model.Status {
			case "MANIPULATED":
				fmt.Println("   💡 This model detected signs of artificial generation/manipulation")
			case "AUTHENTIC":
				fmt.Println("   💡 This model found no signs of manipulation")
			case "NOT_APPLICABLE":
				fmt.Println("   💡 This model is not applicable to this type of content")
			case "ANALYZING":
				fmt.Println("   💡 This model is still processing the content")
			}
			fmt.Println()
		}
	}

	// Provide next steps
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("NEXT STEPS")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Request ID: %s\n", result.RequestID)
	fmt.Println("You can use this Request ID to retrieve results later or for reference.")
	fmt.Println("\nTo analyze another URL, run:")
	fmt.Println("  go run main.go <another-social-media-url>")

	fmt.Println("\n✨ Analysis complete! Thank you for using Reality Defender.")
}
