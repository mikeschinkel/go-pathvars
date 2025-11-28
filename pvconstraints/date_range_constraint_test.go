package pvconstraints_test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvconstraints"
	"github.com/mikeschinkel/go-pathvars/pvtypes"

	_ "github.com/mikeschinkel/go-pathvars/dtclassifiers"
)

var _ pvtypes.Constraint = (*pvconstraints.DateRangeConstraint)(nil)

func TestDateRangeConstraintParsing(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		// Valid specs
		{"same-year-range", "2023-01-01..2023-12-31", false},
		{"multi-year-range", "2020-01-01..2025-12-31", false},
		{"single-day", "2023-06-15..2023-06-15", false},
		{"leap-year", "2024-02-29..2024-03-01", false},

		// Invalid specs
		{"wrong-separator", "2023/01/01..2023/12/31", true},
		{"reversed-dates", "2023-12-31..2023-01-01", true},
		{"invalid-date-format", "01-01-2023..12-31-2023", true},
		{"non-date-min", "not-a-date..2023-12-31", true},
		{"non-date-max", "2023-01-01..also-not-a-date", true},
		{"empty-spec", "", true},
		{"invalid-month", "2023-13-01..2023-12-31", true},
		{"invalid-day", "2023-12-32..2023-12-31", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseDateRangeConstraint(tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDateRangeConstraint() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseDateRangeConstraint() unexpected error: %v", err)
			}

			if constraint == nil {
				t.Fatal("ParseDateRangeConstraint() returned nil constraint without error")
			}

			// Verify constraint type
			if constraint.Type() != pvtypes.RangeConstraintType {
				t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.RangeConstraintType)
			}
		})
	}
}

func TestDateRangeConstraintValidation(t *testing.T) {
	tests := []struct {
		name      string
		rangeSpec string
		testValue string
		wantValid bool
	}{
		// Valid range [2023-01-01..2023-12-31]
		{"min-boundary", "2023-01-01..2023-12-31", "2023-01-01", true},
		{"max-boundary", "2023-01-01..2023-12-31", "2023-12-31", true},
		{"middle-date", "2023-01-01..2023-12-31", "2023-06-15", true},
		{"before-min", "2023-01-01..2023-12-31", "2022-12-31", false},
		{"after-max", "2023-01-01..2023-12-31", "2024-01-01", false},

		// Multi-year range [2020-01-01..2025-12-31]
		{"multi-year-start", "2020-01-01..2025-12-31", "2020-01-01", true},
		{"multi-year-end", "2020-01-01..2025-12-31", "2025-12-31", true},
		{"multi-year-middle", "2020-01-01..2025-12-31", "2023-06-15", true},
		{"multi-year-before", "2020-01-01..2025-12-31", "2019-12-31", false},
		{"multi-year-after", "2020-01-01..2025-12-31", "2026-01-01", false},

		// Leap year handling
		{"leap-year-valid", "2024-02-28..2024-03-01", "2024-02-29", true},
		{"non-leap-invalid", "2023-02-28..2023-03-01", "2023-02-29", false},

		// Edge cases
		{"same-day-range", "2023-06-15..2023-06-15", "2023-06-15", true},
		{"same-day-before", "2023-06-15..2023-06-15", "2023-06-14", false},
		{"same-day-after", "2023-06-15..2023-06-15", "2023-06-16", false},

		// Invalid date formats
		{"wrong-format", "2023-01-01..2023-12-31", "01/15/2023", false},
		{"non-date", "2023-01-01..2023-12-31", "not-a-date", false},
		{"empty-string", "2023-01-01..2023-12-31", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseDateRangeConstraint(tt.rangeSpec)
			if err != nil {
				t.Fatalf("ParseDateRangeConstraint() failed: %v", err)
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

func TestDateRangeConstraintInterface(t *testing.T) {
	constraint, err := pvconstraints.ParseDateRangeConstraint("2023-01-01..2023-12-31")
	if err != nil {
		t.Fatalf("Failed to parse date range constraint: %v", err)
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
	if rule != "2023-01-01..2023-12-31" {
		t.Errorf("Rule() = %q, want %q", rule, "2023-01-01..2023-12-31")
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

	// Verify date type is valid
	found := false
	for _, dt := range validTypes {
		if dt == pvtypes.DateType {
			found = true
			break
		}
	}
	if !found {
		t.Error("ValidDataTypes() should include DateType")
	}
}
