package pvconstraints_test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvconstraints"
	"github.com/mikeschinkel/go-pathvars/pvtypes"

	_ "github.com/mikeschinkel/go-pathvars/dtclassifiers"
)

var _ pvtypes.Constraint = (*pvconstraints.EnumConstraint)(nil)

func TestEnumConstraintParsing(t *testing.T) {
	tests := []struct {
		name     string
		spec     string
		dataType pvtypes.PVDataType
		wantErr  bool
		wantLen  int
	}{
		// Valid enum specifications
		{"simple-enum", "active,inactive,pending", pvtypes.StringType, false, 3},
		{"numeric-enum", "1,2,3,4,5", pvtypes.StringType, false, 5},
		{"single-value", "true", pvtypes.StringType, false, 1},
		{"two-values", "Active,Inactive", pvtypes.StringType, false, 2},
		{"with-spaces", " alpha , beta , gamma ", pvtypes.StringType, false, 3},

		// Invalid enum specifications
		{"empty-spec", "", pvtypes.StringType, true, 0},
		{"empty-value", "active,,pending", pvtypes.StringType, true, 0},
		{"all-empty-values", ",,", pvtypes.StringType, true, 0},
		{"trailing-comma", "active,inactive,", pvtypes.StringType, true, 0},
		{"leading-comma", ",active,inactive", pvtypes.StringType, true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseEnumConstraint(tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseEnumConstraint() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseEnumConstraint() unexpected error: %v", err)
			}

			if constraint == nil {
				t.Fatal("ParseEnumConstraint() returned nil constraint without error")
			}

			// Verify constraint type
			if constraint.Type() != pvtypes.EnumConstraintType {
				t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.EnumConstraintType)
			}

			// Verify the constraint can be used
			rule := constraint.Rule()
			if rule == "" {
				t.Error("Rule() returned empty string")
			}
		})
	}
}

func TestEnumConstraintValidation(t *testing.T) {
	tests := []struct {
		name      string
		enumSpec  string
		testValue string
		wantValid bool
	}{
		// Basic enum validation
		{"valid-first", "active,inactive,pending", "active", true},
		{"valid-middle", "active,inactive,pending", "inactive", true},
		{"valid-last", "active,inactive,pending", "pending", true},
		{"invalid-value", "active,inactive,pending", "unknown", false},

		// Case sensitivity
		{"case-sensitive-lowercase", "Active,Inactive", "active", false},
		{"case-sensitive-match", "Active,Inactive", "Active", true},

		// Numeric enum
		{"numeric-valid", "1,2,3,4,5", "3", true},
		{"numeric-invalid", "1,2,3,4,5", "6", false},
		{"numeric-zero", "0,1,2", "0", true},

		// Single value enum
		{"single-value-match", "true", "true", true},
		{"single-value-mismatch", "true", "false", false},

		// Edge cases
		{"whitespace-handling", "alpha,beta,gamma", "beta", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseEnumConstraint(tt.enumSpec)
			if err != nil {
				t.Fatalf("ParseEnumConstraint() failed: %v", err)
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

func TestEnumConstraintInterface(t *testing.T) {
	constraint, err := pvconstraints.ParseEnumConstraint("active,inactive,pending")
	if err != nil {
		t.Fatalf("Failed to parse enum constraint: %v", err)
	}

	// Test Type()
	if constraint.Type() != pvtypes.EnumConstraintType {
		t.Errorf("Type() = %v, want %v", constraint.Type(), pvtypes.EnumConstraintType)
	}

	// Test String()
	str := constraint.String()
	if str == "" {
		t.Error("String() returned empty string")
	}
	// // t.Logf("String() = %q", str)

	// Test Rule()
	rule := constraint.Rule()
	if rule != "active,inactive,pending" {
		t.Errorf("Rule() = %q, want %q", rule, "active,inactive,pending")
	}

	// Test ValidatesType()
	if constraint.ValidatesType() {
		t.Error("ValidatesType() should return false for enum constraints")
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

func TestEnumConstraintExamples(t *testing.T) {
	tests := []struct {
		name        string
		enumSpec    string
		wantExample string
	}{
		{"first-value-example", "active,inactive,pending", "active"},
		{"numeric-example", "1,2,3", "1"},
		{"single-value-example", "only", "only"},
		{"case-preserved", "Alpha,Beta,Gamma", "Alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := pvconstraints.ParseEnumConstraint(tt.enumSpec)
			if err != nil {
				t.Fatalf("ParseEnumConstraint() failed: %v", err)
			}

			example := constraint.Example(nil)
			if example == nil {
				t.Fatal("Example() returned nil")
			}

			exampleStr, ok := example.(string)
			if !ok {
				t.Fatalf("Example() returned non-string: %T", example)
			}

			if exampleStr != tt.wantExample {
				t.Errorf("Example() = %q, want %q", exampleStr, tt.wantExample)
			}

			// Verify the example validates
			if err := constraint.Validate(exampleStr); err != nil {
				t.Errorf("Example %q failed validation: %v", exampleStr, err)
			}
		})
	}
}

func TestEnumConstraintErrorMessages(t *testing.T) {
	constraint, err := pvconstraints.ParseEnumConstraint("red,green,blue")
	if err != nil {
		t.Fatalf("Failed to parse enum constraint: %v", err)
	}

	// Create a parameter for error detail testing
	param := pvtypes.NewParameter(pvtypes.ParameterArgs{
		NameProps: pvtypes.NameSpecProps{
			Name: "color",
		},
		Location:    pvtypes.PathLocation,
		DataType:    pvtypes.StringType,
		Constraints: []pvtypes.Constraint{constraint},
		Original:    "{color:string:enum[red,green,blue]}",
	})

	// Test error detail
	invalidValue := "yellow"
	detail := constraint.ErrorDetail(&param, invalidValue)

	if detail == "" {
		t.Error("ErrorDetail() returned empty string")
	}

	// Verify error message contains key information
	if !enumContains(detail, "color") {
		t.Errorf("ErrorDetail() should mention parameter name 'color': %s", detail)
	}
	if !enumContains(detail, invalidValue) {
		t.Errorf("ErrorDetail() should mention invalid value: %s", detail)
	}
	if !enumContains(detail, "red") || !enumContains(detail, "green") || !enumContains(detail, "blue") {
		t.Errorf("ErrorDetail() should list allowed values: %s", detail)
	}

	// // t.Logf("Error detail: %s", detail)
}

// Helper function to check if a string contains a substring
func enumContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
