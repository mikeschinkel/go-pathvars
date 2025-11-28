package pvconstraints_test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvconstraints"
	"github.com/mikeschinkel/go-pathvars/pvtypes"

	_ "github.com/mikeschinkel/go-pathvars/dtclassifiers"
)

var _ pvtypes.Constraint = (*pvconstraints.DateFormatConstraint)(nil)

func TestDateFormatConstraintParsing(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
	}{
		// Built-in format aliases
		{"dateonly-alias", "dateonly", false},
		{"utc-alias", "utc", false},
		{"local-alias", "local", false},
		{"datetime-alias", "datetime", false},

		// Custom formats - Date only
		{"yyyy-mm-dd", "yyyy-mm-dd", false},
		{"mm-dd-yyyy", "mm-dd-yyyy", false},
		{"dd-mm-yyyy", "dd-mm-yyyy", false},

		// Custom formats - Time only
		{"hh:mm:ss", "hh:mm:ss", false},
		{"hh:mm", "hh:mm", false},

		// Custom formats - Date + Time
		{"yyyy-mm-dd_hh:mm:ss", "yyyy-mm-dd_hh:mm:ss", false},
		{"yyyy-mm-dd_hh:mm", "yyyy-mm-dd_hh:mm", false},
		{"dd-mm-yyyy_hh:mm:ss", "dd-mm-yyyy_hh:mm:ss", false},
		{"mm-dd-yyyy_hh:mm:ss", "mm-dd-yyyy_hh:mm:ss", false},

		// Ambiguous mm token (should fail - needs ii for minutes when standalone)
		{"mm-only-ambiguous", "mm", true},
		{"mm-ii-disambiguated", "mm_ii", false},

		// Invalid - no tokens
		{"empty-spec", "", true},
		{"no-tokens", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseDateFormatConstraint(tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDateFormatConstraint() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseDateFormatConstraint() unexpected error: %v", err)
			}

			if constraint == nil {
				t.Fatal("ParseDateFormatConstraint() returned nil constraint without error")
			}

			// Verify constraint type
			if constraint.Type() != pvtypes.FormatConstraintType {
				t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.FormatConstraintType)
			}
		})
	}
}

func TestDateFormatConstraintValidation_BuiltinAliases(t *testing.T) {
	tests := []struct {
		name      string
		spec      string
		testValue string
		wantValid bool
	}{
		// dateonly format (yyyy-mm-dd)
		{"dateonly-valid", "dateonly", "2023-12-25", true},
		{"dateonly-invalid-with-time", "dateonly", "2023-12-25T10:30:00", false},
		{"dateonly-invalid-with-z", "dateonly", "2023-12-25T10:30:00Z", false},
		{"dateonly-invalid-format", "dateonly", "12-25-2023", false},

		// utc format (strict UTC with Z required)
		{"utc-valid", "utc", "2023-12-25T10:30:00Z", true},
		{"utc-invalid-missing-z", "utc", "2023-12-25T10:30:00", false},
		{"utc-invalid-date-only", "utc", "2023-12-25", false},

		// local format (timezone-naive, Z forbidden)
		{"local-valid", "local", "2023-12-25T10:30:00", true},
		{"local-invalid-with-z", "local", "2023-12-25T10:30:00Z", false},
		{"local-invalid-date-only", "local", "2023-12-25", false},

		// datetime format (flexible, Z optional)
		{"datetime-valid-with-z", "datetime", "2023-12-25T10:30:00Z", true},
		{"datetime-valid-without-z", "datetime", "2023-12-25T10:30:00", true},
		{"datetime-invalid-date-only", "datetime", "2023-12-25", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseDateFormatConstraint(tt.spec)
			if err != nil {
				t.Fatalf("ParseDateFormatConstraint() failed: %v", err)
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

func TestDateFormatConstraintValidation_CustomFormats(t *testing.T) {
	tests := []struct {
		name      string
		spec      string
		testValue string
		wantValid bool
	}{
		// YYYY-MM-DD format
		{"yyyy-mm-dd-valid", "yyyy-mm-dd", "2023-12-25", true},
		{"yyyy-mm-dd-invalid-format", "yyyy-mm-dd", "12-25-2023", false},

		// MM-DD-YYYY format
		{"mm-dd-yyyy-valid", "mm-dd-yyyy", "12-25-2023", true},
		{"mm-dd-yyyy-invalid-format", "mm-dd-yyyy", "2023-12-25", false},

		// DD-MM-YYYY format
		{"dd-mm-yyyy-valid", "dd-mm-yyyy", "25-12-2023", true},
		{"dd-mm-yyyy-invalid-format", "dd-mm-yyyy", "12-25-2023", false},

		// HH:MM:SS format
		{"hh:mm:ss-valid", "hh:mm:ss", "15:30:00", true},
		{"hh:mm:ss-invalid-hour", "hh:mm:ss", "25:30:00", false},
		{"hh:mm:ss-invalid-minute", "hh:mm:ss", "15:61:00", false},
		{"hh:mm:ss-invalid-second", "hh:mm:ss", "15:30:99", false},

		// YYYY-MM-DD_HH:MM:SS format
		{"yyyy-mm-dd_hh:mm:ss-valid", "yyyy-mm-dd_hh:mm:ss", "2023-12-25_10:30:00", true},
		{"yyyy-mm-dd_hh:mm:ss-invalid-month", "yyyy-mm-dd_hh:mm:ss", "2023-13-25_10:30:00", false},
		{"yyyy-mm-dd_hh:mm:ss-invalid-day", "yyyy-mm-dd_hh:mm:ss", "2023-12-32_10:30:00", false},
		{"yyyy-mm-dd_hh:mm:ss-invalid-hour", "yyyy-mm-dd_hh:mm:ss", "2023-12-25_25:30:00", false},
		{"yyyy-mm-dd_hh:mm:ss-invalid-minute", "yyyy-mm-dd_hh:mm:ss", "2023-12-25_10:61:00", false},
		{"yyyy-mm-dd_hh:mm:ss-invalid-second", "yyyy-mm-dd_hh:mm:ss", "2023-12-25_10:30:61", false},
		{"yyyy-mm-dd_hh:mm:ss-invalid-format", "yyyy-mm-dd_hh:mm:ss", "25-12-2023_10:30:00", false},

		// YYYY-MM-DD_HH:MM format
		{"yyyy-mm-dd_hh:mm-valid", "yyyy-mm-dd_hh:mm", "2023-12-25_10:30", true},
		{"yyyy-mm-dd_hh:mm-invalid-month", "yyyy-mm-dd_hh:mm", "2023-13-25_10:30", false},
		{"yyyy-mm-dd_hh:mm-invalid-hour", "yyyy-mm-dd_hh:mm", "2023-12-25_25:30", false},
		{"yyyy-mm-dd_hh:mm-invalid-minute", "yyyy-mm-dd_hh:mm", "2023-12-25_10:61", false},

		// YYYY-MM-DD_HH format
		{"yyyy-mm-dd_hh-valid", "yyyy-mm-dd_hh", "2023-12-25_10", true},
		{"yyyy-mm-dd_hh-invalid-month", "yyyy-mm-dd_hh", "2023-13-25_10", false},
		{"yyyy-mm-dd_hh-invalid-hour", "yyyy-mm-dd_hh", "2023-12-25_25", false},

		// DD-MM-YYYY_HH:MM:SS format
		{"dd-mm-yyyy_hh:mm:ss-valid", "dd-mm-yyyy_hh:mm:ss", "25-12-2023_10:30:00", true},
		{"dd-mm-yyyy_hh:mm:ss-invalid-day", "dd-mm-yyyy_hh:mm:ss", "32-12-2023_10:30:00", false},
		{"dd-mm-yyyy_hh:mm:ss-invalid-month", "dd-mm-yyyy_hh:mm:ss", "25-13-2023_10:30:00", false},
		{"dd-mm-yyyy_hh:mm:ss-invalid-hour", "dd-mm-yyyy_hh:mm:ss", "25-12-2023_25:30:00", false},
		{"dd-mm-yyyy_hh:mm:ss-invalid-minute", "dd-mm-yyyy_hh:mm:ss", "25-12-2023_10:61:00", false},
		{"dd-mm-yyyy_hh:mm:ss-invalid-second", "dd-mm-yyyy_hh:mm:ss", "25-12-2023_10:30:61", false},
		{"dd-mm-yyyy_hh:mm:ss-invalid-format", "dd-mm-yyyy_hh:mm:ss", "2023-12-25_10:30:00", false},

		// MM-DD-YYYY_HH:MM:SS format
		{"mm-dd-yyyy_hh:mm:ss-valid", "mm-dd-yyyy_hh:mm:ss", "12-25-2023_10:30:00", true},
		{"mm-dd-yyyy_hh:mm:ss-invalid-month", "mm-dd-yyyy_hh:mm:ss", "13-25-2023_10:30:00", false},
		{"mm-dd-yyyy_hh:mm:ss-invalid-day", "mm-dd-yyyy_hh:mm:ss", "12-32-2023_10:30:00", false},
		{"mm-dd-yyyy_hh:mm:ss-invalid-hour", "mm-dd-yyyy_hh:mm:ss", "12-25-2023_25:30:00", false},
		{"mm-dd-yyyy_hh:mm:ss-invalid-minute", "mm-dd-yyyy_hh:mm:ss", "12-25-2023_10:61:00", false},
		{"mm-dd-yyyy_hh:mm:ss-invalid-second", "mm-dd-yyyy_hh:mm:ss", "12-25-2023_10:30:61", false},
		{"mm-dd-yyyy_hh:mm:ss-invalid-format", "mm-dd-yyyy_hh:mm:ss", "2023-12-25_10:30:00", false},

		// MM-II disambiguation (mm=month, ii=minutes)
		{"mm-ii-valid", "mm_ii", "12_30", true},
		{"mm-ii-invalid-month", "mm_ii", "13_30", false},
		{"mm-ii-invalid-minute", "mm_ii", "12_61", false},
		{"mm-ii-invalid-format", "mm_ii", "12:30", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseDateFormatConstraint(tt.spec)
			if err != nil {
				t.Fatalf("ParseDateFormatConstraint() failed: %v", err)
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

func TestDateFormatConstraintInterface(t *testing.T) {
	constraint, err := pvconstraints.ParseDateFormatConstraint("yyyy-mm-dd")
	if err != nil {
		t.Fatalf("Failed to parse date format constraint: %v", err)
	}

	// Test Type()
	if constraint.Type() != pvtypes.FormatConstraintType {
		t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.FormatConstraintType)
	}

	// Test String()
	str := constraint.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	t.Logf("String() = %q", str)

	// Test Rule()
	rule := constraint.Rule()
	if rule != "yyyy-mm-dd" {
		t.Errorf("Rule() = %q, want %q", rule, "yyyy-mm-dd")
	}

	// Test ValidatesType()
	if !constraint.ValidatesType() {
		t.Error("ValidatesType() should return true for format constraints")
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

func TestDateFormatConstraintErrorMessages(t *testing.T) {
	constraint, err := pvconstraints.ParseDateFormatConstraint("yyyy-mm-dd")
	if err != nil {
		t.Fatalf("Failed to parse date format constraint: %v", err)
	}

	// Create a parameter for error message testing
	param := pvtypes.NewParameter(pvtypes.ParameterArgs{
		NameProps: pvtypes.NameSpecProps{
			Name: "event_date",
		},
		Location:    pvtypes.PathLocation,
		DataType:    pvtypes.DateType,
		Constraints: []pvtypes.Constraint{constraint},
		Original:    "{event_date:date:format[yyyy-mm-dd]}",
	})

	// Test ErrorDetail
	invalidValue := "12-25-2023"
	detail := constraint.ErrorDetail(&param, invalidValue)

	if detail == "" {
		t.Error("ErrorDetail() returned empty string")
	}

	t.Logf("Error detail: %s", detail)

	// Test ErrorSuggestion
	suggestion := constraint.ErrorSuggestion(&param, invalidValue, "2023-12-25")

	if suggestion == "" {
		t.Error("ErrorSuggestion() returned empty string")
	}

	t.Logf("Error suggestion: %s", suggestion)
}
