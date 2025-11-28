package test

import (
	"reflect"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func TestSimplePathWithoutParameters(t *testing.T) {
	// Test a simple path without any parameters - just the path part, not method+path
	template, err := pathvars.ParseTemplate("/users")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	tests := []struct {
		path        string
		expectMatch bool
		expectErr   bool
	}{
		{path: "/users", expectMatch: true},
		{path: "/users/123", expectErr: true},
		{path: "/posts", expectErr: true},
		{path: "/users/", expectErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.path[1:], func(t *testing.T) {
			// Test exact match - should succeed
			attempt, err := template.Match(tt.path, "")
			if attempt.PathMatched && tt.expectErr {
				t.Errorf("Path error: Expected: %t, Error: %v", tt.expectErr, err)
			}
			if !attempt.Matched() && tt.expectMatch {
				t.Errorf("Expected '%s' to match template '/users', but it didn't", tt.path)
			}
			if attempt.ValuesMap.Len() != 0 {
				t.Errorf("Expected no variables for simple path, got %d: %v", attempt.ValuesMap.Len(), attempt.ValuesMap)
			}
		})
	}
}

func TestSimplePathWithoutParametersVerbose(t *testing.T) {
	// Test a simple path without any parameters
	template, err := pathvars.ParseTemplate("/users")
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}
	if template == nil {
		t.Fatalf("TEMPLATE IS NIL (SHOULD NOT HAPPEN, but Goland flag as potential nil reference")
	}

	//// Show what the regex looks like for debugging
	// t.Logf("Template regex: %v", template.regex)
	//if template.regex != nil {
	//	// t.Logf("Regex pattern: %s", template.regex.String())
	//}

	tests := []struct {
		path        string
		expectMatch bool
		name        string
		vm          *pvtypes.ValuesMap
	}{
		{"/users", true, "exact match", nil},
		{"/posts", false, "different path", nil},
		{"/users/123", false, "path with extra segments", nil},
		{"/user", false, "similar but different path", nil},
		{"/users/", false, "path with trailing slash", nil},
	}

	for _, test := range tests {
		attempt, err := template.Match(test.path, "")

		// t.Logf("Path: %s, Matched: %v, Vars: %v (%s)", test.path, matched, valuesMap, test.name)
		//goland:noinspection GoDfaErrorMayBeNotNil
		if test.expectMatch && !attempt.PathMatched {
			t.Errorf("Path error: Expected Match: %t, Error: %v", test.expectMatch, err)
		}
		//goland:noinspection GoDfaErrorMayBeNotNil
		valuesMap := attempt.ValuesMap
		if test.vm == nil && valuesMap.Len() != 0 {
			t.Errorf("Path %s expectErr: no parameters, got valuesMap=%v=%v", test.path, valuesMap.Keys(), valuesMap.Values())
		} else if test.vm != nil && reflect.DeepEqual(*test.vm, valuesMap) {
			t.Errorf("Path %s expectErr: valuesMap=%v, got valuesMap=%v", test.path, test.vm, valuesMap)
		}

		//goland:noinspection GoDfaErrorMayBeNotNil
		if test.expectMatch && !attempt.Matched() {
			t.Errorf("Path %s expectErr: match=%v, got match=%v", test.path, test.expectMatch, attempt.Matched())
		}
	}
}
