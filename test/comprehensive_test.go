package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

func TestDataTypeValidation(t *testing.T) {
	tests := []struct {
		name     string
		pathSpec pathvars.PathSpec
		method   string
		path     string
		wantErr  bool
	}{
		// String type (default)
		{"string-valid", "GET /users/{name}", "GET", "/users/john-doe", false},
		{"string-valid-default", "GET /users/{name:string}", "GET", "/users/john-doe", false},
		{"string-empty", "GET /users/{name}", "GET", "/users/", true}, // empty segment

		// Int type
		{"int-valid", "GET /users/{id:int}", "GET", "/users/123", false},
		{"int-negative", "GET /users/{id:int}", "GET", "/users/-456", false},
		{"int-invalid-letters", "GET /users/{id:int}", "GET", "/users/abc", true},
		{"int-invalid-decimal", "GET /users/{id:int}", "GET", "/users/12.5", true},

		// Decimal type
		{"decimal-valid-int", "GET /prices/{price:decimal}", "GET", "/prices/100", false},
		{"decimal-valid-float", "GET /prices/{price:decimal}", "GET", "/prices/19.99", false},
		{"decimal-negative", "GET /prices/{price:decimal}", "GET", "/prices/-5.50", false},
		{"decimal-invalid", "GET /prices/{price:decimal}", "GET", "/prices/abc", true},

		// Real type
		{"real-valid-int", "GET /measurements/{value:real}", "GET", "/measurements/100", false},
		{"real-valid-float", "GET /measurements/{value:real}", "GET", "/measurements/19.99", false},
		{"real-negative", "GET /measurements/{value:real}", "GET", "/measurements/-5.50", false},
		{"real-scientific", "GET /measurements/{value:real}", "GET", "/measurements/1.5e10", false},
		{"real-invalid", "GET /measurements/{value:real}", "GET", "/measurements/abc", true},

		// Identifier type
		{"identifier-valid", "GET /api/{resource:identifier}", "GET", "/api/users", false},
		{"identifier-underscore", "GET /api/{resource:identifier}", "GET", "/api/user_posts", false},
		{"identifier-numbers", "GET /api/{resource:identifier}", "GET", "/api/api2", false},
		{"identifier-invalid-uppercase", "GET /api/{resource:identifier}", "GET", "/api/Users", true},
		{"identifier-invalid-start-digit", "GET /api/{resource:identifier}", "GET", "/api/2users", true},
		{"identifier-invalid-hyphen", "GET /api/{resource:identifier}", "GET", "/api/user-posts", true},

		// UUID type
		{"uuid-valid", "GET /entities/{id:uuid}", "GET", "/entities/550e8400-e29b-41d4-a716-446655440000", false},
		{"uuid-invalid-short", "GET /entities/{id:uuid}", "GET", "/entities/550e8400-e29b-41d4", true},
		{"uuid-invalid-format", "GET /entities/{id:uuid}", "GET", "/entities/not-a-uuid", true},
		{"uuid-invalid-chars", "GET /entities/{id:uuid}", "GET", "/entities/550e8400-g29b-41d4-a716-446655440000", true},

		// Alphanum type
		{"alphanum-valid", "GET /codes/{code:alphanum}", "GET", "/codes/ABC123", false},
		{"alphanum-lowercase", "GET /codes/{code:alphanum}", "GET", "/codes/abc123", false},
		{"alphanum-invalid-hyphen", "GET /codes/{code:alphanum}", "GET", "/codes/ABC-123", true},
		{"alphanum-invalid-space", "GET /codes/{code:alphanum}", "GET", "/codes/ABC%20123", true},

		// Slug type
		{"slug-valid", "GET /posts/{slug:slug}", "GET", "/posts/my-post-title", false},
		{"slug-single-word", "GET /posts/{slug:slug}", "GET", "/posts/hello", false},
		{"slug-numbers", "GET /posts/{slug:slug}", "GET", "/posts/post-123", false},
		{"slug-invalid-uppercase", "GET /posts/{slug:slug}", "GET", "/posts/My-Post", true},
		{"slug-invalid-leading-hyphen", "GET /posts/{slug:slug}", "GET", "/posts/-my-post", true},
		{"slug-invalid-trailing-hyphen", "GET /posts/{slug:slug}", "GET", "/posts/my-post-", true},
		{"slug-invalid-underscore", "GET /posts/{slug:slug}", "GET", "/posts/my_post", true},

		// Boolean type
		{"boolean-true", "GET /settings/{enabled:bool}", "GET", "/settings/true", false},
		{"boolean-false", "GET /settings/{enabled:bool}", "GET", "/settings/false", false},
		{"boolean-invalid-yes", "GET /settings/{enabled:bool}", "GET", "/settings/yes", true},
		{"boolean-invalid-1", "GET /settings/{enabled:bool}", "GET", "/settings/1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			method, path := parsePathSpec(string(tt.pathSpec))
			err := router.AddRoute(pathvars.HTTPMethod(method), pathvars.Template(path), nil)
			if err != nil {
				t.Fatalf("Failed to add route: %v", err)
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			_, err = router.Match(req)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got success")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			}
		})
	}
}

func TestMethodMatching(t *testing.T) {
	tests := []struct {
		name     string
		pathSpec pathvars.PathSpec
		method   string
		path     string
		wantErr  bool
	}{
		// Specific method matching
		{"get-match", "GET /users", "GET", "/users", false},
		{"post-match", "POST /users", "POST", "/users", false},
		{"put-match", "PUT /users", "PUT", "/users", false},
		{"delete-match", "DELETE /users", "DELETE", "/users", false},

		// Method mismatch
		{"get-post-mismatch", "GET /users", "POST", "/users", true},
		{"post-get-mismatch", "POST /users", "GET", "/users", true},

		// Any method (no method specified)
		{"any-method-get", "/users", "GET", "/users", false},
		{"any-method-post", "/users", "POST", "/users", false},
		{"any-method-custom", "/users", "PATCH", "/users", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			method, path := parsePathSpec(string(tt.pathSpec))
			err := router.AddRoute(pathvars.HTTPMethod(method), pathvars.Template(path), nil)
			if err != nil {
				t.Fatalf("Failed to add route: %v", err)
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			_, err = router.Match(req)

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got success")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			}
		})
	}
}

func TestMatchResultMethods(t *testing.T) {
	router := pathvars.NewRouter()
	err := router.AddRoute("GET", "/users/{id:int}/posts/{slug:string}", &pathvars.RouteArgs{
		Index: 42,
	})
	if err != nil {
		t.Fatalf("Failed to add route: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/123/posts/my-post", nil)
	result, err := router.Match(req)
	if err != nil {
		t.Fatalf("Failed to match route: %v", err)
	}

	// Test Index
	if result.Index != 42 {
		t.Errorf("Expected Index=42, got %d", result.Index)
	}

	// Test VarCount
	if result.VarCount() != 2 {
		t.Errorf("Expected VarCount=2, got %d", result.VarCount())
	}

	// Test HasParams
	if !result.HasVars() {
		t.Error("Expected HasParams=true, got false")
	}

	// Test GetParam
	id, found := result.GetValue("id")
	if !found {
		t.Error("Expected to find 'id' parameter")
	}
	if id != "123" {
		t.Errorf("Expected id='123', got '%s'", id)
	}

	slug, found := result.GetValue("slug")
	if !found {
		t.Error("Expected to find 'slug' parameter")
	}
	if slug != "my-post" {
		t.Errorf("Expected slug='my-post', got '%s'", slug)
	}

	// Test non-existent parameter
	_, found = result.GetValue("nonexistent")
	if found {
		t.Error("Expected not to find 'nonexistent' parameter")
	}

	// Test ForEachVar
	paramMap := make(map[pathvars.Identifier]any)
	result.ForEachVar(func(name pathvars.Identifier, value any) bool {
		paramMap[name] = value
		return true
	})

	if len(paramMap) != 2 {
		t.Errorf("Expected 2 parameters in ForEachVar, got %d", len(paramMap))
	}
	if paramMap["id"] != "123" {
		t.Errorf("Expected id='123' in ForEachVar, got '%s'", paramMap["id"])
	}
	if paramMap["slug"] != "my-post" {
		t.Errorf("Expected slug='my-post' in ForEachVar, got '%s'", paramMap["slug"])
	}

	// Test ForEachVar early termination
	count := 0
	result.ForEachVar(func(name pathvars.Identifier, value any) bool {
		count++
		return false // Stop after first parameter
	})
	if count != 1 {
		t.Errorf("Expected ForEachVar to stop after 1 iteration, got %d", count)
	}
}

func TestNoParametersMatchResult(t *testing.T) {
	router := pathvars.NewRouter()
	err := router.AddRoute("GET", "/static/path", nil)
	if err != nil {
		t.Fatalf("Failed to add route: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/static/path", nil)
	result, err := router.Match(req)
	if err != nil {
		t.Fatalf("Failed to match route: %v", err)
	}

	// Test no parameters
	if result.VarCount() != 0 {
		t.Errorf("Expected VarCount=0, got %d", result.VarCount())
	}

	if result.HasVars() {
		t.Error("Expected HasParams=false, got true")
	}

	// Test ForEachVar with no parameters
	called := false
	result.ForEachVar(func(name pathvars.Identifier, value any) bool {
		called = true
		return true
	})
	if called {
		t.Error("ForEachVar should not be called when there are no parameters")
	}
}

func TestComplexPaths(t *testing.T) {
	tests := []struct {
		name     string
		pathSpec pathvars.PathSpec
		method   string
		path     string
		wantErr  bool
		expected map[pathvars.Identifier]any
	}{
		{
			name:     "multi-segment-path",
			pathSpec: "GET /api/v1/users/{id:int}/posts/{slug:string}/comments/{comment_id:int}",
			method:   "GET",
			path:     "/api/v1/users/123/posts/my-post/comments/456",
			wantErr:  false,
			expected: map[pathvars.Identifier]any{"id": "123", "slug": "my-post", "comment_id": "456"},
		},
		{
			name:     "mixed-types",
			pathSpec: "POST /users/{user_id:int}/profile/{field:identifier}/value/{active:bool}",
			method:   "POST",
			path:     "/users/42/profile/email_verified/value/true",
			wantErr:  false,
			expected: map[pathvars.Identifier]any{"user_id": "42", "field": "email_verified", "active": "true"},
		},
		{
			name:     "uuid-in-path",
			pathSpec: "GET /entities/{entity_id:uuid}/data",
			method:   "GET",
			path:     "/entities/550e8400-e29b-41d4-a716-446655440000/data",
			wantErr:  false,
			expected: map[pathvars.Identifier]any{"entity_id": "550e8400-e29b-41d4-a716-446655440000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			method, path := parsePathSpec(string(tt.pathSpec))
			err := router.AddRoute(pathvars.HTTPMethod(method), pathvars.Template(path), nil)
			if err != nil {
				t.Fatalf("Failed to add route: %v", err)
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			result, err := router.Match(req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got success")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected success but got error: %v", err)
				return
			}

			// Check expected parameters
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
