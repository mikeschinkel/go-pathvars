package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

func TestMultiSegmentParameters(t *testing.T) {
	tests := []struct {
		name         string
		method       pathvars.HTTPMethod
		path         pathvars.Template
		testPath     string
		query        string
		params       []pathvars.Parameter
		expectMatch  bool
		expectedDate string
	}{
		// Basic multi-segment date parameter tests using natural slash format
		{"year-only", "GET", "/archive/{post_date*:date:format[yyyy/mm/dd]}", "/archive/2025", "", nil, true, "2025"},
		{"year-month", "GET", "/archive/{post_date*:date:format[yyyy/mm/dd]}", "/archive/2025/09", "", nil, true, "2025/09"},
		{"full-date", "GET", "/archive/{post_date*:date:format[yyyy/mm/dd]}", "/archive/2025/09/18", "", nil, true, "2025/09/18"},

		// Invalid date tests
		{"invalid-year", "GET", "/archive/{post_date*:date:format[yyyy/mm/dd]}", "/archive/abcd", "", nil, false, ""},
		{"invalid-month", "GET", "/archive/{post_date*:date:format[yyyy/mm/dd]}", "/archive/2025/13", "", nil, false, ""},
		{"invalid-day", "GET", "/archive/{post_date*:date:format[yyyy/mm/dd]}", "/archive/2025/09/32", "", nil, false, ""},

		// Different date format tests
		{"us-format-month-day", "GET", "/posts/{date*:date:format[mm/dd/yyyy]}", "/posts/09/18", "", nil, true, "09/18"},
		{"us-format-full", "GET", "/posts/{date*:date:format[mm/dd/yyyy]}", "/posts/09/18/2025", "", nil, true, "09/18/2025"},

		// Multi-segment with literal segments after
		{"with-trailing-literal", "GET", "/archive/{date*:date:format[yyyy/mm/dd]}/list", "/archive/2025/09/18/list", "", nil, true, "2025/09/18"},
		{"with-trailing-literal-partial", "GET", "/archive/{date*:date:format[yyyy/mm/dd]}/list", "/archive/2025/list", "", nil, true, "2025"},

		// Edge cases
		{"too-many-segments", "GET", "/archive/{date*:date:format[yyyy/mm/dd]}", "/archive/2025/09/18/extra", "", nil, false, ""},
		{"empty-segment", "GET", "/archive/{date*:date:format[yyyy/mm/dd]}", "/archive//09/18", "", nil, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			err := router.AddRoute(tt.method, tt.path, &pathvars.RouteArgs{
				Parameters: tt.params,
			})
			if err != nil {
				t.Fatalf("Failed to add route: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", tt.testPath, tt.query), nil)
			result, err := router.Match(req)

			if tt.expectMatch {
				if err != nil {
					t.Errorf("Expected match but got error:\n%v", err)
					return
				}

				actualDate, found := result.GetValue("post_date")
				if !found {
					actualDate, found = result.GetValue("date")
				}
				if !found {
					t.Errorf("Expected to find date parameter")
					return
				}

				if actualDate != tt.expectedDate {
					t.Errorf("Expected date '%s', got '%s'", tt.expectedDate, actualDate)
				}
			} else {
				if err == nil {
					t.Errorf("Expected no match but got success")
				}
			}
		})
	}
}

func TestMultiSegmentParameterParsing(t *testing.T) {
	tests := []struct {
		name                 string
		method               pathvars.HTTPMethod
		path                 pathvars.Template
		query                string
		params               []pathvars.Parameter
		expectError          bool
		expectedParamName    string
		expectedMultiSegment bool
	}{
		{"regular-param", "GET", "/users/{id:int}", "", nil, false, "id", false},
		{"multi-segment-param", "GET", "/archive/{date*:date:format[yyyy/mm/dd]}", "", nil, false, "date", true},
		{"multi-segment-no-type", "GET", "/path/{param*}", "", nil, false, "param", true},
		{"multi-segment-empty-name", "GET", "/path/{*}", "", nil, true, "", false},
		{"multi-segment-only-star", "GET", "/path/{*:string}", "", nil, true, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			err := router.AddRoute(tt.method, tt.path, &pathvars.RouteArgs{
				Parameters: tt.params,
			})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

		})
	}
}

func TestMultiSegmentRegexGeneration(t *testing.T) {
	tests := []struct {
		name     string
		method   pathvars.HTTPMethod
		path     pathvars.Template
		testPath string
		query    string
		params   []pathvars.Parameter
		matches  bool
	}{
		// Test that multi-segment regexes work correctly
		{"basic-multi-segment", "GET", "/archive/{date*}", "/archive/2025/09/18", "", nil, true},
		{"single-segment-capture", "GET", "/archive/{date*}", "/archive/2025", "", nil, true},
		{"no-segments", "GET", "/archive/{date*}", "/archive", "", nil, false},
		{"mixed-params", "GET", "/api/{version}/{data*}", "/api/v1/users/123/posts", "", nil, true},
		{"multi-segment-at-end", "GET", "/files/{path*}", "/files/docs/readme.txt", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			err := router.AddRoute(tt.method, tt.path, &pathvars.RouteArgs{
				Parameters: tt.params,
			})
			if err != nil {
				t.Fatalf("Failed to add route: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", tt.testPath, tt.query), nil)
			_, err = router.Match(req)

			if tt.matches && err != nil {
				t.Errorf("Expected match but got error:\n%v", err)
			} else if !tt.matches && err == nil {
				t.Errorf("Expected no match but got success")
			}
		})
	}
}
