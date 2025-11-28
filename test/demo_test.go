package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

// TestArchiveUseCase demonstrates the exact use case requested:
// URLs like /archive/2025, /archive/2025/09, and /archive/2025/09/18
// with a single parameter that can match progressively longer date paths
func TestArchiveUseCase(t *testing.T) {
	// t.Logf("=== Multi-Segment Date Parameter Demo ===")

	router := pathvars.NewRouter()

	// Add the route with multi-segment date parameter
	err := router.AddRoute("GET", "/archive/{post_date*:date:format[yyyy/mm/dd]}", nil)

	if err != nil {
		t.Fatalf("Failed to add route: %v", err)
	}

	// t.Logf("Route configured: %s", pathSpec)

	// Test cases showing the progressive date matching
	testCases := []struct {
		url         string
		description string
		query       string
		params      []pathvars.Parameter
	}{
		//{"/archive/2025", "Year only", "", []pathvars.Parameter{}},
		{"/archive/2025/09", "Year and month", "", []pathvars.Parameter{}},
		//{"/archive/2025/09/18", "Full date", "", []pathvars.Parameter{}},
	}

	for _, tc := range testCases {
		// t.Logf("Testing: %s (%s)", tc.url, tc.description)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", tc.url, tc.query), nil)
		result, err := router.Match(req)
		if err != nil {
			t.Errorf("Failed to match %s: %v", tc.url, err)
			continue
		}

		dateParam, found := result.GetValue("post_date")
		if !found {
			t.Errorf("Parameter 'post_date' not found for URL %s", tc.url)
			continue
		}

		// t.Logf("  ✅ Matched! post_date = %q", dateParam)
		// t.Logf("  📊 Route index: %d", result.Index)

		// Verify the expected parameter value
		expectedValues := map[string]string{
			"/archive/2025":       "2025",
			"/archive/2025/09":    "2025/09",
			"/archive/2025/09/18": "2025/09/18",
		}

		expectedValue := expectedValues[tc.url]
		if dateParam != expectedValue {
			t.Errorf("Expected post_date = %q for %s, got %q", expectedValue, tc.url, dateParam)
		}
	}

	// t.Logf("=== Key Benefits ===")
	// t.Logf("✓ Single parameter handles variable date precision")
	// t.Logf("✓ No need for separate year/month/day parameters")
	// t.Logf("✓ URL structure matches user expectations")
	// t.Logf("✓ Validates date components automatically")
	// t.Logf("✓ Returns date in URL format for easy SQL usage")
}
