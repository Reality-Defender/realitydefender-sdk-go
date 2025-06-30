# Reality Defender Go SDK

[![codecov](https://codecov.io/gh/Reality-Defender/realitydefender-sdk-go/graph/badge.svg?token=TQZSGX3R7Z)](https://codecov.io/gh/Reality-Defender/realitydefender-sdk-go)

Go client library for the Reality Defender API for deepfake detection and media manipulation analysis.

## Installation

```bash
go get github.com/Reality-Defender/realitydefender-sdk-go
```

### Basic Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Reality-Defender/realitydefender-sdk-go/src/realitydefender"
)

func main() {
	// Initialize the SDK with your API key
	client, err := realitydefender.New(realitydefender.Config{
		APIKey: os.Getenv("REALITY_DEFENDER_API_KEY"),
	})
	if err != nil {
		fmt.Printf("Failed to initialize client: %v\n", err)
		os.Exit(1)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Upload a file for analysis
	filePath := "./image.jpg"
	uploadResult, err := client.Upload(ctx, realitydefender.UploadOptions{
		FilePath: filePath,
	})
	if err != nil {
		fmt.Printf("Upload failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Upload successful! Request ID: %s\n", uploadResult.RequestID)

	// Get the result
	result, err := client.GetResult(ctx, uploadResult.RequestID, nil)
	if err != nil {
		fmt.Printf("Failed to get result: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("Status: %s\n", result.Status)
	if result.Score != nil {
		fmt.Printf("Score: %.4f\n", *result.Score)
	}
}
```

### Event-Based Example

```go
package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Reality-Defender/realitydefender-sdk-go/src/realitydefender"
)

func main() {
	// Initialize the SDK
	client, err := realitydefender.New(realitydefender.Config{
		APIKey: os.Getenv("REALITY_DEFENDER_API_KEY"),
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
		fmt.Printf("Result received: %s\n", result.Status)
		wg.Done()
	})

	client.On("error", func(data interface{}) {
		err := data.(error)
		fmt.Printf("Error received: %v\n", err)
		wg.Done()
	})

	// Upload a file for analysis
	filePath := "./image.jpg"
	uploadResult, err := client.Upload(ctx, realitydefender.UploadOptions{
		FilePath: filePath,
	})
	if err != nil {
		fmt.Printf("Upload failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Upload successful! Request ID: %s\n", uploadResult.RequestID)

	// Start polling using the event-based approach
	err = client.PollForResults(ctx, uploadResult.RequestID, &realitydefender.PollOptions{
		PollingInterval: 2000, // 2 seconds between polls
		Timeout:         60000, // 60 second timeout
	})
	if err != nil {
		fmt.Printf("Polling failed: %v\n", err)
		os.Exit(1)
	}

	// Wait for the result event
	wg.Wait()
}
```

## API Reference

### Client Initialization

```go
client, err := realitydefender.New(realitydefender.Config{
    APIKey: "your-api-key",
    BaseURL: "https://api.dev.realitydefender.xyz", // Optional, defaults to production
})
```

### Upload a File

```go
uploadResult, err := client.Upload(ctx, realitydefender.UploadOptions{
    FilePath: "./path/to/file.jpg",
})
```

### Get Result

```go
result, err := client.GetResult(ctx, uploadResult.RequestID, &realitydefender.GetResultOptions{
    MaxAttempts: 30,          // Optional, defaults to 30
    PollingInterval: 2000,    // Optional, defaults to 2000ms
})
```

### Poll for Results (Event-Based)

```go
// Set up event handlers
client.On("result", func(data interface{}) {
    result := data.(*realitydefender.DetectionResult)
    // Handle result
})

client.On("error", func(data interface{}) {
    err := data.(error)
    // Handle error
})

// Start polling
err = client.PollForResults(ctx, uploadResult.RequestID, &realitydefender.PollOptions{
    PollingInterval: 2000, // Optional, defaults to 2000ms
    Timeout:         60000, // Optional, defaults to 60000ms
})
```

### Convenience Method

```go
// Upload and get result in one step
result, err := client.DetectFile(ctx, "./path/to/file.jpg")
```

## Development

The included `Justfile` has all the shortcuts needed to build the module, run tests, examples, etc.  