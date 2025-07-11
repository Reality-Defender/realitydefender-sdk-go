package realitydefender_test

import (
	"encoding/json"

	"github.com/Reality-Defender/realitydefender-sdk-go/src/realitydefender"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Types and Structures", func() {
	Describe("SDKError", func() {
		It("formats error messages correctly", func() {
			err := &realitydefender.SDKError{
				Message: "Test error message",
				Code:    realitydefender.ErrorCodeInvalidFile,
			}

			errString := err.Error()
			Expect(errString).To(ContainSubstring("Test error message"))
			Expect(errString).To(ContainSubstring("invalid_file"))
			Expect(errString).To(ContainSubstring("Code:"))
		})
	})

	Describe("DetectionResult", func() {
		It("serializes and deserializes correctly", func() {
			// Create a score pointer
			score := 0.95
			modelScore1 := 0.99
			modelScore2 := 0.01

			// Create a test result
			original := &realitydefender.DetectionResult{
				Status: "MANIPULATED",
				Score:  &score,
				Models: []realitydefender.ModelResult{
					{
						Name:   "model1",
						Status: "MANIPULATED",
						Score:  &modelScore1,
					},
					{
						Name:   "model2",
						Status: "REAL",
						Score:  &modelScore2,
					},
				},
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResult
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify all fields match
			Expect(result.Status).To(Equal(original.Status))
			Expect(*result.Score).To(Equal(*original.Score))
			Expect(result.Models).To(HaveLen(2))

			Expect(result.Models[0].Name).To(Equal(original.Models[0].Name))
			Expect(result.Models[0].Status).To(Equal(original.Models[0].Status))
			Expect(*result.Models[0].Score).To(Equal(*original.Models[0].Score))

			Expect(result.Models[1].Name).To(Equal(original.Models[1].Name))
			Expect(result.Models[1].Status).To(Equal(original.Models[1].Status))
			Expect(*result.Models[1].Score).To(Equal(*original.Models[1].Score))
		})

		It("handles null scores correctly", func() {
			// Create a result with nil score
			original := &realitydefender.DetectionResult{
				Status: "ANALYZING",
				Score:  nil,
				Models: []realitydefender.ModelResult{
					{
						Name:   "model1",
						Status: "ANALYZING",
						Score:  nil,
					},
				},
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResult
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify all fields match
			Expect(result.Status).To(Equal(original.Status))
			Expect(result.Score).To(BeNil())
			Expect(result.Models).To(HaveLen(1))
			Expect(result.Models[0].Name).To(Equal(original.Models[0].Name))
			Expect(result.Models[0].Status).To(Equal(original.Models[0].Status))
			Expect(result.Models[0].Score).To(BeNil())
		})
	})

	Describe("Config", func() {
		It("validates configuration settings", func() {
			// Valid config
			validConfig := realitydefender.Config{
				APIKey:  "test-api-key",
				BaseURL: "https://custom.api.example.com",
			}

			client, err := realitydefender.New(validConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())

			// Missing API key
			invalidConfig := realitydefender.Config{
				BaseURL: "https://custom.api.example.com",
			}

			client, err = realitydefender.New(invalidConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("API key is required"))
			Expect(client).To(BeNil())
		})
	})
})
