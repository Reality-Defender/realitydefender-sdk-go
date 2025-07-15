package realitydefender_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRealityDefender(t *testing.T) {
	// Set a test timeout to prevent hanging tests
	t.Parallel()
	timeout := 30 * time.Second
	done := make(chan struct{})

	go func() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "RealityDefender Suite")
		close(done)
	}()

	select {
	case <-done:
		// Tests completed normally
	case <-time.After(timeout):
		t.Fatal("Test execution timed out after", timeout)
	}
}
