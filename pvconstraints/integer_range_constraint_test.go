package pvconstraints_test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvconstraints"
	"github.com/mikeschinkel/go-pathvars/pvtypes"

	_ "github.com/mikeschinkel/go-pathvars/dtclassifiers"
)

var _ pvtypes.Constraint = (*pvconstraints.IntegerRangeConstraint)(nil)

func TestIntegerRangeConstraintParsing(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		// Valid specs
		{"positive-range", "0..100", false},
		{"negative-range", "-20..50", false},
		{"large-range", "1000..9999", false},
		{"single-value-range", "5..5", false},
		{"negative-to-negative", "-100..-10", false},

		// Invalid specs
		{"missing-separator", "0-100", true},
		{"reversed-range", "100..0", true},
		{"non-numeric-min", "abc..100", true},
		{"non-numeric-max", "0..xyz", true},
		{"empty-spec", "", true},
		{"single-value", "100", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseIntRangeConstraint(tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseIntRangeConstraint() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseIntRangeConstraint() unexpected error: %v", err)
			}

			if constraint == nil {
				t.Fatal("ParseIntRangeConstraint() returned nil constraint without error")
			}

			// Verify constraint type
			if constraint.Type() != pvtypes.RangeConstraintType {
				t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.RangeConstraintType)
			}
		})
	}
}

func TestIntegerRangeConstraintValidation(t *testing.T) {
	tests := []struct {
		name      string
		rangeSpec string
		testValue string
		wantValid bool
	}{
		// Valid range [0..100]
		{"min-boundary", "0..100", "0", true},
		{"max-boundary", "0..100", "100", true},
		{"middle-value", "0..100", "50", true},
		{"below-min", "0..100", "-1", false},
		{"above-max", "0..100", "101", false},

		// Negative ranges [-20..50]
		{"negative-valid", "-20..50", "-10", true},
		{"negative-min", "-20..50", "-20", true},
		{"negative-max", "-20..50", "50", true},
		{"negative-below-min", "-20..50", "-21", false},
		{"negative-above-max", "-20..50", "51", false},

		// Large numbers [1000..9999]
		{"large-valid", "1000..9999", "5000", true},
		{"large-min", "1000..9999", "1000", true},
		{"large-max", "1000..9999", "9999", true},
		{"large-below", "1000..9999", "999", false},
		{"large-above", "1000..9999", "10000", false},

		// Edge cases
		{"exact-range", "5..5", "5", true},
		{"exact-range-invalid", "5..5", "4", false},

		// Non-integer values
		{"non-integer-decimal", "0..100", "50.5", false},
		{"non-integer-text", "0..100", "abc", false},
		{"empty-string", "0..100", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseIntRangeConstraint(tt.rangeSpec)
			if err != nil {
				t.Fatalf("ParseIntRangeConstraint() failed: %v", err)
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

func TestIntegerRangeConstraintInterface(t *testing.T) {
	constraint, err := pvconstraints.ParseIntRangeConstraint("0..100")
	if err != nil {
		t.Fatalf("Failed to parse integer range constraint: %v", err)
	}

	// Test Type()
	if constraint.Type() != pvtypes.RangeConstraintType {
		t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.RangeConstraintType)
	}

	// Test String()
	str := constraint.String()
	if str == "" {
		t.Error("String() returned empty string")
	}

	// Test Rule()
	rule := constraint.Rule()
	if rule != "0..100" {
		t.Errorf("Rule() = %q, want %q", rule, "0..100")
	}

	// Test ValidatesType()
	if constraint.ValidatesType() {
		t.Error("ValidatesType() should return false for range constraints")
	}

	// Test ValidDataTypes()
	validTypes := constraint.ValidDataTypes()
	if len(validTypes) == 0 {
		t.Error("ValidDataTypes() returned empty slice")
	}

	// Verify integer type is valid
	found := false
	for _, dt := range validTypes {
		if dt == pvtypes.IntegerType {
			found = true
			break
		}
	}
	if !found {
		t.Error("ValidDataTypes() should include IntegerType")
	}
}
