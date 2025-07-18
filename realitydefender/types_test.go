package realitydefender_test

import (
	"encoding/json"
	"github.com/Reality-Defender/realitydefender-sdk-go/realitydefender"
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

	Describe("DetectionResultList", func() {
		It("serializes and deserializes correctly", func() {
			// Create test scores
			score1 := 0.95
			score2 := 0.05
			modelScore1 := 0.99
			modelScore2 := 0.01

			// Create a test result list
			original := &realitydefender.DetectionResultList{
				TotalItems:            50,
				CurrentPageItemsCount: 2,
				TotalPages:            25,
				CurrentPage:           1,
				Items: []realitydefender.DetectionResult{
					{
						Status: "MANIPULATED",
						Score:  &score1,
						Models: []realitydefender.ModelResult{
							{
								Name:   "model1",
								Status: "MANIPULATED",
								Score:  &modelScore1,
							},
						},
					},
					{
						Status: "AUTHENTIC",
						Score:  &score2,
						Models: []realitydefender.ModelResult{
							{
								Name:   "model2",
								Status: "AUTHENTIC",
								Score:  &modelScore2,
							},
						},
					},
				},
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResultList
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify pagination fields
			Expect(result.TotalItems).To(Equal(original.TotalItems))
			Expect(result.CurrentPageItemsCount).To(Equal(original.CurrentPageItemsCount))
			Expect(result.TotalPages).To(Equal(original.TotalPages))
			Expect(result.CurrentPage).To(Equal(original.CurrentPage))

			// Verify items
			Expect(result.Items).To(HaveLen(2))

			// Verify first item
			Expect(result.Items[0].Status).To(Equal(original.Items[0].Status))
			Expect(*result.Items[0].Score).To(Equal(*original.Items[0].Score))
			Expect(result.Items[0].Models).To(HaveLen(1))
			Expect(result.Items[0].Models[0].Name).To(Equal(original.Items[0].Models[0].Name))
			Expect(result.Items[0].Models[0].Status).To(Equal(original.Items[0].Models[0].Status))
			Expect(*result.Items[0].Models[0].Score).To(Equal(*original.Items[0].Models[0].Score))

			// Verify second item
			Expect(result.Items[1].Status).To(Equal(original.Items[1].Status))
			Expect(*result.Items[1].Score).To(Equal(*original.Items[1].Score))
			Expect(result.Items[1].Models).To(HaveLen(1))
			Expect(result.Items[1].Models[0].Name).To(Equal(original.Items[1].Models[0].Name))
			Expect(result.Items[1].Models[0].Status).To(Equal(original.Items[1].Models[0].Status))
			Expect(*result.Items[1].Models[0].Score).To(Equal(*original.Items[1].Models[0].Score))
		})

		It("handles empty items list correctly", func() {
			// Create a result list with no items
			original := &realitydefender.DetectionResultList{
				TotalItems:            0,
				CurrentPageItemsCount: 0,
				TotalPages:            0,
				CurrentPage:           1,
				Items:                 []realitydefender.DetectionResult{},
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResultList
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify all fields match
			Expect(result.TotalItems).To(Equal(0))
			Expect(result.CurrentPageItemsCount).To(Equal(0))
			Expect(result.TotalPages).To(Equal(0))
			Expect(result.CurrentPage).To(Equal(1))
			Expect(result.Items).To(HaveLen(0))
			Expect(result.Items).NotTo(BeNil())
		})

		It("handles nil items list correctly", func() {
			// Create a result list with nil items
			original := &realitydefender.DetectionResultList{
				TotalItems:            0,
				CurrentPageItemsCount: 0,
				TotalPages:            0,
				CurrentPage:           1,
				Items:                 nil,
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResultList
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify all fields match
			Expect(result.TotalItems).To(Equal(0))
			Expect(result.CurrentPageItemsCount).To(Equal(0))
			Expect(result.TotalPages).To(Equal(0))
			Expect(result.CurrentPage).To(Equal(1))
			Expect(result.Items).To(BeNil())
		})

		It("handles large pagination values correctly", func() {
			// Create a result list with large values
			original := &realitydefender.DetectionResultList{
				TotalItems:            1000000,
				CurrentPageItemsCount: 100,
				TotalPages:            10000,
				CurrentPage:           5000,
				Items:                 []realitydefender.DetectionResult{},
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResultList
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify all fields match
			Expect(result.TotalItems).To(Equal(1000000))
			Expect(result.CurrentPageItemsCount).To(Equal(100))
			Expect(result.TotalPages).To(Equal(10000))
			Expect(result.CurrentPage).To(Equal(5000))
			Expect(result.Items).To(HaveLen(0))
		})

		It("handles mixed processing states in items", func() {
			// Create test scores
			score1 := 0.85

			// Create a result list with mixed processing states
			original := &realitydefender.DetectionResultList{
				TotalItems:            3,
				CurrentPageItemsCount: 3,
				TotalPages:            1,
				CurrentPage:           1,
				Items: []realitydefender.DetectionResult{
					{
						Status: "MANIPULATED",
						Score:  &score1,
						Models: []realitydefender.ModelResult{
							{
								Name:   "model1",
								Status: "MANIPULATED",
								Score:  &score1,
							},
						},
					},
					{
						Status: "ANALYZING",
						Score:  nil,
						Models: []realitydefender.ModelResult{
							{
								Name:   "model2",
								Status: "ANALYZING",
								Score:  nil,
							},
						},
					},
					{
						Status: "ERROR",
						Score:  nil,
						Models: []realitydefender.ModelResult{},
					},
				},
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResultList
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify pagination fields
			Expect(result.TotalItems).To(Equal(3))
			Expect(result.CurrentPageItemsCount).To(Equal(3))
			Expect(result.TotalPages).To(Equal(1))
			Expect(result.CurrentPage).To(Equal(1))

			// Verify items
			Expect(result.Items).To(HaveLen(3))

			// Verify first item (completed)
			Expect(result.Items[0].Status).To(Equal("MANIPULATED"))
			Expect(result.Items[0].Score).NotTo(BeNil())
			Expect(*result.Items[0].Score).To(Equal(0.85))
			Expect(result.Items[0].Models).To(HaveLen(1))

			// Verify second item (processing)
			Expect(result.Items[1].Status).To(Equal("ANALYZING"))
			Expect(result.Items[1].Score).To(BeNil())
			Expect(result.Items[1].Models).To(HaveLen(1))
			Expect(result.Items[1].Models[0].Score).To(BeNil())

			// Verify third item (error)
			Expect(result.Items[2].Status).To(Equal("ERROR"))
			Expect(result.Items[2].Score).To(BeNil())
			Expect(result.Items[2].Models).To(HaveLen(0))
		})

		It("validates pagination consistency", func() {
			// Test case where pagination values are consistent
			validList := &realitydefender.DetectionResultList{
				TotalItems:            100,
				CurrentPageItemsCount: 10,
				TotalPages:            10,
				CurrentPage:           5,
				Items:                 make([]realitydefender.DetectionResult, 10),
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(validList)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResultList
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify pagination fields are preserved
			Expect(result.TotalItems).To(Equal(100))
			Expect(result.CurrentPageItemsCount).To(Equal(10))
			Expect(result.TotalPages).To(Equal(10))
			Expect(result.CurrentPage).To(Equal(5))
			Expect(result.Items).To(HaveLen(10))
		})

		It("handles zero pagination values", func() {
			// Create a result list with zero values
			original := &realitydefender.DetectionResultList{
				TotalItems:            0,
				CurrentPageItemsCount: 0,
				TotalPages:            0,
				CurrentPage:           0,
				Items:                 []realitydefender.DetectionResult{},
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResultList
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify all fields match
			Expect(result.TotalItems).To(Equal(0))
			Expect(result.CurrentPageItemsCount).To(Equal(0))
			Expect(result.TotalPages).To(Equal(0))
			Expect(result.CurrentPage).To(Equal(0))
			Expect(result.Items).To(HaveLen(0))
		})

		It("handles single item pagination", func() {
			// Create test score
			score := 0.75

			// Create a result list with single item
			original := &realitydefender.DetectionResultList{
				TotalItems:            1,
				CurrentPageItemsCount: 1,
				TotalPages:            1,
				CurrentPage:           1,
				Items: []realitydefender.DetectionResult{
					{
						Status: "AUTHENTIC",
						Score:  &score,
						Models: []realitydefender.ModelResult{
							{
								Name:   "singleModel",
								Status: "AUTHENTIC",
								Score:  &score,
							},
						},
					},
				},
			}

			// Serialize to JSON
			jsonData, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())

			// Deserialize back
			var result realitydefender.DetectionResultList
			err = json.Unmarshal(jsonData, &result)
			Expect(err).NotTo(HaveOccurred())

			// Verify pagination fields
			Expect(result.TotalItems).To(Equal(1))
			Expect(result.CurrentPageItemsCount).To(Equal(1))
			Expect(result.TotalPages).To(Equal(1))
			Expect(result.CurrentPage).To(Equal(1))

			// Verify single item
			Expect(result.Items).To(HaveLen(1))
			Expect(result.Items[0].Status).To(Equal("AUTHENTIC"))
			Expect(*result.Items[0].Score).To(Equal(0.75))
			Expect(result.Items[0].Models).To(HaveLen(1))
			Expect(result.Items[0].Models[0].Name).To(Equal("singleModel"))
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
