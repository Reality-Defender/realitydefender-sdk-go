package realitydefender_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	realitydefender "github.com/Reality-Defender/realitydefender-sdk-go/realitydefender"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DeleteRequest", func() {
	var (
		server *httptest.Server
		client *realitydefender.Client
	)

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Context("with a valid request id", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal(http.MethodDelete))
				Expect(r.URL.Path).To(Equal("/api/media/req-1"))
				Expect(r.Header.Get("X-API-KEY")).To(Equal("test-api-key"))
				w.WriteHeader(http.StatusNoContent)
			}))

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("succeeds on 204 No Content", func() {
			err := client.DeleteRequest(context.Background(), "req-1")
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when the media is missing", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"code":"not-found","response":"Media not found"}`))
			}))

			var err error
			client, err = realitydefender.New(realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: server.URL,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns not_found", func() {
			err := client.DeleteRequest(context.Background(), "missing")
			Expect(err).To(HaveOccurred())
			sdkErr := err.(*realitydefender.SDKError)
			Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeNotFound))
		})
	})

	Context("with a blank request id", func() {
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

		It("returns invalid_request", func() {
			err := client.DeleteRequest(context.Background(), "  ")
			Expect(err).To(HaveOccurred())
			sdkErr := err.(*realitydefender.SDKError)
			Expect(sdkErr.Code).To(Equal(realitydefender.ErrorCodeInvalidRequest))
		})
	})
})
