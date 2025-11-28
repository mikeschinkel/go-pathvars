package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

func TestTemplateParsing(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		expectError bool
		description string
	}{
		// Valid templates
		{"simple-literal", "/users", false, "Simple literal path"},
		{"single-param", "/users/{id}", false, "Single parameter"},
		{"param-with-type", "/users/{id:int}", false, "Parameter with type"},
		{"multiple-params", "/users/{id:int}/posts/{slug:string}", false, "Multiple parameters"},
		{"mixed-segments", "/api/v1/users/{id:int}/profile", false, "Mixed literal and parameter segments"},

		// Complex valid cases
		{"all-types", "/test/{str:string}/{num:int}/{dec:decimal}/{ident:identifier}/{uuid:uuid}/{alpha:alphanum}/{slug:slug}/{bool:bool}", false, "All supported types"},
		{"long-path", "/very/long/path/with/many/{segments:string}/and/{parameters:int}/here", false, "Long path with many segments"},

		// Auto-corrected templates (no longer errors)
		{"empty-template", "", false, "Empty template gets converted to '/'"},
		{"no-leading-slash", "users/{id}", false, "No leading slash gets auto-corrected"},
		{"empty-braces", "/users/{}", true, "Empty parameter braces"},
		{"unmatched-open", "/users/{id", true, "Unmatched opening brace"},
		{"unmatched-close", "/users/id}", true, "Unmatched closing brace - now consistently an error"},
		{"nested-braces", "/users/{{id}}", true, "Nested braces not allowed (at this time, maybe later if needed)"},

		// Invalid parameter definitions
		{"no-param-name", "/users/{:int}", true, "No parameter name"},
		{"invalid-type", "/users/{id:invalid}", true, "Invalid parameter type"},
		{"extra-colons", "/users/{id:int:extra:stuff}", true, "Extra colons - invalid constraints should cause error"},

		// Edge cases
		{"single-slash", "/", false, "Root path"},
		{"trailing-slash", "/users/", false, "Trailing slash"},
		{"multiple-slashes", "//users", false, "Multiple consecutive slashes"},
		{"param-at-start", "/{category}/items", false, "Parameter at start"},
		{"param-at-end", "/items/{id}", false, "Parameter at end"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			err := router.AddRoute("GET", pathvars.Template(tt.template), nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for template '%s' but got none", tt.template)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for template '%s' but got: %v", tt.template, err)
				}
			}
		})
	}
}

func TestParameterExtraction(t *testing.T) {
	tests := []struct {
		name       string
		template   string
		testPath   string
		expected   map[pathvars.Identifier]string
		params     []pathvars.Parameter
		query      string
		shouldFail bool
	}{
		{
			name:     "single-param",
			template: "/users/{id}",
			testPath: "/users/123",
			expected: map[pathvars.Identifier]string{"id": "123"},
		},
		{
			name:     "multiple-params",
			template: "/users/{user_id}/posts/{post_id}",
			testPath: "/users/42/posts/789",
			expected: map[pathvars.Identifier]string{"user_id": "42", "post_id": "789"},
		},
		{
			name:     "param-with-special-chars",
			template: "/files/{filename}",
			testPath: "/files/my-file.txt",
			expected: map[pathvars.Identifier]string{"filename": "my-file.txt"},
		},
		{
			name:     "param-with-numbers",
			template: "/api/{version}/users/{id}",
			testPath: "/api/v2/users/123",
			expected: map[pathvars.Identifier]string{"version": "v2", "id": "123"},
		},
		{
			name:     "param-with-underscores",
			template: "/users/{user_id}/settings/{setting_name}",
			testPath: "/users/123/settings/email_notifications",
			expected: map[pathvars.Identifier]string{"user_id": "123", "setting_name": "email_notifications"},
		},
		{
			name:     "complex-path",
			template: "/organizations/{org_id}/projects/{project_id}/issues/{issue_number}/comments/{comment_id}",
			testPath: "/organizations/acme/projects/webapp/issues/42/comments/1",
			expected: map[pathvars.Identifier]string{
				"org_id":       "acme",
				"project_id":   "webapp",
				"issue_number": "42",
				"comment_id":   "1",
			},
		},

		// Cases that should fail parameter validation
		{
			name:       "int-param-invalid",
			template:   "/users/{id:int}",
			testPath:   "/users/abc",
			shouldFail: true,
		},
		{
			name:       "uuid-param-invalid",
			template:   "/entities/{id:uuid}",
			testPath:   "/entities/not-a-uuid",
			shouldFail: true,
		},
		{
			name:       "boolean-param-invalid",
			template:   "/settings/{enabled:bool}",
			testPath:   "/settings/yes",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			err := router.AddRoute("GET", pathvars.Template(tt.template), &pathvars.RouteArgs{
				Parameters: tt.params,
			})
			if err != nil {
				t.Fatalf("Failed to add route: %v", err)
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", tt.testPath, tt.query), nil)
			result, err := router.Match(req)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected parameter validation to fail but it succeeded")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected success but got error: %v", err)
				return
			}

			// Check all expected parameters
			for name, expectedValue := range tt.expected {
				actualValue, found := result.GetValue(name)
				if !found {
					t.Errorf("Expected to find parameter '%s'", name)
				} else if actualValue != expectedValue {
					t.Errorf("Parameter '%s': expected '%s', got '%s'", name, expectedValue, actualValue)
				}
			}

			// Check parameter count
			if result.VarCount() != len(tt.expected) {
				t.Errorf("Expected %d parameters, got %d", len(tt.expected), result.VarCount())
			}
		})
	}
}

func TestRegexGeneration(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		shouldMatch    []string
		shouldNotMatch []string
		params         []pathvars.Parameter
		query          string
	}{
		{
			name:           "literal-only",
			template:       "/api/v1/users",
			shouldMatch:    []string{"/api/v1/users"},
			shouldNotMatch: []string{"/api/v1/user", "/api/v1/users/", "/api/v2/users", "/api/v1/posts"},
		},
		{
			name:           "single-param",
			template:       "/users/{id}",
			shouldMatch:    []string{"/users/123", "/users/abc", "/users/user-name"},
			shouldNotMatch: []string{"/users", "/users/", "/users/123/posts", "/posts/123"},
		},
		{
			name:           "multiple-params",
			template:       "/users/{id}/posts/{slug}",
			shouldMatch:    []string{"/users/123/posts/my-post", "/users/abc/posts/hello-world"},
			shouldNotMatch: []string{"/users/123/posts", "/users/123", "/users/123/posts/my-post/extra"},
		},
		{
			name:           "mixed-literal-param",
			template:       "/api/v1/users/{id}/profile",
			shouldMatch:    []string{"/api/v1/users/123/profile", "/api/v1/users/abc/profile"},
			shouldNotMatch: []string{"/api/v1/users/123", "/api/v1/users/123/profile/extra", "/api/v2/users/123/profile"},
		},
		{
			name:           "special-chars-in-literal",
			template:       "/api-v1.0/users/{category}",
			shouldMatch:    []string{"/api-v1.0/users/premium", "/api-v1.0/users/basic"},
			shouldNotMatch: []string{"/api_v1.0/users/premium", "/api-v1.0/users", "/api-v1.0/user/premium"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			err := router.AddRoute("GET", pathvars.Template(tt.template), &pathvars.RouteArgs{
				Parameters: tt.params,
			})
			if err != nil {
				t.Fatalf("Failed to add route: %v", err)
			}

			// Test paths that should match
			for _, path := range tt.shouldMatch {
				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", path, tt.query), nil)
				_, err = router.Match(req)
				if err != nil {
					t.Errorf("Path '%s' should match template '%s' but got error: %v", path, tt.template, err)
				}
			}

			// Test paths that should not match
			for _, path := range tt.shouldNotMatch {
				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", path, tt.query), nil)
				_, err = router.Match(req)
				if err == nil {
					t.Errorf("Path '%s' should NOT match template '%s' but it did", path, tt.template)
				}
			}
		})
	}
}

func TestParameterTypes(t *testing.T) {
	typeTests := map[string]struct {
		validValues   []string
		invalidValues []string
		params        []pathvars.Parameter
		query         string
	}{
		"string": {
			validValues:   []string{"hello", "world", "123", "test-value", "test_value", "test.value"},
			invalidValues: []string{}, // All strings are valid
		},
		"integer": {
			validValues:   []string{"123", "-456", "0", "999999"},
			invalidValues: []string{"abc", "12.5", "1.0", "hello", "123abc"},
		},
		"decimal": {
			validValues:   []string{"123", "12.5", "-45.67", "0.0", "999.999"},
			invalidValues: []string{"abc", "12.5.6", "hello", "123abc"},
		},
		"identifier": {
			validValues:   []string{"hello", "user_name", "api2", "test_123"},
			invalidValues: []string{"Hello", "2user", "user-name", "123", "user%20name"},
		},
		"uuid": {
			validValues:   []string{"550e8400-e29b-41d4-a716-446655440000", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
			invalidValues: []string{"550e8400-e29b-41d4", "not-a-uuid", "550e8400e29b41d4a716446655440000", "550e8400-g29b-41d4-a716-446655440000"},
		},
		"alphanumeric": {
			validValues:   []string{"ABC123", "hello", "WORLD", "abc", "123"},
			invalidValues: []string{"hello-world", "test_value", "hello%20world", "test.value"},
		},
		"slug": {
			validValues:   []string{"hello-world", "my-post", "test", "post-123"},
			invalidValues: []string{"Hello-World", "-hello", "hello-", "hello_world", "hello%20world"},
		},
		"boolean": {
			validValues:   []string{"true", "false"},
			invalidValues: []string{"yes", "no", "1", "0", "TRUE", "FALSE", "True", "False"},
		},
	}

	for name, tt := range typeTests {
		t.Run("type-"+name, func(t *testing.T) {
			template := "/test/{value:" + name + "}"
			router := pathvars.NewRouter()
			err := router.AddRoute("GET", pathvars.Template(template), &pathvars.RouteArgs{
				Parameters: tt.params,
			})

			if err != nil {
				t.Fatalf("Failed to add route: %v", err)
			}

			// Test valid values
			for _, value := range tt.validValues {
				path := "/test/" + value
				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", path, tt.query), nil)
				result, err := router.Match(req)
				if err != nil {
					t.Errorf("Valid %s value '%s' should match but got error: %v", name, value, err)
				} else {
					extractedValue, found := result.GetValue("value")
					if !found {
						t.Errorf("Parameter 'value' not found for %s value '%s'", name, value)
					} else if extractedValue != value {
						t.Errorf("Expected extracted value '%s', got '%s'", value, extractedValue)
					}
				}
			}

			// Test invalid values
			for _, value := range tt.invalidValues {
				path := "/test/" + value
				req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", path, tt.query), nil)
				_, err = router.Match(req)
				if err == nil {
					t.Errorf("Invalid %s value '%s' should NOT match but it did", name, value)
				}
			}
		})
	}
}
