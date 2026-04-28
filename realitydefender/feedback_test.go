package realitydefender_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	realitydefender "github.com/Reality-Defender/realitydefender-sdk-go/realitydefender"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CreateUserFeedbackV2", func() {
	var (
		server *httptest.Server
		client *realitydefender.Client
	)

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Context("with valid options", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal("POST"))
				Expect(r.URL.Path).To(Equal("/api/v2/user-feedback"))
				Expect(r.Header.Get("X-API-KEY")).To(Equal("test-api-key"))

				body, err := io.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())
				var payload map[string]interface{}
				Expect(json.Unmarshal(body, &payload)).To(Succeed())
				Expect(payload["requestId"]).To(Equal("req-1"))
				Expect(payload["label"]).To(Equal("REAL"))
				Expect(payload["feedbackCategory"]).To(Equal("CONFIRMATION"))

				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"fb-1","requestId":"req-1","category":"CONFIRMATION"}`))
			}))

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns parsed user feedback on 201", func() {
			ctx := context.Background()
			out, err := client.CreateUserFeedbackV2(ctx, realitydefender.CreateUserFeedbackV2Options{
				RequestID:        "req-1",
				Label:            "REAL",
				FeedbackCategory: "CONFIRMATION",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(BeNil())
			Expect(out.ID).To(Equal("fb-1"))
			Expect(out.RequestID).To(Equal("req-1"))
		})

		It("includes comment in JSON when set", func() {
			server.Close()
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				Expect(err).NotTo(HaveOccurred())
				var payload map[string]interface{}
				Expect(json.Unmarshal(body, &payload)).To(Succeed())
				Expect(payload["comment"]).To(Equal("note"))

				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"fb-2"}`))
			}))

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())

			c := "note"
			_, err = client.CreateUserFeedbackV2(context.Background(), realitydefender.CreateUserFeedbackV2Options{
				RequestID:        "req-2",
				Label:            "REAL",
				FeedbackCategory: "OTHER",
				Comment:          &c,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("with missing required fields", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Fail("server should not be called when validation fails")
			}))
			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns invalid_request when requestId is empty", func() {
			_, err := client.CreateUserFeedbackV2(context.Background(), realitydefender.CreateUserFeedbackV2Options{
				Label:            "REAL",
				FeedbackCategory: "CONFIRMATION",
			})
			Expect(err).To(HaveOccurred())
			var sdkErr *realitydefender.SDKError
			Expect(err).To(BeAssignableToTypeOf(sdkErr))
			sdkErr = err.(*realitydefender.SDKError)
			Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
		})
	})
})
