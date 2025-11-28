package test

import (
	"net/http/httptest"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

// TestValueDecomposition tests that multi-segment parameters are automatically
// decomposed into their component parts.
func TestValueDecomposition(t *testing.T) {
	tests := []struct {
		name           string
		template       pathvars.Template
		testPath       string
		expectedValues map[string]string
	}{
		{
			name:     "date full decomposition",
			template: "/archive/{date*:date}",
			testPath: "/archive/2025/10/15",
			expectedValues: map[string]string{
				"date":       "2025/10/15",
				"date_year":  "2025",
				"date_month": "10",
				"date_day":   "15",
			},
		},
		{
			name:     "date partial year only",
			template: "/archive/{date*:date}",
			testPath: "/archive/2025",
			expectedValues: map[string]string{
				"date":      "2025",
				"date_year": "2025",
			},
		},
		{
			name:     "date partial year-month",
			template: "/archive/{date*:date}",
			testPath: "/archive/2025/10",
			expectedValues: map[string]string{
				"date":       "2025/10",
				"date_year":  "2025",
				"date_month": "10",
			},
		},
		{
			name:     "string multi-segment numeric suffixes",
			template: "/files/{path*:string}",
			testPath: "/files/docs/readme.txt",
			expectedValues: map[string]string{
				"path":   "docs/readme.txt",
				"path_1": "docs",
				"path_2": "readme.txt",
			},
		},
		{
			name:     "string single segment",
			template: "/files/{path*:string}",
			testPath: "/files/readme.txt",
			expectedValues: map[string]string{
				"path":   "readme.txt",
				"path_1": "readme.txt",
			},
		},
		{
			name:     "string many segments",
			template: "/api/{resource*:string}",
			testPath: "/api/v1/users/123/posts",
			expectedValues: map[string]string{
				"resource":   "v1/users/123/posts",
				"resource_1": "v1",
				"resource_2": "users",
				"resource_3": "123",
				"resource_4": "posts",
			},
		},
		{
			name:     "multiple multi-segment params",
			template: "/archive/{date*:date}/category/{cat*:string}",
			testPath: "/archive/2025/10/15/category/tech/golang",
			expectedValues: map[string]string{
				"date":       "2025/10/15",
				"date_year":  "2025",
				"date_month": "10",
				"date_day":   "15",
				"cat":        "tech/golang",
				"cat_1":      "tech",
				"cat_2":      "golang",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			err := router.AddRoute("GET", tt.template, nil)
			if err != nil {
				t.Fatalf("Failed to add route: %v", err)
			}

			req := httptest.NewRequest("GET", tt.testPath, nil)
			result, err := router.Match(req)
			if err != nil {
				t.Fatalf("Expected match but got error: %v", err)
			}

			// Verify all expected values are present
			for key, expectedValue := range tt.expectedValues {
				actualValue, found := result.GetValue(pathvars.Identifier(key))
				if !found {
					t.Errorf("Expected to find key %q in values map", key)
					continue
				}

				actualStr, ok := actualValue.(string)
				if !ok {
					t.Errorf("Value for %q is not a string: %T", key, actualValue)
					continue
				}

				if actualStr != expectedValue {
					t.Errorf("For key %q: expected %q, got %q", key, expectedValue, actualStr)
				}
			}

			// Verify no unexpected values are present
			result.ForEachVar(func(name pathvars.Identifier, value any) bool {
				if _, expected := tt.expectedValues[string(name)]; !expected {
					t.Errorf("Unexpected key in values map: %q = %v", name, value)
				}
				return true
			})
		})
	}
}

// TestValueDecompositionWithFormatConstraint tests decomposition works with format constraints
func TestValueDecompositionWithFormatConstraint(t *testing.T) {
	router := pathvars.NewRouter()
	err := router.AddRoute("GET", "/posts/{date*:date:format[yyyy/mm/dd]}", nil)
	if err != nil {
		t.Fatalf("Failed to add route: %v", err)
	}

	req := httptest.NewRequest("GET", "/posts/2025/09/18", nil)
	result, err := router.Match(req)
	if err != nil {
		t.Fatalf("Expected match but got error: %v", err)
	}

	expectedValues := map[string]string{
		"date":       "2025/09/18",
		"date_year":  "2025",
		"date_month": "09",
		"date_day":   "18",
	}

	for key, expectedValue := range expectedValues {
		actualValue, found := result.GetValue(pathvars.Identifier(key))
		if !found {
			t.Errorf("Expected to find key %q", key)
			continue
		}

		actualStr, ok := actualValue.(string)
		if !ok {
			t.Errorf("Value for %q is not a string: %T", key, actualValue)
			continue
		}

		if actualStr != expectedValue {
			t.Errorf("For key %q: expected %q, got %q", key, expectedValue, actualStr)
		}
	}
}
