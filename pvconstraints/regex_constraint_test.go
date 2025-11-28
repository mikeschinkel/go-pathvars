package pvconstraints_test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvconstraints"
	"github.com/mikeschinkel/go-pathvars/pvtypes"

	_ "github.com/mikeschinkel/go-pathvars/dtclassifiers"
)

var _ pvtypes.Constraint = (*pvconstraints.RegexConstraint)(nil)

func TestRegexConstraintParsing(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		// Valid patterns
		{"digits-only", "[0-9]+", false},
		{"letters-only", "[a-zA-Z]+", false},
		{"email-like", "[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z]{2,}", false},
		{"version-pattern", "v[0-9]+", false},
		{"alphanumeric", "[a-zA-Z0-9_-]+", false},
		{"word-boundary", "\\w+", false},
		{"optional-group", "(abc)?", false},
		{"alternation", "cat|dog|bird", false},
		{"escaped-chars", "\\[\\]\\(\\)", false},

		// Invalid patterns - empty
		{"empty-pattern", "", true},

		// Invalid patterns - with anchors (rejected by implementation)
		{"with-start-anchor", "^[0-9]+", true},
		{"with-end-anchor", "[0-9]+$", true},
		{"with-both-anchors", "^[0-9]+$", true},

		// Invalid patterns - malformed regex
		{"unclosed-bracket", "[0-9", true},
		{"unclosed-group", "(abc", true},
		{"unclosed-char-class", "[a-z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseRegexConstraint(tt.pattern)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRegexConstraint() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseRegexConstraint() unexpected error: %v", err)
			}

			if constraint == nil {
				t.Fatal("ParseRegexConstraint() returned nil constraint without error")
			}

			// Verify constraint type
			if constraint.Type() != pvtypes.RegexConstraintType {
				t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.RegexConstraintType)
			}

			// Verify rule returns original pattern (without auto-added anchors)
			rule := constraint.Rule()
			if rule != tt.pattern {
				t.Errorf("Rule() = %q, want %q", rule, tt.pattern)
			}
		})
	}
}

func TestRegexConstraintValidation(t *testing.T) {
	tests := []struct {
		name      string
		pattern   string
		testValue string
		wantValid bool
	}{
		// Digit patterns
		{"digits-match", "[0-9]+", "123", true},
		{"digits-nomatch", "[0-9]+", "abc", false},
		{"digits-partial-nomatch", "[0-9]+", "abc123", false}, // Auto-anchored, no partial match
		{"digits-mixed-nomatch", "[0-9]+", "12abc34", false},

		// Letter patterns
		{"letters-match", "[a-zA-Z]+", "Hello", true},
		{"letters-nomatch", "[a-zA-Z]+", "Hello123", false},
		{"letters-uppercase", "[A-Z]+", "HELLO", true},
		{"letters-lowercase", "[a-z]+", "hello", true},
		{"letters-case-mismatch", "[a-z]+", "Hello", false},

		// Complex email-like pattern
		{"email-valid", "[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z]{2,}", "user@example.com", true},
		{"email-invalid-no-at", "[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z]{2,}", "userexample.com", false},
		{"email-invalid-no-dot", "[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z]{2,}", "user@examplecom", false},
		{"email-invalid-short-tld", "[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z]{2,}", "user@example.c", false},
		{"email-invalid-partial", "[a-zA-Z0-9]+@[a-zA-Z0-9]+\\.[a-zA-Z]{2,}", "invalid-email", false},

		// Version patterns
		{"version-match", "v[0-9]+", "v1", true},
		{"version-multi-digit", "v[0-9]+", "v123", true},
		{"version-nomatch-no-v", "v[0-9]+", "1", false},
		{"version-nomatch-uppercase", "v[0-9]+", "V1", false},

		// Alphanumeric with special chars
		{"alphanum-dash-underscore", "[a-zA-Z0-9_-]+", "hello-world_123", true},
		{"alphanum-no-space", "[a-zA-Z0-9_-]+", "hello world", false},
		{"alphanum-no-special", "[a-zA-Z0-9_-]+", "hello@world", false},

		// Optional groups
		{"optional-present", "(abc)?def", "abcdef", true},
		{"optional-absent", "(abc)?def", "def", true},
		{"optional-partial", "(abc)?def", "abc", false},

		// Alternation
		{"alternation-first", "cat|dog|bird", "cat", true},
		{"alternation-middle", "cat|dog|bird", "dog", true},
		{"alternation-last", "cat|dog|bird", "bird", true},
		{"alternation-nomatch", "cat|dog|bird", "fish", false},

		// Edge cases
		{"empty-string-plus", "[a-z]+", "", false}, // + requires at least one
		{"empty-string-star", "[a-z]*", "", true},  // * allows zero
		{"single-char", "[a-z]", "a", true},
		{"single-char-nomatch", "[a-z]", "ab", false}, // Exactly one char required
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseRegexConstraint(tt.pattern)
			if err != nil {
				t.Fatalf("ParseRegexConstraint() failed: %v", err)
			}

			err = constraint.Validate(tt.testValue)

			if tt.wantValid && err != nil {
				t.Errorf("Validate(%q) expected valid but got error: %v", tt.testValue, err)
			}

			if !tt.wantValid && err == nil {
				t.Errorf("Validate(%q) expected invalid but got no error", tt.testValue)
			}
		})
	}
}

func TestRegexConstraintAutoAnchoring(t *testing.T) {
	// Test that patterns are automatically anchored for full string matching
	tests := []struct {
		name      string
		pattern   string
		testValue string
		wantValid bool
		reason    string
	}{
		{
			name:      "no-partial-match-prefix",
			pattern:   "[0-9]+",
			testValue: "123abc",
			wantValid: false,
			reason:    "Pattern should not match if digits are only at start (auto-anchored)",
		},
		{
			name:      "no-partial-match-suffix",
			pattern:   "[0-9]+",
			testValue: "abc123",
			wantValid: false,
			reason:    "Pattern should not match if digits are only at end (auto-anchored)",
		},
		{
			name:      "no-partial-match-middle",
			pattern:   "[0-9]+",
			testValue: "abc123def",
			wantValid: false,
			reason:    "Pattern should not match if digits are in middle (auto-anchored)",
		},
		{
			name:      "full-string-match",
			pattern:   "[0-9]+",
			testValue: "123",
			wantValid: true,
			reason:    "Pattern should match when entire string is digits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseRegexConstraint(tt.pattern)
			if err != nil {
				t.Fatalf("ParseRegexConstraint() failed: %v", err)
			}

			err = constraint.Validate(tt.testValue)
			valid := err == nil

			if valid != tt.wantValid {
				t.Errorf("Validate(%q) = %v, want %v: %s", tt.testValue, valid, tt.wantValid, tt.reason)
			}
		})
	}
}

func TestRegexConstraintInterface(t *testing.T) {
	constraint, err := pvconstraints.ParseRegexConstraint("[0-9]+")
	if err != nil {
		t.Fatalf("Failed to parse regex constraint: %v", err)
	}

	// Test Type()
	if constraint.Type() != pvtypes.RegexConstraintType {
		t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.RegexConstraintType)
	}

	// Test String()
	str := constraint.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	// // t.Logf("String() = %q", str)

	// Test Rule() - should return original pattern without anchors
	rule := constraint.Rule()
	if rule != "[0-9]+" {
		t.Errorf("Rule() = %q, want %q", rule, "[0-9]+")
	}

	// Test ValidatesType()
	if constraint.ValidatesType() {
		t.Error("ValidatesType() should return false for regex constraints")
	}

	// Test ValidDataTypes()
	validTypes := constraint.ValidDataTypes()
	if len(validTypes) == 0 {
		t.Error("ValidDataTypes() returned empty slice")
	}

	// Verify string type is valid
	found := false
	for _, dt := range validTypes {
		if dt == pvtypes.StringType {
			found = true
			break
		}
	}
	if !found {
		t.Error("ValidDataTypes() should include StringType")
	}
}

func TestRegexConstraintErrorMessages(t *testing.T) {
	constraint, err := pvconstraints.ParseRegexConstraint("[0-9]+")
	if err != nil {
		t.Fatalf("Failed to parse regex constraint: %v", err)
	}

	// Create a parameter for error message testing
	param := pvtypes.NewParameter(pvtypes.ParameterArgs{
		NameProps: pvtypes.NameSpecProps{
			Name: "code",
		},
		Location:    pvtypes.PathLocation,
		DataType:    pvtypes.StringType,
		Constraints: []pvtypes.Constraint{constraint},
		Original:    "{code:string:regex[[0-9]+]}",
	})

	// Test ErrorSuggestion - should mention anchor restriction
	invalidValue := "abc123"
	suggestion := constraint.ErrorSuggestion(&param, invalidValue, "123")

	if suggestion == "" {
		t.Error("ErrorSuggestion() returned empty string")
	}

	// Verify suggestion mentions anchors (implementation detail)
	//if !regexContains(suggestion, "anchor") && !regexContains(suggestion, "^") && !regexContains(suggestion, "$") {
	// t.Logf("Note: ErrorSuggestion might want to mention anchors: %s", suggestion)
	//}

	// t.Logf("Error suggestion: %s", suggestion)
}

func TestRegexConstraintAnchorRejection(t *testing.T) {
	// Test that patterns with anchors are explicitly rejected
	tests := []struct {
		name    string
		pattern string
		wantErr string // Expected error content
	}{
		{
			name:    "start-anchor-rejected",
			pattern: "^[0-9]+",
			wantErr: "anchor",
		},
		{
			name:    "end-anchor-rejected",
			pattern: "[0-9]+$",
			wantErr: "anchor",
		},
		{
			name:    "both-anchors-rejected",
			pattern: "^[0-9]+$",
			wantErr: "anchor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pvconstraints.ParseRegexConstraint(tt.pattern)

			if err == nil {
				t.Errorf("ParseRegexConstraint() expected error for pattern with anchors but got none")
				return
			}

			errMsg := err.Error()
			if !regexContains(errMsg, tt.wantErr) {
				t.Errorf("Error message should contain %q but got: %s", tt.wantErr, errMsg)
			}

			// // t.Logf("Expected error: %v", err)
		})
	}
}

// Helper function to check if a string contains a substring
func regexContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
