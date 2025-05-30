package main

import (
	"encoding/json"
	"fmt"

	"github.com/Reality-Defender/realitydefender-sdk-go/src/realitydefender"
)

// Sample JSON response matching the API structure
const sampleResponse = `{
  "name": "test-file.jpg",
  "filename": "test-file.jpg",
  "originalFileName": "test-file.jpg",
  "requestId": "request-123",
  "uploadedDate": "2023-06-25T12:34:56Z",
  "mediaType": "IMAGE",
  "overallStatus": "ARTIFICIAL",
  "resultsSummary": {
    "status": "ARTIFICIAL",
    "metadata": {
      "finalScore": 0.95
    }
  },
  "models": [
    {
      "name": "model-1",
      "data": { "score": 0.95, "decision": "ARTIFICIAL", "raw_score": 0.95 },
      "status": "ARTIFICIAL",
      "finalScore": 0.95
    },
    {
      "name": "model-2",
      "data": null,
      "code": "not_applicable",
      "status": "NOT_APPLICABLE",
      "finalScore": null
    },
    {
      "name": "model-3",
      "data": { "score": 0.2, "decision": "HUMAN", "raw_score": 0.2 },
      "status": "AUTHENTIC",
      "finalScore": 0.2
    }
  ]
}`

func formatScore(score *float64) string {
	if score == nil {
		return "None"
	}
	return fmt.Sprintf("%.4f (%.1f%%)", *score, *score*100)
}

func main() {
	fmt.Println("Testing response parsing with sample data")

	// Create the MediaResponse struct
	var mediaResponse realitydefender.MediaResponse

	// Parse the sample JSON
	err := json.Unmarshal([]byte(sampleResponse), &mediaResponse)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return
	}

	// Format the response
	result := realitydefender.FormatResult(&mediaResponse)

	// Display the results
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

	fmt.Println("\nTest complete!")
}
