# Reality Defender Go SDK

[![codecov](https://codecov.io/gh/Reality-Defender/realitydefender-sdk-go/graph/badge.svg?token=TQZSGX3R7Z)](https://codecov.io/gh/Reality-Defender/realitydefender-sdk-go)

Go client library for the Reality Defender API for deepfake detection and media manipulation analysis.

## Installation

```bash
go get github.com/Reality-Defender/realitydefender-sdk-go
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