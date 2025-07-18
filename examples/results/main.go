package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Reality-Defender/realitydefender-sdk-go/realitydefender"
)

func main() {
	// Initialize the SDK with your API key
	client, err := realitydefender.New(realitydefender.Config{
		APIKey: os.Getenv("REALITY_DEFENDER_API_KEY"),
	})
	if err != nil {
		fmt.Printf("Failed to initialize client: %v\n", err)
		os.Exit(1)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== Reality Defender GetResults Example ===\n")

	// Example 1: Get results with default parameters (all nil)
	fmt.Println("1. Getting results with default parameters...")
	results, err := client.GetResults(ctx, nil, nil, nil, nil, nil, nil)
	if err != nil {
		fmt.Printf("Failed to get results: %v\n", err)
		os.Exit(1)
	}
	printResultsSummary("Default parameters", results)

	// Example 2: Get results with pagination
	fmt.Println("\n2. Getting results with pagination (page 1, size 5)...")
	pageNumber := 1
	pageSize := 5
	results, err = client.GetResults(ctx, &pageNumber, &pageSize, nil, nil, nil, nil)
	if err != nil {
		fmt.Printf("Failed to get paginated results: %v\n", err)
		os.Exit(1)
	}
	printResultsSummary("Pagination", results)

	// Example 3: Get results with name filter
	fmt.Println("\n3. Getting results with name filter...")
	name := "test" // Filter by name containing "test"
	results, err = client.GetResults(ctx, nil, nil, &name, nil, nil, nil)
	if err != nil {
		fmt.Printf("Failed to get filtered results: %v\n", err)
		os.Exit(1)
	}
	printResultsSummary("Name filter", results)

	// Example 4: Get results with date range
	fmt.Println("\n4. Getting results with date range (last 7 days)...")
	startDate := time.Now().AddDate(0, 0, -7) // 7 days ago
	endDate := time.Now()
	results, err = client.GetResults(ctx, nil, nil, nil, &startDate, &endDate, nil)
	if err != nil {
		fmt.Printf("Failed to get date-filtered results: %v\n", err)
		os.Exit(1)
	}
	printResultsSummary("Date range", results)

	// Example 5: Get results with all parameters
	fmt.Println("\n5. Getting results with all parameters...")
	pageNumber = 1
	pageSize = 3
	name = "detection"
	startDate = time.Now().AddDate(0, 0, -30) // 30 days ago
	endDate = time.Now()
	results, err = client.GetResults(ctx, &pageNumber, &pageSize, &name, &startDate, &endDate, nil)
	if err != nil {
		fmt.Printf("Failed to get comprehensive results: %v\n", err)
		os.Exit(1)
	}
	printResultsSummary("All parameters", results)

	// Example 6: Demonstrate pagination by getting multiple pages
	fmt.Println("\n6. Demonstrating pagination across multiple pages...")
	pageSize = 2
	for page := 1; page <= 3; page++ {
		fmt.Printf("\n   Page %d:\n", page)
		pageNumber = page
		results, err = client.GetResults(ctx, &pageNumber, &pageSize, nil, nil, nil, nil)
		if err != nil {
			fmt.Printf("   Failed to get page %d: %v\n", page, err)
			continue
		}

		if len(results.Items) == 0 {
			fmt.Printf("   No more results on page %d\n", page)
			break
		}

		for i, item := range results.Items {
			fmt.Printf("   Item %d: Status=%s", i+1, item.Status)
			if item.Score != nil {
				fmt.Printf(", Score=%.4f", *item.Score)
			}
			fmt.Printf(", Models=%d\n", len(item.Models))
		}

		// Stop if we've reached the last page
		if page >= results.TotalPages {
			break
		}
	}

	fmt.Println("\n=== Example completed successfully! ===")
}

func printResultsSummary(title string, results *realitydefender.DetectionResultList) {
	fmt.Printf("   %s Results:\n", title)
	fmt.Printf("   Total Items: %d\n", results.TotalItems)
	fmt.Printf("   Current Page: %d of %d\n", results.CurrentPage, results.TotalPages)
	fmt.Printf("   Items on this page: %d\n", results.CurrentPageItemsCount)

	if len(results.Items) > 0 {
		fmt.Printf("   Sample items:\n")
		for i, item := range results.Items {
			// Limit to first 3 items for readability
			if i >= 3 {
				fmt.Printf("   ... and %d more items\n", len(results.Items)-3)
				break
			}

			fmt.Printf("     Item %d: Status=%s", i+1, item.Status)
			if item.Score != nil {
				fmt.Printf(", Score=%.4f", *item.Score)
			}
			fmt.Printf(", Models=%d\n", len(item.Models))

			// Show model details for first item
			if i == 0 && len(item.Models) > 0 {
				fmt.Printf("       Model details:\n")
				for j, model := range item.Models {
					if j >= 2 { // Limit to first 2 models
						fmt.Printf("       ... and %d more models\n", len(item.Models)-2)
						break
					}
					fmt.Printf("         %s: %s", model.Name, model.Status)
					if model.Score != nil {
						fmt.Printf(" (%.4f)", *model.Score)
					}
					fmt.Println()
				}
			}
		}
	} else {
		fmt.Printf("   No items found\n")
	}
}
