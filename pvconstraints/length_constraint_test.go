package pvconstraints_test

import (
	"strings"
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvconstraints"
	"github.com/mikeschinkel/go-pathvars/pvtypes"

	_ "github.com/mikeschinkel/go-pathvars/dtclassifiers"
)

var _ pvtypes.Constraint = (*pvconstraints.LengthConstraint)(nil)

func TestLengthConstraintParsing(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		dataType pvtypes.PVDataType
		wantErr  bool
		wantMin  int
		wantMax  int
	}{
		// Valid length specifications
		{"valid-range", "5..50", pvtypes.StringType, false, 5, 50},
		{"exact-length", "3..3", pvtypes.StringType, false, 3, 3},
		{"zero-min", "0..10", pvtypes.StringType, false, 0, 10},
		{"large-range", "100..1000", pvtypes.StringType, false, 100, 1000},
		{"single-digit", "1..9", pvtypes.StringType, false, 1, 9},

		// Invalid length specifications
		{"missing-separator", "5-50", pvtypes.StringType, true, 0, 0},
		{"single-value", "10", pvtypes.StringType, true, 0, 0},
		{"empty-spec", "", pvtypes.StringType, true, 0, 0},
		{"non-numeric-min", "abc..50", pvtypes.StringType, true, 0, 0},
		{"non-numeric-max", "5..xyz", pvtypes.StringType, true, 0, 0},
		{"negative-min", "-5..50", pvtypes.StringType, true, 0, 0},
		{"min-greater-than-max", "50..5", pvtypes.StringType, true, 0, 0},
		{"reversed-range", "10..1", pvtypes.StringType, true, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseLengthConstraint(tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseLengthConstraint() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseLengthConstraint() unexpected error: %v", err)
			}

			if constraint == nil {
				t.Fatal("ParseLengthConstraint() returned nil constraint without error")
			}

			// Verify constraint type
			if constraint.Type() != pvtypes.LengthConstraintType {
				t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.LengthConstraintType)
			}

			// Verify rule format
			rule := constraint.Rule()
			expectedRule := tt.spec
			if rule != expectedRule {
				t.Errorf("Rule() = %q, want %q", rule, expectedRule)
			}
		})
	}
}

func TestLengthConstraintValidation(t *testing.T) {
	tests := []struct {
		name       string
		lengthSpec string
		testValue  string
		wantValid  bool
	}{
		// Valid lengths
		{"min-length", "5..50", "hello", true},
		{"max-length", "5..50", strings.Repeat("a", 50), true},
		{"middle-length", "5..50", "medium-length", true},
		{"exact-match-min", "3..3", "abc", true},

		// Invalid lengths
		{"too-short", "5..50", "hi", false},
		{"too-long", "5..50", strings.Repeat("a", 51), false},
		{"exact-too-long", "3..3", "abcd", false},
		{"exact-too-short", "3..3", "ab", false},
		{"empty-string", "1..10", "", false},
		{"zero-length-excluded", "1..10", "", false},

		// Zero-length handling
		{"zero-allowed-empty", "0..10", "", true},
		{"zero-allowed-value", "0..10", "hello", true},
		{"zero-only-empty", "0..0", "", true},
		{"zero-only-nonempty", "0..0", "x", false},

		// Edge cases
		{"unicode-characters", "5..10", "hello", true},
		{"spaces-count", "3..10", "a b c", true},
		{"single-char-valid", "1..5", "x", true},
		{"single-char-invalid", "2..5", "x", false},

		// Large values
		{"hundred-chars", "50..150", strings.Repeat("x", 100), true},
		{"thousand-chars", "500..1500", strings.Repeat("y", 1000), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseLengthConstraint(tt.lengthSpec)
			if err != nil {
				t.Fatalf("ParseLengthConstraint() failed: %v", err)
			}

			err = constraint.Validate(tt.testValue)

			if tt.wantValid && err != nil {
				t.Errorf("Validate(%q) expected valid but got error: %v", tt.testValue, err)
			}

			if !tt.wantValid && err == nil {
				t.Errorf("Validate(%q) expected invalid but got no error (len=%d)", tt.testValue, len(tt.testValue))
			}
		})
	}
}

func TestLengthConstraintInterface(t *testing.T) {
	constraint, err := pvconstraints.ParseLengthConstraint("5..50")
	if err != nil {
		t.Fatalf("Failed to parse length constraint: %v", err)
	}

	// Test Type()
	if constraint.Type() != pvtypes.LengthConstraintType {
		t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.LengthConstraintType)
	}

	// Test String()
	str := constraint.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	// // t.Logf("String() = %q", str)

	// Test Rule()
	rule := constraint.Rule()
	if rule != "5..50" {
		t.Errorf("Rule() = %q, want %q", rule, "5..50")
	}

	// Test ValidatesType()
	if constraint.ValidatesType() {
		t.Error("ValidatesType() should return false for length constraints")
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

func TestLengthConstraintBoundaryConditions(t *testing.T) {
	tests := []struct {
		name       string
		lengthSpec string
		values     []struct {
			value string
			valid bool
		}
	}{
		{
			name:       "boundary-5-to-50",
			lengthSpec: "5..50",
			values: []struct {
				value string
				valid bool
			}{
				{strings.Repeat("a", 4), false},  // One below min
				{strings.Repeat("a", 5), true},   // Exactly min
				{strings.Repeat("a", 6), true},   // One above min
				{strings.Repeat("a", 49), true},  // One below max
				{strings.Repeat("a", 50), true},  // Exactly max
				{strings.Repeat("a", 51), false}, // One above max
			},
		},
		{
			name:       "boundary-exact-3",
			lengthSpec: "3..3",
			values: []struct {
				value string
				valid bool
			}{
				{"ab", false},
				{"abc", true},
				{"abcd", false},
			},
		},
		{
			name:       "boundary-zero-included",
			lengthSpec: "0..5",
			values: []struct {
				value string
				valid bool
			}{
				{"", true},
				{"a", true},
				{"abcde", true},
				{"abcdef", false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseLengthConstraint(tt.lengthSpec)
			if err != nil {
				t.Fatalf("ParseLengthConstraint() failed: %v", err)
			}

			for _, val := range tt.values {
				err := constraint.Validate(val.value)
				if val.valid && err != nil {
					t.Errorf("Validate(%q) [len=%d] expected valid but got error: %v",
						val.value, len(val.value), err)
				}
				if !val.valid && err == nil {
					t.Errorf("Validate(%q) [len=%d] expected invalid but got no error",
						val.value, len(val.value))
				}
			}
		})
	}
}
