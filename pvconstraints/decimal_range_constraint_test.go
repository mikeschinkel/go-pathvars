package pvconstraints_test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvconstraints"
	"github.com/mikeschinkel/go-pathvars/pvtypes"

	_ "github.com/mikeschinkel/go-pathvars/dtclassifiers"
)

var _ pvtypes.Constraint = (*pvconstraints.DecimalRangeConstraint)(nil)

func TestDecimalRangeConstraintParsing(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		// Valid specs
		{"decimal-range", "0.5..9.5", false},
		{"negative-decimals", "-5.5..5.5", false},
		{"integer-values", "1..10", false}, // Integers are valid decimals
		{"scientific-notation", "1e-3..1e3", false},
		{"large-decimals", "100.5..999.9", false},

		// Invalid specs
		{"missing-separator", "1.5-9.5", true},
		{"reversed-range", "9.5..1.5", true},
		{"non-numeric-min", "abc..9.5", true},
		{"non-numeric-max", "1.5..xyz", true},
		{"empty-spec", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseDecimalRangeConstraint(tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDecimalRangeConstraint() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseDecimalRangeConstraint() unexpected error: %v", err)
			}

			if constraint == nil {
				t.Fatal("ParseDecimalRangeConstraint() returned nil constraint without error")
			}

			// Verify constraint type
			if constraint.Type() != pvtypes.RangeConstraintType {
				t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.RangeConstraintType)
			}
		})
	}
}

func TestDecimalRangeConstraintValidation(t *testing.T) {
	tests := []struct {
		name      string
		rangeSpec string
		testValue string
		wantValid bool
	}{
		// Valid range [0.5..9.5]
		{"min-boundary", "0.5..9.5", "0.5", true},
		{"max-boundary", "0.5..9.5", "9.5", true},
		{"middle-value", "0.5..9.5", "5.0", true},
		{"below-min", "0.5..9.5", "0.4", false},
		{"above-max", "0.5..9.5", "9.6", false},

		// Negative ranges [-5.5..5.5]
		{"negative-valid", "-5.5..5.5", "-2.5", true},
		{"negative-min", "-5.5..5.5", "-5.5", true},
		{"negative-max", "-5.5..5.5", "5.5", true},
		{"negative-below-min", "-5.5..5.5", "-5.6", false},
		{"negative-above-max", "-5.5..5.5", "5.6", false},

		// Integer values in decimal range [1..10]
		{"integer-min", "1..10", "1", true},
		{"integer-max", "1..10", "10", true},
		{"integer-middle", "1..10", "5", true},
		{"integer-below", "1..10", "0", false},
		{"integer-above", "1..10", "11", false},

		// Decimal precision
		{"precise-value", "1.1..1.9", "1.5", true},
		{"many-decimals", "0.1..0.9", "0.55555", true},

		// Non-numeric values
		{"non-numeric", "0..10", "abc", false},
		{"empty-string", "0..10", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseDecimalRangeConstraint(tt.rangeSpec)
			if err != nil {
				t.Fatalf("ParseDecimalRangeConstraint() failed: %v", err)
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

func TestDecimalRangeConstraintInterface(t *testing.T) {
	constraint, err := pvconstraints.ParseDecimalRangeConstraint("0.5..9.5")
	if err != nil {
		t.Fatalf("Failed to parse decimal range constraint: %v", err)
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
	t.Logf("String() = %q", str)

	// Test Rule()
	rule := constraint.Rule()
	if rule == "" {
		t.Error("Rule() returned empty string")
	}
	t.Logf("Rule() = %q", rule)

	// Test ValidatesType()
	if constraint.ValidatesType() {
		t.Error("ValidatesType() should return false for range constraints")
	}

	// Test ValidDataTypes()
	validTypes := constraint.ValidDataTypes()
	if len(validTypes) == 0 {
		t.Error("ValidDataTypes() returned empty slice")
	}

	// Verify decimal/real types are valid
	foundDecimal := false
	foundReal := false
	for _, dt := range validTypes {
		if dt == pvtypes.DecimalType {
			foundDecimal = true
		}
		if dt == pvtypes.RealType {
			foundReal = true
		}
	}
	if !foundDecimal && !foundReal {
		t.Error("ValidDataTypes() should include DecimalType or RealType")
	}
}
