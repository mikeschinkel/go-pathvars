package pvconstraints_test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvconstraints"
	"github.com/mikeschinkel/go-pathvars/pvtypes"

	_ "github.com/mikeschinkel/go-pathvars/dtclassifiers"
)

var _ pvtypes.Constraint = (*pvconstraints.NotEmptyConstraint)(nil)

func TestNotEmptyConstraintParsing(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		// Valid specs (no arguments expected)
		{"empty-spec", "", false},

		// Invalid specs (any arguments are invalid)
		{"with-argument", "notempty", true},
		{"with-value", "true", true},
		{"with-number", "1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseNotEmptyConstraint(tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseNotEmptyConstraint() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseNotEmptyConstraint() unexpected error: %v", err)
			}

			if constraint == nil {
				t.Fatal("ParseNotEmptyConstraint() returned nil constraint without error")
			}

			// Verify constraint type
			if constraint.Type() != pvtypes.NotEmptyConstraintType {
				t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.NotEmptyConstraintType)
			}
		})
	}
}

func TestNotEmptyConstraintValidation(t *testing.T) {
	tests := []struct {
		name      string
		testValue string
		wantValid bool
	}{
		// Valid values (non-empty)
		{"single-char", "a", true},
		{"word", "hello", true},
		{"sentence", "hello world", true},
		{"number", "123", true},
		{"special-chars", "!@#$%", true},
		{"whitespace-only", " ", true}, // Whitespace is not empty
		{"tab", "\t", true},            // Tab is not empty
		{"newline", "\n", true},        // Newline is not empty
		{"mixed-whitespace", " \t\n", true},

		// Invalid values (empty)
		{"empty-string", "", false},
	}

	constraint, err := pvconstraints.ParseNotEmptyConstraint("")
	if err != nil {
		t.Fatalf("Failed to parse not-empty constraint: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := constraint.Validate(tt.testValue)

			if tt.wantValid && err != nil {
				t.Errorf("Validate(%q) expected valid but got error: %v", tt.testValue, err)
			}

			if !tt.wantValid && err == nil {
				t.Errorf("Validate(%q) expected invalid but got no error", tt.testValue)
			}
		})
	}
}

func TestNotEmptyConstraintInterface(t *testing.T) {
	constraint, err := pvconstraints.ParseNotEmptyConstraint("")
	if err != nil {
		t.Fatalf("Failed to parse not-empty constraint: %v", err)
	}

	// Test Type()
	if constraint.Type() != pvtypes.NotEmptyConstraintType {
		t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.NotEmptyConstraintType)
	}

	// Test String()
	str := constraint.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	// // t.Logf("String() = %q", str)

	// Test Rule() - should return empty for notempty (no parameters)
	rule := constraint.Rule()
	if rule != "" {
		t.Errorf("Rule() = %q, want empty string (notempty has no parameters)", rule)
	}

	// Test ValidatesType()
	if constraint.ValidatesType() {
		t.Error("ValidatesType() should return false for notempty constraints")
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

func TestNotEmptyConstraintExamples(t *testing.T) {
	constraint, err := pvconstraints.ParseNotEmptyConstraint("")
	if err != nil {
		t.Fatalf("Failed to parse not-empty constraint: %v", err)
	}

	example := constraint.Example(nil)
	if example == nil {
		t.Fatal("Example() returned nil")
	}

	exampleStr, ok := example.(string)
	if !ok {
		t.Fatalf("Example() returned non-string: %T", example)
	}

	// Verify the example is not empty
	if exampleStr == "" {
		t.Error("Example() returned empty string, but notempty constraint requires non-empty values")
	}

	// Verify the example validates
	if err := constraint.Validate(exampleStr); err != nil {
		t.Errorf("Example %q failed validation: %v", exampleStr, err)
	}

	// // t.Logf("Example value: %q", exampleStr)
}

func TestNotEmptyConstraintErrorMessages(t *testing.T) {
	constraint, err := pvconstraints.ParseNotEmptyConstraint("")
	if err != nil {
		t.Fatalf("Failed to parse not-empty constraint: %v", err)
	}

	// Create a parameter for error message testing
	param := pvtypes.NewParameter(pvtypes.ParameterArgs{
		NameProps: pvtypes.NameSpecProps{
			Name: "username",
		},
		Location:    pvtypes.PathLocation,
		DataType:    pvtypes.StringType,
		Constraints: []pvtypes.Constraint{constraint},
		Original:    "{username:string:notempty}",
	})

	// Test ErrorDetail
	emptyValue := ""
	detail := constraint.ErrorDetail(&param, emptyValue)

	if detail == "" {
		t.Error("ErrorDetail() returned empty string")
	}

	// Verify error message contains key information
	if !notEmptyContains(detail, "username") {
		t.Errorf("ErrorDetail() should mention parameter name 'username': %s", detail)
	}
	if !notEmptyContains(detail, "empty") {
		t.Errorf("ErrorDetail() should mention 'empty': %s", detail)
	}

	// // t.Logf("Error detail: %s", detail)

	// Test ErrorSuggestion
	suggestion := constraint.ErrorSuggestion(&param, emptyValue, "example")

	if suggestion == "" {
		t.Error("ErrorSuggestion() returned empty string")
	}

	if !notEmptyContains(suggestion, "username") {
		t.Errorf("ErrorSuggestion() should mention parameter name: %s", suggestion)
	}

	// // t.Logf("Error suggestion: %s", suggestion)
}

// Helper function to check if a string contains a substring
func notEmptyContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
