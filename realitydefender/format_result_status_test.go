package realitydefender

import "testing"

func TestFormatResultNullSummaryUsesOverallStatus(t *testing.T) {
	response := MediaResponse{
		RequestID:     "req-social",
		OverallStatus: "DOWNLOADING",
	}

	result := FormatResult(&response)

	if result.Status != "DOWNLOADING" {
		t.Fatalf("expected DOWNLOADING, got %q", result.Status)
	}
}

func TestFormatResultNullSummaryUsesAnalyzingOverallStatus(t *testing.T) {
	response := MediaResponse{
		RequestID:     "req-social",
		OverallStatus: "ANALYZING",
	}

	result := FormatResult(&response)

	if result.Status != "ANALYZING" {
		t.Fatalf("expected ANALYZING, got %q", result.Status)
	}
}
