package test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func TestRouterErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		method      pathvars.HTTPMethod
		path        pathvars.Template
		expectError bool
		errorType   error
		query       string
		params      []pathvars.Parameter
	}{
		// Valid cases
		{"valid-simple", "GET", "/users", false, nil, "", nil},
		{"valid-with-param", "GET", "/users/{id:int}", false, nil, "", nil},
		{"valid-no-method", "", "/users", false, nil, "", nil},

		// Path specs that are auto-corrected by router
		{"empty-spec", "", "", false, nil, "", nil},
		{"no-leading-slash", "GET", "users", false, nil, "", nil},
		{"empty-path-after-method", "GET", "", false, nil, "", nil},

		// Invalid parameter syntax
		{"empty-braces", "GET", "/users/{}", true, pvtypes.ErrInvalidParameter, "", nil},
		{"unmatched-open-brace", "GET", "/users/{id", true, pvtypes.ErrInvalidParameter, "", nil},
		{"unmatched-close-brace", "GET", "/users/id}", true, pvtypes.ErrInvalidParameter, "", nil}, // Now consistent - error like unmatched opening
		{"no-param-name", "GET", "/users/{:int}", true, pvtypes.ErrNameSpecNameCannotBeEmpty, "", nil},

		// Invalid types
		{"invalid-type", "GET", "/users/{id:invalid}", true, pvtypes.ErrInvalidParameterType, "", nil},
		{"typo-in-type", "GET", "/users/{id:integr}", true, pvtypes.ErrInvalidParameterType, "", nil},
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
					return
				}
				if tt.errorType != nil && !errors.Is(err, tt.errorType) {
					t.Errorf("Expected error type %v, got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestNoMatchingRoute(t *testing.T) {
	router := pathvars.NewRouter()
	err := router.AddRoute("GET", "/users/{id}", nil)
	if err != nil {
		t.Fatalf("Failed to add route: %v", err)
	}

	tests := []struct {
		name   string
		method string
		path   string
		query  string
	}{
		{"different-path", "GET", "/posts/123", ""},
		{"different-method", "POST", "/users/123", ""},
		{"wrong-segments", "GET", "/users", ""},
		{"extra-segments", "GET", "/users/123/extra", ""},
		{"empty-path", "GET", "/", ""},
		{"root-path", "GET", "/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, fmt.Sprintf("%s?%s", tt.path, tt.query), nil)
			_, err = router.Match(req)
			if err == nil {
				t.Error("Expected ErrNoMatch but got no error")
			}
			if !errors.Is(err, pathvars.ErrNoMatch) {
				t.Errorf("Expected ErrNoMatch, got %v", err)
			}
		})
	}
}

func TestMultipleRoutes(t *testing.T) {
	router := pathvars.NewRouter()

	routes := []struct {
		method pathvars.HTTPMethod
		path   pathvars.Template
		index  int
		params []pathvars.Parameter
	}{
		{"POST", "/users", 1, nil},
		{"GET", "/users/{id:int}", 2, nil},
		{"PUT", "/users/{id:int}", 3, nil},
		{"GET", "/posts/{slug:slug}", 4, nil},
		{"", "/health", 5, nil}, // Any method
		{"GET", "/users", 6, nil},
	}

	for _, route := range routes {
		err := router.AddRoute(route.method, route.path, &pathvars.RouteArgs{
			Parameters: route.params,
			Index:      route.index,
		})
		if err != nil {
			t.Fatalf("Failed to add route %s %s: %v", route.method, route.path, err)
		}
	}

	tests := []struct {
		name          string
		method        string
		path          string
		query         string
		expectedIndex int
		wantErr       bool
	}{
		{"post-users", "POST", "/users", "", 1, false},
		{"get-user-by-id", "GET", "/users/123", "", 2, false},
		{"put-user", "PUT", "/users/456", "", 3, false},
		{"get-post-by-slug", "GET", "/posts/my-post", "", 4, false},
		{"health-get", "GET", "/health", "", 5, false},
		{"health-post", "POST", "/health", "", 5, false},
		{"health-any", "PATCH", "/health", "", 5, false},
		{"get-users-list", "GET", "/users", "", 6, false},

		// Error cases
		{"delete-users", "DELETE", "/users", "", -1, true},
		{"invalid-user-id", "GET", "/users/abc", "", -1, true},
		{"invalid-slug", "GET", "/posts/Invalid-Slug", "", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := router.Match(httptest.NewRequest(tt.method, fmt.Sprintf("%s?%s", tt.path, tt.query), nil))

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got success")
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
					return
				}
				if result.Index != tt.expectedIndex {
					t.Errorf("Expected index %d, got %d", tt.expectedIndex, result.Index)
				}
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		pathSpec pathvars.PathSpec
		method   string
		path     string
		query    string
		params   []pathvars.Parameter
		wantErr  bool
	}{
		// Path edge cases
		{"single-slash", "GET /", "GET", "/", "", nil, true},                    // Root path with no segments
		{"multiple-slashes", "GET //users", "GET", "//users", "", nil, true},    // Multiple slashes create empty segments
		{"trailing-slash-spec", "GET /users/", "GET", "/users/", "", nil, true}, // Trailing slash creates empty segment
		{"trailing-slash-path", "GET /users", "GET", "/users/", "", nil, true},  // Different paths

		// Parameter edge cases
		{"single-char-param", "GET /{a}", "GET", "/x", "", nil, false},
		{"long-param-name", "GET /{very_long_parameter_name_here}", "GET", "/value", "", nil, false},
		{"numbers-in-param", "GET /{param123}", "GET", "/value", "", nil, false},

		// Special characters in literal segments
		{"dash-in-literal", "GET /api-v1/users", "GET", "/api-v1/users", "", nil, false},
		{"underscore-in-literal", "GET /api_v1/users", "GET", "/api_v1/users", "", nil, false},
		{"dot-in-literal", "GET /api.v1/users", "GET", "/api.v1/users", "", nil, false},

		// Case sensitivity
		{"case-sensitive-method", "get /users", "GET", "/users", "", nil, true}, // Method case matters
		{"case-sensitive-path", "GET /Users", "GET", "/users", "", nil, true},   // Path case matters
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			m, p := parsePathSpec(string(tt.pathSpec))
			err := router.AddRoute(pathvars.HTTPMethod(m), pathvars.Template(p), &pathvars.RouteArgs{
				Parameters: tt.params,
			})
			if err != nil {
				// Some edge cases might fail at add time
				if !tt.wantErr {
					t.Errorf("Unexpected error adding route: %v", err)
				}
				return
			}

			_, err = router.Match(httptest.NewRequest(tt.method, fmt.Sprintf("%s?%s", tt.path, tt.query), nil))

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got success")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Expected success but got error: %v", err)
			}
		})
	}
}

func TestParameterParsing(t *testing.T) {
	tests := []struct {
		name         string
		method       pathvars.HTTPMethod
		path         pathvars.Template
		testPath     string
		query        string
		params       []pathvars.Parameter
		expectedName pathvars.Identifier
		expectedType string
		wantErr      bool
	}{
		// Basic parameter parsing
		{"name-only", "GET", "/{name}", "/test", "", nil, "name", "string", false},
		{"name-with-type", "GET", "/{id:int}", "/123", "", nil, "id", "integer", false},
		{"complex-name", "GET", "/{user_id:int}", "/456", "", nil, "user_id", "integer", false},

		// All supported types
		{"type-string", "GET", "/{value:string}", "/hello", "", nil, "value", "string", false},
		{"type-int", "GET", "/{value:int}", "/42", "", []pathvars.Parameter{}, "value", "integer", false},
		{"type-decimal", "GET", "/{value:decimal}", "/3.14", "", []pathvars.Parameter{}, "value", "decimal", false},
		{"type-real", "GET", "/{value:real}", "/2.71", "", []pathvars.Parameter{}, "value", "real", false},
		{"type-identifier", "GET", "/{value:identifier}", "/valid_id", "", []pathvars.Parameter{}, "value", "identifier", false},
		{"type-uuid", "GET", "/{value:uuid}", "/550e8400-e29b-41d4-a716-446655440000", "", []pathvars.Parameter{}, "value", "uuid", false},
		{"type-alphanum", "GET", "/{value:alphanum}", "/ABC123", "", []pathvars.Parameter{}, "value", "alphanumeric", false},
		{"type-slug", "GET", "/{value:slug}", "/my-slug", "", []pathvars.Parameter{}, "value", "slug", false},
		{"type-bool", "GET", "/{value:bool}", "/true", "", []pathvars.Parameter{}, "value", "boolean", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			err := router.AddRoute(tt.method, tt.path, &pathvars.RouteArgs{
				Parameters: tt.params,
			})
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Unexpected error adding route: %v", err)
				}
				return
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s", tt.testPath, tt.query), nil)
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

			// Verify parameter was extracted correctly
			value, found := result.GetValue(tt.expectedName)
			if !found {
				t.Errorf("Expected to find parameter '%s'", tt.expectedName)
				return
			}

			expectedValue := tt.testPath[1:] // Remove leading slash
			if value != expectedValue {
				t.Errorf("Expected parameter value '%s', got '%s'", expectedValue, value)
			}
		})
	}
}
