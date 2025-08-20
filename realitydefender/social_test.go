package realitydefender_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	realitydefender "github.com/Reality-Defender/realitydefender-sdk-go/realitydefender"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Social Media Upload", func() {
	var (
		server *httptest.Server
		client *realitydefender.Client
	)

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("UploadSocialMedia", func() {
		Context("with valid social media URLs", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))
					Expect(r.URL.Path).To(Equal("/api/files/social"))
					Expect(r.Header.Get("X-API-Key")).To(Equal("test-api-key"))
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

					// Verify request payload
					var payload map[string]interface{}
					err := json.NewDecoder(r.Body).Decode(&payload)
					Expect(err).NotTo(HaveOccurred())
					Expect(payload).To(HaveKey("socialLink"))

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"requestId":"social-request-id"}`))
				}))

				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: server.URL,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("uploads a valid HTTPS URL successfully", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://twitter.com/example/status/123456789",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.RequestID).To(Equal("social-request-id"))
			})

			It("uploads a valid HTTP URL successfully", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "http://facebook.com/page/123",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.RequestID).To(Equal("social-request-id"))
			})

			It("uploads a URL with query parameters successfully", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://instagram.com/p/ABC123/?utm_source=ig_web_copy_link",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
			})
		})

		Context("with invalid URLs", func() {
			BeforeEach(func() {
				// Server should not be called for validation errors
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Fail("Server should not be called for validation errors")
				}))

				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: server.URL,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error when social link is empty", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Social media link is required"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
			})

			It("returns an error when URL has no scheme", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "twitter.com/example",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid social media link: twitter.com/example"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
			})

			It("returns an error when URL has invalid scheme", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "ftp://example.com/file",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid social media link: ftp://example.com/file"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
			})

			It("returns an error when URL has custom scheme", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "custom://example.com",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid social media link: custom://example.com"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
			})

			It("returns an error when URL has no host", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid social media link: https://"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
			})

			It("returns an error when URL has scheme but empty host", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https:///path/only",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid social media link: https:///path/only"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
			})

			It("returns an error for completely invalid URL format", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "not-a-url-at-all",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid social media link: not-a-url-at-all"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
			})
		})

		Context("with malformed URLs that cause parsing errors", func() {
			BeforeEach(func() {
				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: "http://not-used-for-validation-errors",
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error for URL with invalid IPv6 format", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "http://[::1:invalid",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid social media link:"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
			})
		})

		Context("when server returns errors", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "Internal server error"}`))
				}))

				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: server.URL,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("handles server errors correctly", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://twitter.com/example",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Social media link upload failed:"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeUploadFailed))
			})
		})

		Context("when server returns invalid JSON", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`invalid json response`))
				}))

				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: server.URL,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("handles JSON parsing errors correctly", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://twitter.com/example",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid response from API:"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeServerError))
			})
		})

		Context("when network connection fails", func() {
			BeforeEach(func() {
				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: "http://localhost:1", // Invalid URL that will cause connection errors
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("handles network errors correctly", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://twitter.com/example",
				})

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("Social media link upload failed:"))

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeUploadFailed))
			})
		})

		Context("when server response is missing required fields", func() {
			It("handles missing requestId in response", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{}`)) // Empty response
				}))

				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: server.URL,
				})
				Expect(err).NotTo(HaveOccurred())

				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://twitter.com/example",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid response from API"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeServerError))
			})

			It("handles null requestId in response", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"requestId":null}`))
				}))

				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: server.URL,
				})
				Expect(err).NotTo(HaveOccurred())

				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://twitter.com/example",
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Invalid response from API"))
				Expect(result).To(BeNil())

				var sdkErr *realitydefender.SDKError
				Expect(err).To(BeAssignableToTypeOf(sdkErr))
				sdkErr = err.(*realitydefender.SDKError)
				Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeServerError))
			})
		})

		Context("with context handling", func() {
			BeforeEach(func() {
				// Server that checks for context cancellation
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					select {
					case <-r.Context().Done():
						return
					default:
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"requestId":"test-id"}`))
					}
				}))

				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: server.URL,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("handles context cancellation", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://twitter.com/example",
				})

				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})

		Context("with payload validation", func() {
			BeforeEach(func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Verify request headers
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))
					Expect(r.Header.Get("X-API-Key")).To(Equal("test-api-key"))

					// Read and verify payload
					var payload map[string]interface{}
					err := json.NewDecoder(r.Body).Decode(&payload)
					Expect(err).NotTo(HaveOccurred())

					socialLink, exists := payload["socialLink"]
					Expect(exists).To(BeTrue())
					Expect(socialLink).To(Equal("https://twitter.com/test"))

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"requestId":"payload-test-id"}`))
				}))

				var err error
				client, err = realitydefender.New(realitydefender.Config{
					APIKey:  "test-api-key",
					BaseURL: server.URL,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("sends correct payload to the server", func() {
				ctx := context.Background()
				result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
					SocialLink: "https://twitter.com/test",
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.RequestID).To(Equal("payload-test-id"))
			})
		})

		Context("with different HTTP status codes", func() {
			DescribeTable("handling various HTTP error responses",
				func(statusCode int, responseBody string, expectedErrorCode realitydefender.ErrorCode, expectedErrorSubstring string) {
					server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(statusCode)
						w.Write([]byte(responseBody))
					}))

					var err error
					client, err = realitydefender.New(realitydefender.Config{
						APIKey:  "test-api-key",
						BaseURL: server.URL,
					})
					Expect(err).NotTo(HaveOccurred())

					ctx := context.Background()
					result, err := client.UploadSocialMedia(ctx, realitydefender.UploadSocialMediaOptions{
						SocialLink: "https://twitter.com/example",
					})

					Expect(err).To(HaveOccurred())
					Expect(result).To(BeNil())
					Expect(err.Error()).To(ContainSubstring(expectedErrorSubstring))

					var sdkErr *realitydefender.SDKError
					Expect(err).To(BeAssignableToTypeOf(sdkErr))
					sdkErr = err.(*realitydefender.SDKError)
					Expect(sdkErr.Code).To(Equal(expectedErrorCode))
				},
				Entry("400 Bad Request", http.StatusBadRequest, `{"code":"bad-request","response":"Invalid URL"}`, realitydefender.ErrorCodeUploadFailed, "Social media link upload failed:"),
				Entry("401 Unauthorized", http.StatusUnauthorized, `{"error":"Unauthorized"}`, realitydefender.ErrorCodeUploadFailed, "Social media link upload failed:"),
				Entry("403 Forbidden", http.StatusForbidden, `{"error":"Forbidden"}`, realitydefender.ErrorCodeUploadFailed, "Social media link upload failed:"),
				Entry("404 Not Found", http.StatusNotFound, `{"error":"Not found"}`, realitydefender.ErrorCodeUploadFailed, "Social media link upload failed:"),
				Entry("429 Too Many Requests", http.StatusTooManyRequests, `{"error":"Rate limited"}`, realitydefender.ErrorCodeUploadFailed, "Social media link upload failed:"),
				Entry("500 Internal Server Error", http.StatusInternalServerError, `{"error":"Server error"}`, realitydefender.ErrorCodeUploadFailed, "Social media link upload failed:"),
			)
		})
	})
})
