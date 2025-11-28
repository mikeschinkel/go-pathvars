package pvconstraints_test

import (
	"fmt"
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvconstraints"
	"github.com/mikeschinkel/go-pathvars/pvtypes"

	_ "github.com/mikeschinkel/go-pathvars/dtclassifiers"
)

var _ pvtypes.Constraint = (*pvconstraints.DecimalRangeConstraint)(nil)

func TestParseRangeConstraint(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		dataType pvtypes.PVDataType
		wantErr  bool
		wantType string
	}{
		// Integer range constraints
		{"int-valid-range", "1..10", pvtypes.IntegerType, false, "*pvconstraints.IntegerRangeConstraint"},
		{"int-negative-range", "-10..10", pvtypes.IntegerType, false, "*pvconstraints.IntegerRangeConstraint"},
		{"int-large-range", "1000..9999", pvtypes.IntegerType, false, "*pvconstraints.IntegerRangeConstraint"},
		{"int-invalid-format", "1-10", pvtypes.IntegerType, true, ""},
		{"int-invalid-order", "10..1", pvtypes.IntegerType, true, ""},
		{"int-non-numeric", "abc..def", pvtypes.IntegerType, true, ""},

		// Decimal/Real range constraints
		{"decimal-valid-range", "0.5..9.5", pvtypes.DecimalType, false, "*pvconstraints.DecimalRangeConstraint"},
		{"decimal-negative-range", "-5.5..5.5", pvtypes.DecimalType, false, "*pvconstraints.DecimalRangeConstraint"},
		{"decimal-integer-values", "1..10", pvtypes.DecimalType, false, "*pvconstraints.DecimalRangeConstraint"},
		{"real-valid-range", "0.1..99.9", pvtypes.RealType, false, "*pvconstraints.DecimalRangeConstraint"},
		{"real-scientific", "1e-3..1e3", pvtypes.RealType, false, "*pvconstraints.DecimalRangeConstraint"},
		{"decimal-invalid-format", "1.5-9.5", pvtypes.DecimalType, true, ""},
		{"decimal-invalid-order", "9.5..1.5", pvtypes.DecimalType, true, ""},

		// Date range constraints
		{"date-valid-range", "2023-01-01..2023-12-31", pvtypes.DateType, false, "*pvconstraints.DateRangeConstraint"},
		{"date-different-years", "2020-01-01..2025-12-31", pvtypes.DateType, false, "*pvconstraints.DateRangeConstraint"},
		{"date-invalid-format", "2023/01/01..2023/12/31", pvtypes.DateType, true, ""},
		{"date-invalid-order", "2023-12-31..2023-01-01", pvtypes.DateType, true, ""},
		{"date-malformed", "not-a-date..also-not-a-date", pvtypes.DateType, true, ""},

		// Unsupported data types
		{"string-unsupported", "a..z", pvtypes.StringType, true, ""},
		{"boolean-unsupported", "true..false", pvtypes.BooleanType, true, ""},
		{"uuid-unsupported", "uuid1..uuid2", pvtypes.UUIDType, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseRangeConstraint(tt.spec, tt.dataType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRangeConstraint(%q, %v) expected error, got nil", tt.spec, tt.dataType)
				}
				if constraint != nil {
					t.Errorf("ParseRangeConstraint(%q, %v) expected nil constraint on error, got %T", tt.spec, tt.dataType, constraint)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRangeConstraint(%q, %v) unexpected error: %v", tt.spec, tt.dataType, err)
				return
			}

			if constraint == nil {
				t.Errorf("ParseRangeConstraint(%q, %v) returned nil constraint", tt.spec, tt.dataType)
				return
			}

			// Check constraint type
			constraintType := fmt.Sprintf("%T", constraint)
			if constraintType != tt.wantType {
				t.Errorf("ParseRangeConstraint(%q, %v) returned %s, wantSpec %s", tt.spec, tt.dataType, constraintType, tt.wantType)
			}

			// Check that constraint type is correct
			if constraint.Type() != pvtypes.RangeConstraintType {
				t.Errorf("ParseRangeConstraint(%q, %v) constraint.Type() = %v, wantSpec %v", tt.spec, tt.dataType, constraint.Type(), pvtypes.RangeConstraintType)
			}
		})
	}
}

func TestParseRangeConstraintValidation(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		dataType pvtypes.PVDataType
		testVal  string
		wantErr  bool
	}{
		// Test that parsed constraints actually validate correctly
		{"int-range-valid-min", "10..20", pvtypes.IntegerType, "10", false},
		{"int-range-valid-max", "10..20", pvtypes.IntegerType, "20", false},
		{"int-range-valid-middle", "10..20", pvtypes.IntegerType, "15", false},
		{"int-range-invalid-low", "10..20", pvtypes.IntegerType, "9", true},
		{"int-range-invalid-high", "10..20", pvtypes.IntegerType, "21", true},

		{"decimal-range-valid", "1.5..2.5", pvtypes.DecimalType, "2.0", false},
		{"decimal-range-invalid", "1.5..2.5", pvtypes.DecimalType, "3.0", true},

		{"real-range-valid", "0.1..0.9", pvtypes.RealType, "0.5", false},
		{"real-range-invalid", "0.1..0.9", pvtypes.RealType, "1.0", true},

		{"date-range-valid", "2023-01-01..2023-12-31", pvtypes.DateType, "2023-06-15", false},
		{"date-range-invalid", "2023-01-01..2023-12-31", pvtypes.DateType, "2024-01-01", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseRangeConstraint(tt.spec, tt.dataType)
			if err != nil {
				t.Fatalf("ParseRangeConstraint(%q, %v) failed: %v", tt.spec, tt.dataType, err)
			}

			err = constraint.Validate(tt.testVal)
			if tt.wantErr && err == nil {
				t.Errorf("constraint.Validate(%q) expected error, got nil", tt.testVal)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("constraint.Validate(%q) unexpected error: %v", tt.testVal, err)
			}
		})
	}
}
