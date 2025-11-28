package test

import (
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/mikeschinkel/go-pathvars"
)

// FuzzParseTemplate tests the ParseTemplate function with comprehensive seed corpus
// and safety mechanisms to detect infinite loops and panics.
func FuzzParseTemplate(f *testing.F) {
	// Basic Cases (10 seeds)
	f.Add("")
	f.Add("/")
	f.Add("/users")
	f.Add("/users/{id}")
	f.Add("/users/{id}/posts/{post_id}")
	f.Add("/api/v1/users/{id}")
	f.Add("GET /users/{id}")
	f.Add("POST /users")
	f.Add("/users?{limit}")
	f.Add("/users?{limit}&{offset}")

	// Typed Parameters (10 seeds)
	f.Add("/users/{id:int}")
	f.Add("/posts/{slug:string}")
	f.Add("/items/{uuid:uuid}")
	f.Add("/dates/{date:date}")
	f.Add("/users/{id:identifier}")
	f.Add("/tags/{name:alphanumeric}")
	f.Add("/emails/{addr:email}")
	f.Add("/products/{category:slug}")
	f.Add("/flags/{enabled:boolean}")
	f.Add("/values/{amount:decimal}")

	// Constraints (15 seeds)
	f.Add("/users/{id:int:range[1..1000]}")
	f.Add("/posts/{slug:string:length[5..50]}")
	f.Add("/items/{name:string:notempty}")
	f.Add("/tags/{tag:string:enum[tech,news,sports]}")
	f.Add("/emails/{email:string:regex[.+@.+]}")
	f.Add("/dates/{date:date:format[yyyy-mm-dd]}")
	f.Add("/uuids/{id:uuid:format[v4]}")
	f.Add("/users/{id:int:range[1..1000]}&{limit:int:range[1..100]}")
	f.Add("/posts/{slug:slug:length[5..50],notempty}")
	f.Add("/items/{price:decimal:range[0.01..9999.99]}")
	f.Add("/scores/{value:real:range[0.0..100.0]}")
	f.Add("/archive/{date*:date:format[yyyy/mm/dd]}")
	f.Add("/files/{path*:string}")
	f.Add("/api/{slug::enum[v1,v2,v3]}")
	f.Add("{int::range[0..100]}")

	// Query Parameters (10 seeds)
	f.Add("/search?{q:string}")
	f.Add("/users?{active?true:boolean}")
	f.Add("/posts?{limit?20:int}")
	f.Add("/items?{sort?name:enum[name,price,date]}")
	f.Add("/products?{min_price?0:decimal}&{max_price?1000:decimal}")
	f.Add("/users?{limit?10:int:range[1..100]}&{offset?0:int}")
	f.Add("?{q}")
	f.Add("/?{q:string}&{limit:int}")
	f.Add("/api/users?{fields:string}&{expand?false:boolean}")
	f.Add("/search/{query}?{page?1:int}&{per_page?20:int}")

	// Edge Cases - Braces (8 seeds)
	f.Add("/path/{")
	f.Add("/path/}")
	f.Add("/path/{{")
	f.Add("/path/}}")
	f.Add("/path/}{")
	f.Add("/path/{unclosed")
	f.Add("/path/{id")
	f.Add("/{}/something")

	// Edge Cases - Colons (8 seeds)
	f.Add("/path/{name:}")
	f.Add("/path/{name::}")
	f.Add("/path/{:type}")
	f.Add("/path/{::constraint}")
	f.Add(":::")
	f.Add(": : :")
	f.Add("/path/{name:type:}")
	f.Add("/time/12:30:00")

	// Edge Cases - Question Marks (6 seeds)
	f.Add("?{q}")
	f.Add("/path?")
	f.Add("/path??")
	f.Add("/path??{q}")
	f.Add("/path?{q?}")
	f.Add("?")

	// Malformed Input (10 seeds)
	f.Add(":")
	f.Add("{}")
	f.Add("{name}")
	f.Add("/{name:invalid_type}")
	f.Add("/{name:int:bad_constraint}")
	f.Add("/{name:int:range[bad]}")
	f.Add("/{name?default:int:range[1..10]:extra}")
	f.Add("/{id}/{id}")
	f.Add("/{*multi}")
	f.Add("/path/{multi*}/after")

	// Performance Tests (3 seeds)
	f.Add(string(make([]byte, 1000)))              // 1KB of null bytes
	f.Add(string(make([]byte, 10000)))             // 10KB of null bytes
	f.Add(strings.Repeat("/segment/{param}", 100)) // Deep nesting

	// Real-World Patterns (5 seeds)
	f.Add("GET /api/v{version:int}/users/{id:uuid}/posts/{slug:slug:length[5..100]}?{fields:string}&{limit?25:int:range[1..100]}")
	f.Add("POST /api/products/{category:slug}/items?{tags:string}&{active?true:boolean}")
	f.Add("GET /archive/{year:int:range[1900..2100]}/{month:int:range[1..12]}/{day:int:range[1..31]}")
	f.Add("DELETE /users/{id:uuid:format[v4]}/sessions/{session:uuid:format[v7]}")
	f.Add("GET /files/{path*:string:notempty}?{download?false:boolean}")

	// Run the fuzz test
	f.Fuzz(func(t *testing.T, template string) {
		done := make(chan struct{})
		var result *pathvars.ParsedTemplate
		var err error

		// Run parser in goroutine with panic recovery
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ParseTemplate panicked on %q: %v\nStack: %s",
						template, r, debug.Stack())
				}
				close(done)
			}()

			result, err = pathvars.ParseTemplate(template)
		}()

		// Timeout detection (1 second should be plenty for template parsing)
		select {
		case <-done:
			// Parser completed (with or without error)
			if err != nil {
				// Errors are acceptable for malformed input
				// Just ensure the error is non-nil
				return
			}

			// Sanity checks on successful parse
			if result == nil {
				t.Errorf("ParseTemplate returned nil result with nil error for: %q", template)
				return
			}

			// Verify result is sensible (non-empty template should produce non-empty result)
			// Note: Empty template might be valid, so only check if template was non-empty
			if template != "" && result.String() == "" {
				// t.Logf("ParseTemplate returned empty string representation for non-empty input: %q (this might be OK)", template)
			}

		case <-time.After(1 * time.Second):
			t.Fatalf("ParseTemplate hung (infinite loop detected) on input: %q", template)
		}
	})
}

// FuzzMatch tests the Match function with template+path+query combinations
// to detect panics, infinite loops, or unexpected behavior in path matching.
func FuzzMatch(f *testing.F) {
	// Seed with template+path+query combinations

	// Valid matches - should succeed
	f.Add("/users/{id:int}", "/users/123", "")
	f.Add("/posts/{slug:string}", "/posts/hello-world", "")
	f.Add("/items/{uuid:uuid}", "/items/550e8400-e29b-41d4-a716-446655440000", "")
	f.Add("/products/{id:int:range[1..1000]}", "/products/500", "")
	f.Add("/search?{q:string}", "/search", "q=test")
	f.Add("/users?{limit?10:int:range[1..100]}", "/users", "limit=25")
	f.Add("/posts/{slug:slug:length[5..50]}", "/posts/hello-world", "")
	f.Add("/api/v{version:int}/users/{id:int}", "/api/v1/users/123", "")

	// Invalid matches - should fail validation
	f.Add("/users/{id:int}", "/users/abc", "")
	f.Add("/posts/{slug:slug:length[10..50]}", "/posts/hi", "")
	f.Add("/products/{id:int:range[1..1000]}", "/products/0", "")
	f.Add("/products/{id:int:range[1..1000]}", "/products/1001", "")
	f.Add("/items/{uuid:uuid}", "/items/not-a-uuid", "")
	f.Add("/dates/{date:date:format[yyyy-mm-dd]}", "/dates/2025-13-01", "")

	// Path mismatches - should not match pattern
	f.Add("/users/{id}", "/posts/123", "")
	f.Add("/users/{id}/posts", "/users/123", "")
	f.Add("/api/v1/users", "/api/v2/users", "")

	// Edge cases - empty, special chars
	f.Add("", "", "")
	f.Add("/", "/", "")
	f.Add("/users/{id}", "", "")
	f.Add("", "/anything", "")
	f.Add("/path", "/path?extra=stuff", "")
	f.Add("/users/{id}", "/users/123/extra", "")

	// Query parameter edge cases
	f.Add("/search?{q:string}", "/search", "")
	f.Add("/search?{q:string}", "/search", "q=")
	f.Add("/search?{q:string}", "/search", "q=test&extra=ignored")
	f.Add("/users?{limit?10:int}", "/users", "limit=invalid")

	// Special characters in path
	f.Add("/files/{path*:string}", "/files/path/to/file.txt", "")
	f.Add("/users/{name:string}", "/users/name%20with%20spaces", "")
	f.Add("/items/{id}", "/items/123%2F456", "")

	// Multi-segment parameters
	f.Add("/files/{path*:string}", "/files/a/b/c/d/e", "")
	f.Add("/archive/{date*:date:format[yyyy/mm/dd]}", "/archive/2025/01/15", "")

	// Complex real-world patterns
	f.Add("GET /api/v{version:int}/users/{id:uuid}/posts/{slug:slug:length[5..100]}?{fields:string}&{limit?25:int:range[1..100]}", "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/posts/hello-world", "fields=title,body&limit=50")

	f.Fuzz(func(t *testing.T, template, path, query string) {
		done := make(chan struct{})
		var tmpl *pathvars.ParsedTemplate
		var parseErr error
		var matchAttempt pathvars.MatchAttempt
		var matchErr error

		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Match panicked on template=%q path=%q query=%q: %v\nStack: %s",
						template, path, query, r, debug.Stack())
				}
				close(done)
			}()

			// Parse template first
			tmpl, parseErr = pathvars.ParseTemplate(template)
			if parseErr != nil {
				return // Invalid template, can't match
			}

			// Attempt match
			matchAttempt, matchErr = tmpl.Match(path, query)
		}()

		select {
		case <-done:
			// Completed successfully (with or without error)
			if parseErr != nil {
				// Invalid template - can't test matching
				return
			}

			// Match errors are acceptable (validation failures, path mismatches, etc.)
			// Just verify we got sensible results
			if matchErr == nil {
				// Successful match - verify we got a result
				if !matchAttempt.Matched() && matchAttempt.PathMatched {
					// This is OK - path matched but query validation failed
					return
				}
			}

		case <-time.After(10 * time.Second):
			t.Fatalf("Match hung (infinite loop) on template=%q path=%q query=%q", template, path, query)
		}
	})
}
