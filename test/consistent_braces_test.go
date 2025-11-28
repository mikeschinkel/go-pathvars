package test

import (
	"errors"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func TestConsistentBraceHandling(t *testing.T) {
	tests := []struct {
		name    string
		method  pathvars.HTTPMethod
		path    pathvars.Template
		wantErr bool
		reason  string
	}{
		{"valid-param", "GET", "/users/{id}", false, "Valid parameter"},
		{"empty-braces", "GET", "/users/{}", true, "Empty parameter - error"},
		{"unmatched-open", "GET", "/users/{id", true, "Unmatched opening brace - error"},
		{"unmatched-close", "GET", "/users/id}", true, "Unmatched closing brace - error (now consistent!)"},
		{"valid-literal", "GET", "/users/myid", false, "No braces at all - valid literal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			err := router.AddRoute(tt.method, tt.path, nil)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %s but got none", tt.reason)
					return
				}
				if !errors.Is(err, pvtypes.ErrInvalidParameter) && !errors.Is(err, pathvars.ErrInvalidTemplate) {
					t.Errorf("Expected InvalidParameter or InvalidTemplate error, got: %v", err)
					return
				}
				// t.Logf("✅ Correctly rejected: %s", tt.reason)
				return
			}

			if err != nil {
				t.Errorf("Expected no error for %s but got: %v", tt.reason, err)
				return
			}
			// t.Logf("✅ Correctly accepted: %s", tt.reason)
		})
	}
}
