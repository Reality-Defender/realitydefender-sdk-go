package realitydefender_test

import (
	"context"
	"encoding/json"
	realitydefender2 "github.com/Reality-Defender/realitydefender-sdk-go/realitydefender"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Detection Functions", func() {
	var (
		server  *httptest.Server
		client  *realitydefender2.Client
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "detection-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
		os.RemoveAll(tempDir)
	})

	Describe("FormatResult", func() {
		It("correctly formats a result with scores above 1 (0-100 scale)", func() {
			// Create a test MediaResponse with scores on a 0-100 scale
			responseJSON := `{
				"name": "test-media",
				"filename": "test.jpg",
				"originalFileName": "test.jpg",
				"requestId": "test-request-id",
				"uploadedDate": "2023-01-01T00:00:00Z",
				"mediaType": "image",
				"overallStatus": "analyzed",
				"resultsSummary": {
					"status": "FAKE",
					"metadata": {
						"finalScore": 87.5
					}
				},
				"models": [
					{
						"name": "model1",
						"status": "FAKE",
						"finalScore": 90.2,
						"data": null
					},
					{
						"name": "model2",
						"status": "AUTHENTIC",
						"finalScore": 35.8,
						"data": null
					}
				]
			}`

			var mediaResponse realitydefender2.MediaResponse
			err := json.Unmarshal([]byte(responseJSON), &mediaResponse)
			Expect(err).NotTo(HaveOccurred())

			// Format the result
			result := realitydefender2.FormatResult(&mediaResponse)

			// Verify the result
			Expect(result.Status).To(Equal("MANIPULATED"))
			Expect(*result.Score).To(BeNumerically("~", 0.875, 0.001))
			Expect(len(result.Models)).To(Equal(2))

			// Check first model
			Expect(result.Models[0].Name).To(Equal("model1"))
			Expect(result.Models[0].Status).To(Equal("MANIPULATED"))
			Expect(*result.Models[0].Score).To(BeNumerically("~", 0.902, 0.001))

			// Check second model
			Expect(result.Models[1].Name).To(Equal("model2"))
			Expect(result.Models[1].Status).To(Equal("AUTHENTIC"))
			Expect(*result.Models[1].Score).To(BeNumerically("~", 0.358, 0.001))
		})

		It("correctly formats a result with scores already on 0-1 scale", func() {
			// Create a test MediaResponse with scores on a 0-1 scale
			responseJSON := `{
				"name": "test-media",
				"filename": "test.jpg",
				"originalFileName": "test.jpg",
				"requestId": "test-request-id",
				"uploadedDate": "2023-01-01T00:00:00Z",
				"mediaType": "image",
				"overallStatus": "analyzed",
				"resultsSummary": {
					"status": "AUTHENTIC",
					"metadata": {
						"finalScore": 0.25
					}
				},
				"models": [
					{
						"name": "model1",
						"status": "AUTHENTIC",
						"finalScore": 0.15,
						"data": null
					},
					{
						"name": "model2",
						"status": "FAKE",
						"finalScore": 0.92,
						"data": null
					}
				]
			}`

			var mediaResponse realitydefender2.MediaResponse
			err := json.Unmarshal([]byte(responseJSON), &mediaResponse)
			Expect(err).NotTo(HaveOccurred())

			// Format the result
			result := realitydefender2.FormatResult(&mediaResponse)

			// Verify the result
			Expect(result.Status).To(Equal("AUTHENTIC"))
			Expect(*result.Score).To(BeNumerically("~", 0.25, 0.001))
			Expect(len(result.Models)).To(Equal(2))

			// Check first model
			Expect(result.Models[0].Name).To(Equal("model1"))
			Expect(result.Models[0].Status).To(Equal("AUTHENTIC"))
			Expect(*result.Models[0].Score).To(BeNumerically("~", 0.15, 0.001))

			// Check second model
			Expect(result.Models[1].Name).To(Equal("model2"))
			Expect(result.Models[1].Status).To(Equal("MANIPULATED"))
			Expect(*result.Models[1].Score).To(BeNumerically("~", 0.92, 0.001))
		})

		It("handles null scores properly", func() {
			// Create a test MediaResponse with null scores
			responseJSON := `{
				"name": "test-media",
				"filename": "test.jpg",
				"originalFileName": "test.jpg",
				"requestId": "test-request-id",
				"uploadedDate": "2023-01-01T00:00:00Z",
				"mediaType": "image",
				"overallStatus": "analyzed",
				"resultsSummary": {
					"status": "ANALYZING",
					"metadata": {
						"finalScore": null
					}
				},
				"models": [
					{
						"name": "model1",
						"status": "ANALYZING",
						"finalScore": null,
						"data": null
					},
					{
						"name": "model2",
						"status": "NOT_APPLICABLE",
						"finalScore": null,
						"data": null
					}
				]
			}`

			var mediaResponse realitydefender2.MediaResponse
			err := json.Unmarshal([]byte(responseJSON), &mediaResponse)
			Expect(err).NotTo(HaveOccurred())

			// Format the result
			result := realitydefender2.FormatResult(&mediaResponse)

			// Verify the result
			Expect(result.Status).To(Equal("ANALYZING"))
			Expect(result.Score).To(BeNil())
			Expect(len(result.Models)).To(Equal(2))

			// Check first model
			Expect(result.Models[0].Name).To(Equal("model1"))
			Expect(result.Models[0].Status).To(Equal("ANALYZING"))
			Expect(result.Models[0].Score).To(BeNil())

			// Check second model
			Expect(result.Models[1].Name).To(Equal("model2"))
			Expect(result.Models[1].Status).To(Equal("NOT_APPLICABLE"))
			Expect(result.Models[1].Score).To(BeNil())
		})

		It("handles empty models array", func() {
			// Create a test MediaResponse with no models
			responseJSON := `{
				"name": "test-media",
				"filename": "test.jpg",
				"originalFileName": "test.jpg",
				"requestId": "test-request-id",
				"uploadedDate": "2023-01-01T00:00:00Z",
				"mediaType": "image",
				"overallStatus": "analyzed",
				"resultsSummary": {
					"status": "ERROR",
					"metadata": {
						"finalScore": null
					}
				},
				"models": []
			}`

			var mediaResponse realitydefender2.MediaResponse
			err := json.Unmarshal([]byte(responseJSON), &mediaResponse)
			Expect(err).NotTo(HaveOccurred())

			// Format the result
			result := realitydefender2.FormatResult(&mediaResponse)

			// Verify the result
			Expect(result.Status).To(Equal("ERROR"))
			Expect(result.Score).To(BeNil())
			Expect(len(result.Models)).To(Equal(0))
		})
	})

	Describe("Detection Result Polling", func() {
		It("handles result polling with exponential backoff", func() {
			// Mock server that returns processing status twice, then completed
			requestCount := 0

			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCount++

				if requestCount <= 2 {
					// First two calls return processing
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"overallStatus": "analyzing",
						"resultsSummary": {
							"status": "ANALYZING"
						}
					}`))
				} else {
					// Third call returns completed result
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"overallStatus": "complete",
						"resultsSummary": {
							"status": "FAKE",
							"metadata": {
								"finalScore": 0.95
							}
						},
						"models": [
							{
								"name": "test-model",
								"status": "FAKE",
								"finalScore": 0.95
							}
						]
					}`))
				}
			}))

			var err error
			client, err = realitydefender2.New(realitydefender2.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())

			// Use a reasonable context timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Set short polling interval to speed up test
			options := &realitydefender2.GetResultOptions{
				MaxAttempts:     5,
				PollingInterval: 50, // 50ms between attempts
			}

			result, err := client.GetResult(ctx, "test-request-id", options)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Status).To(Equal("MANIPULATED"))
			Expect(*result.Score).To(Equal(0.95))

			// Verify we made exactly 3 requests (2 processing + 1 completed)
			Expect(requestCount).To(Equal(3))
		})
	})
})
