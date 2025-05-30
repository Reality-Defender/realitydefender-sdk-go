package realitydefender_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Reality-Defender/realitydefender-sdk-go/src/realitydefender"
)

var _ = Describe("RealityDefender SDK", func() {
	var (
		server  *httptest.Server
		client  *realitydefender.Client
		tempDir string
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "realitydefender-test")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
		os.RemoveAll(tempDir)
	})

	Describe("New", func() {
		It("returns an error when API key is missing", func() {
			client, err := realitydefender.New(realitydefender.Config{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("API key is required"))
			Expect(client).To(BeNil())
		})

		It("creates a client with default base URL when not provided", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Header.Get("X-API-Key")).To(Equal("test-api-key"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"request_id":"test-id"}`))
			}))

			client, err := realitydefender.New(realitydefender.Config{
				APIKey: "test-api-key",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})

		It("creates a client with custom base URL when provided", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Header.Get("X-API-Key")).To(Equal("test-api-key"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"request_id":"test-id"}`))
			}))

			client, err := realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})
	})

	Describe("Upload", func() {
		BeforeEach(func() {
			mux := http.NewServeMux()
			server = httptest.NewServer(mux)

			// Handle presigned URL request
			mux.HandleFunc("/api/files/aws-presigned", func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("POST"))
				Expect(r.Header.Get("X-API-Key")).To(Equal("test-api-key"))

				// Create a valid upload URL with the server's URL as base
				uploadURL := server.URL + "/upload-endpoint"

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"code":"success","response":{"signedUrl":"` + uploadURL + `"},"errno":0,"mediaId":"test-media-id","requestId":"test-request-id"}`))
			})

			// Handle the PUT request to the signed URL
			mux.HandleFunc("/upload-endpoint", func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("PUT"))
				w.WriteHeader(http.StatusOK)
			})

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error when file path is empty", func() {
			ctx := context.Background()
			result, err := client.Upload(ctx, realitydefender.UploadOptions{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("file path is required"))
			Expect(result).To(BeNil())
		})

		It("returns an error when file does not exist", func() {
			ctx := context.Background()
			result, err := client.Upload(ctx, realitydefender.UploadOptions{
				FilePath: "/path/to/non/existent/file.jpg",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("file not found"))
			Expect(result).To(BeNil())
		})

		It("uploads a file successfully", func() {
			// Create a temporary file
			filePath := filepath.Join(tempDir, "test.jpg")
			err := os.WriteFile(filePath, []byte("test-content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			ctx := context.Background()
			result, err := client.Upload(ctx, realitydefender.UploadOptions{
				FilePath: filePath,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.RequestID).To(Equal("test-request-id"))
			Expect(result.MediaID).To(Equal("test-media-id"))
		})
	})

	Describe("GetResult", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))
				Expect(r.URL.Path).To(Equal("/api/media/users/test-request-id"))
				Expect(r.Header.Get("X-API-Key")).To(Equal("test-api-key"))

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"name": "test.jpg",
					"filename": "test.jpg",
					"originalFileName": "test.jpg",
					"requestId": "test-request-id",
					"uploadedDate": "2025-05-28T13:00:00Z",
					"mediaType": "image",
					"overallStatus": "complete",
					"resultsSummary": {
						"status": "ARTIFICIAL",
						"metadata": {
							"finalScore": 0.95
						}
					},
					"models": [
						{
							"name": "test-model",
							"status": "ARTIFICIAL",
							"finalScore": 0.95
						}
					]
				}`))
			}))

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("gets a result successfully", func() {
			ctx := context.Background()
			result, err := client.GetResult(ctx, "test-request-id", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Status).To(Equal("ARTIFICIAL"))
			Expect(*result.Score).To(Equal(0.95))
			Expect(result.Models).To(HaveLen(1))
			Expect(result.Models[0].Name).To(Equal("test-model"))
			Expect(result.Models[0].Status).To(Equal("ARTIFICIAL"))
			Expect(*result.Models[0].Score).To(Equal(0.95))
		})
	})

	Describe("DetectFile", func() {
		BeforeEach(func() {
			mux := http.NewServeMux()
			server = httptest.NewServer(mux)

			// Handle presigned URL request
			mux.HandleFunc("/api/files/aws-presigned", func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("POST"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"code":"success","response":{"signedUrl":"` + server.URL + `/upload-endpoint"},"errno":0,"mediaId":"test-media-id","requestId":"test-request-id"}`))
			})

			// Handle the PUT request to the signed URL
			mux.HandleFunc("/upload-endpoint", func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("PUT"))
				w.WriteHeader(http.StatusOK)
			})

			// Handle result
			mux.HandleFunc("/api/media/users/test-request-id", func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"resultsSummary": {
						"status": "ARTIFICIAL",
						"metadata": {
							"finalScore": 0.95
						}
					},
					"models": [
						{
							"name": "test-model",
							"status": "ARTIFICIAL",
							"finalScore": 0.95
						}
					]
				}`))
			})

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error when file doesn't exist", func() {
			ctx := context.Background()
			result, err := client.DetectFile(ctx, "/nonexistent/file.jpg")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("file not found"))
			Expect(result).To(BeNil())
		})

		It("detects a file successfully", func() {
			// Create a temporary file
			filePath := filepath.Join(tempDir, "test-detect.jpg")
			err := os.WriteFile(filePath, []byte("test-content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			ctx := context.Background()
			result, err := client.DetectFile(ctx, filePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Status).To(Equal("ARTIFICIAL"))
			Expect(*result.Score).To(Equal(0.95))
		})
	})

	Describe("PUT Request handling", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("PUT"))
				Expect(r.Header.Get("Content-Type")).To(Equal("application/octet-stream"))

				// Read the request body to verify it
				body, err := io.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(body)).To(Equal("test-file-content"))

				w.WriteHeader(http.StatusOK)
			}))

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: "http://not-used-in-this-test",
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("handles PUT request errors correctly", func() {
			// Close the server to force connection errors
			server.Close()

			ctx := context.Background()

			// Create a temporary file
			filePath := filepath.Join(tempDir, "test-put.jpg")
			err := os.WriteFile(filePath, []byte("test-file-content"), 0644)
			Expect(err).NotTo(HaveOccurred())

			// Create a new upload options and try to uploadToSignedURL (via the Upload method)
			// This will fail because the server is closed
			_, err = client.Upload(ctx, realitydefender.UploadOptions{
				FilePath: filePath,
			})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("request failed"))
		})
	})

	Describe("Error Handling", func() {
		BeforeEach(func() {
			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: "http://localhost:1", // Invalid URL that will cause connection errors
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("handles HTTP errors correctly in GetResult", func() {
			ctx := context.Background()
			result, err := client.GetResult(ctx, "test-request-id", nil)

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("request failed"))
		})

		It("handles file validation errors in Upload", func() {
			ctx := context.Background()
			result, err := client.Upload(ctx, realitydefender.UploadOptions{
				FilePath: "", // Empty path should cause validation error
			})

			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("file path is required"))
		})
	})

	Describe("PollForResults", func() {
		var analyzingCount int

		BeforeEach(func() {
			analyzingCount = 2
			mux := http.NewServeMux()
			server = httptest.NewServer(mux)

			// Handle result with initial ANALYZING status, then ARTIFICIAL status
			mux.HandleFunc("/api/media/users/test-request-id", func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))

				if analyzingCount > 0 {
					analyzingCount--
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"name": "test.jpg",
						"filename": "test.jpg",
						"originalFileName": "test.jpg",
						"requestId": "test-request-id",
						"uploadedDate": "2025-05-28T13:00:00Z",
						"mediaType": "image",
						"overallStatus": "analyzing",
						"resultsSummary": {
							"status": "ANALYZING",
							"metadata": {
								"finalScore": null
							}
						},
						"models": []
					}`))
				} else {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"name": "test.jpg",
						"filename": "test.jpg",
						"originalFileName": "test.jpg",
						"requestId": "test-request-id",
						"uploadedDate": "2025-05-28T13:00:00Z",
						"mediaType": "image",
						"overallStatus": "complete",
						"resultsSummary": {
							"status": "ARTIFICIAL",
							"metadata": {
								"finalScore": 0.95
							}
						},
						"models": [
							{
								"name": "test-model",
								"status": "ARTIFICIAL",
								"finalScore": 0.95
							}
						]
					}`))
				}
			})

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("polls for results until they are ready", func() {
			// Use a longer context timeout
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			resultCh := make(chan *realitydefender.DetectionResult, 1)

			client.On("result", func(data interface{}) {
				result := data.(*realitydefender.DetectionResult)
				resultCh <- result
			})

			// Set very short polling interval and reasonable timeout
			err := client.PollForResults(ctx, "test-request-id", &realitydefender.PollOptions{
				PollingInterval: 50,   // 50ms for faster test
				Timeout:         5000, // 5s timeout should be enough
			})
			Expect(err).NotTo(HaveOccurred())

			var result *realitydefender.DetectionResult
			select {
			case result = <-resultCh:
				// Result received
			case <-time.After(5 * time.Second):
				Fail("Timed out waiting for result")
			}

			Expect(result).NotTo(BeNil())
			Expect(result.Status).To(Equal("ARTIFICIAL"))
			Expect(*result.Score).To(Equal(0.95))
		})

		It("polls for results until complete", func() {
			ctx := context.Background()

			// Set up result handler
			var resultReceived bool
			var resultData *realitydefender.DetectionResult

			client.On("result", func(data interface{}) {
				resultReceived = true
				resultData = data.(*realitydefender.DetectionResult)
			})

			// Start polling with short interval for test
			err := client.PollForResults(ctx, "test-request-id", &realitydefender.PollOptions{
				PollingInterval: 100,  // 100ms
				Timeout:         5000, // 5s
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(resultReceived).To(BeTrue())
			Expect(resultData).NotTo(BeNil())
			Expect(resultData.Status).To(Equal("ARTIFICIAL"))
			Expect(*resultData.Score).To(Equal(0.95))
		})

		It("handles context cancellation", func() {
			// Set up server to respond with processing status
			mux := http.NewServeMux()
			server = httptest.NewServer(mux)

			mux.HandleFunc("/api/media/users/test-request-id", func(w http.ResponseWriter, r *http.Request) {
				// Sleep briefly to ensure we can cancel before response
				time.Sleep(50 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"overallStatus": "processing",
					"resultsSummary": {
						"status": "PROCESSING"
					}
				}`))
			})

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create a cancellable context
			ctx, cancel := context.WithCancel(context.Background())

			// Cancel after a short delay
			go func() {
				time.Sleep(10 * time.Millisecond)
				cancel()
			}()

			// This should be cancelled by our context
			err = client.PollForResults(ctx, "test-request-id", nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("context canceled"))
		})
	})

	Describe("Event handling", func() {
		It("registers and triggers event handlers", func() {
			client, err := realitydefender.New(realitydefender.Config{
				APIKey: "test-api-key",
			})
			Expect(err).NotTo(HaveOccurred())

			// Create a channel to track if the event was triggered
			resultCalled := make(chan interface{}, 1)
			errorCalled := make(chan interface{}, 1)

			// Register event handlers
			client.On("result", func(data interface{}) {
				resultCalled <- data
			})

			client.On("error", func(data interface{}) {
				errorCalled <- data
			})

			// Create test server for GetResult
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))
				Expect(r.URL.Path).To(Equal("/api/media/users/test-request-id"))

				// Return a completed result
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"name": "test.jpg",
					"requestId": "test-request-id",
					"mediaType": "image",
					"overallStatus": "complete",
					"resultsSummary": {
						"status": "ARTIFICIAL",
						"metadata": {
							"finalScore": 95
						}
					},
					"models": [
						{
							"name": "model1",
							"status": "ARTIFICIAL",
							"finalScore": 95
						}
					]
				}`))
			}))
			defer server.Close()

			// Update client to use test server
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())

			// Register the handlers again for the new client
			client.On("result", func(data interface{}) {
				resultCalled <- data
			})

			client.On("error", func(data interface{}) {
				errorCalled <- data
			})

			// Start polling with very short timeout
			ctx := context.Background()
			err = client.PollForResults(ctx, "test-request-id", &realitydefender.PollOptions{
				PollingInterval: 10,
				Timeout:         1000,
			})
			Expect(err).NotTo(HaveOccurred())

			// Check if the result event was triggered
			var result interface{}
			select {
			case result = <-resultCalled:
				// Event was triggered
			case <-time.After(100 * time.Millisecond):
				Fail("Result event was not triggered")
			}

			// Verify the result is a DetectionResult
			detectionResult, ok := result.(*realitydefender.DetectionResult)
			Expect(ok).To(BeTrue())
			Expect(detectionResult.Status).To(Equal("ARTIFICIAL"))
		})

		It("handles multiple registered handlers for the same event", func() {
			client, err := realitydefender.New(realitydefender.Config{
				APIKey: "test-api-key",
			})
			Expect(err).NotTo(HaveOccurred())

			// Create channels to track events
			handler1Called := make(chan bool, 1)
			handler2Called := make(chan bool, 1)

			// Register multiple handlers for the same event
			client.On("result", func(data interface{}) {
				handler1Called <- true
			})

			client.On("result", func(data interface{}) {
				handler2Called <- true
			})

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"name": "test.jpg",
					"requestId": "test-request-id",
					"mediaType": "image",
					"overallStatus": "complete",
					"resultsSummary": {
						"status": "ARTIFICIAL",
						"metadata": {
							"finalScore": 95
						}
					},
					"models": [
						{
							"name": "model1",
							"status": "ARTIFICIAL",
							"finalScore": 95
						}
					]
				}`))
			}))
			defer server.Close()

			// Update client to use test server
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())

			// Register the handlers again for the new client
			client.On("result", func(data interface{}) {
				handler1Called <- true
			})

			client.On("result", func(data interface{}) {
				handler2Called <- true
			})

			// Start polling
			ctx := context.Background()
			err = client.PollForResults(ctx, "test-request-id", &realitydefender.PollOptions{
				PollingInterval: 10,
				Timeout:         1000,
			})
			Expect(err).NotTo(HaveOccurred())

			// Check if both handlers were called
			select {
			case <-handler1Called:
				// First handler was called
			case <-time.After(100 * time.Millisecond):
				Fail("First handler was not called")
			}

			select {
			case <-handler2Called:
				// Second handler was called
			case <-time.After(100 * time.Millisecond):
				Fail("Second handler was not called")
			}
		})
	})

	Describe("PollForResults", func() {
		It("polls until result is ready", func() {
			// Counter to track number of requests
			requestCount := 0

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("GET"))
				Expect(r.URL.Path).To(Equal("/api/media/users/test-request-id"))

				requestCount++

				// First request returns analyzing status
				if requestCount == 1 {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"name": "test.jpg",
						"requestId": "test-request-id",
						"mediaType": "image",
						"overallStatus": "analyzing",
						"resultsSummary": {
							"status": "ANALYZING",
							"metadata": {}
						},
						"models": [
							{
								"name": "model1",
								"status": "ANALYZING"
							}
						]
					}`))
				} else {
					// Subsequent requests return completed status
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"name": "test.jpg",
						"requestId": "test-request-id",
						"mediaType": "image",
						"overallStatus": "complete",
						"resultsSummary": {
							"status": "ARTIFICIAL",
							"metadata": {
								"finalScore": 95
							}
						},
						"models": [
							{
								"name": "model1",
								"status": "ARTIFICIAL",
								"finalScore": 95
							}
						]
					}`))
				}
			}))
			defer server.Close()

			client, err := realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())

			// Track if result event was called
			resultCalled := make(chan interface{}, 1)
			client.On("result", func(data interface{}) {
				resultCalled <- data
			})

			// Start polling
			ctx := context.Background()
			err = client.PollForResults(ctx, "test-request-id", &realitydefender.PollOptions{
				PollingInterval: 10,
				Timeout:         1000,
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify that multiple requests were made
			Expect(requestCount).To(BeNumerically(">", 1))

			// Check if result event was triggered
			var result interface{}
			select {
			case result = <-resultCalled:
				// Event was triggered
			case <-time.After(100 * time.Millisecond):
				Fail("Result event was not triggered")
			}

			// Verify the result
			detectionResult, ok := result.(*realitydefender.DetectionResult)
			Expect(ok).To(BeTrue())
			Expect(detectionResult.Status).To(Equal("ARTIFICIAL"))
		})

		It("handles polling timeout", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Always return analyzing status
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"name": "test.jpg",
					"requestId": "test-request-id",
					"mediaType": "image",
					"overallStatus": "analyzing",
					"resultsSummary": {
						"status": "ANALYZING",
						"metadata": {}
					},
					"models": [
						{
							"name": "model1",
							"status": "ANALYZING"
						}
					]
				}`))
			}))
			defer server.Close()

			client, err := realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())

			// Track if error event was called
			errorCalled := make(chan interface{}, 1)
			client.On("error", func(data interface{}) {
				errorCalled <- data
			})

			// Use a timeout context to prevent test hanging
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Start polling with very short timeout
			err = client.PollForResults(ctx, "test-request-id", &realitydefender.PollOptions{
				PollingInterval: 10,
				Timeout:         50, // Very short timeout to force timeout error
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("context deadline exceeded"))

			// No need to check for error event as it won't be triggered with context cancellation
		})

		It("handles context cancellation", func() {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Always return analyzing status
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"name": "test.jpg",
					"requestId": "test-request-id",
					"mediaType": "image",
					"overallStatus": "analyzing",
					"resultsSummary": {
						"status": "ANALYZING",
						"metadata": {}
					},
					"models": [
						{
							"name": "model1",
							"status": "ANALYZING"
						}
					]
				}`))
			}))
			defer server.Close()

			client, err := realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())

			// Create a context that can be cancelled
			ctx, cancel := context.WithCancel(context.Background())

			// Cancel the context after a short delay
			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()

			// Start polling
			err = client.PollForResults(ctx, "test-request-id", &realitydefender.PollOptions{
				PollingInterval: 10,
				Timeout:         1000,
			})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(context.Canceled))
		})
	})

	Describe("DetectFile", func() {
		It("combines upload and get result in one step", func() {
			// This is a stub test as it's difficult to test file uploads without real files
			// We'd need to create temporary files or mock at a lower level
			// For now, we're just validating the client parameter validation
			client, err := realitydefender.New(realitydefender.Config{
				APIKey: "test-api-key",
			})
			Expect(err).NotTo(HaveOccurred())

			ctx := context.Background()
			result, err := client.DetectFile(ctx, "non-existent-file.jpg")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("file not found"))
			Expect(result).To(BeNil())
		})
	})
})
