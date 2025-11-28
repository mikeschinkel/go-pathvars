package pvconstraints_test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func TestUUIDConstraintRegistry(t *testing.T) {
	// Test that UUID format constraints are properly registered with the new DataType+ConstraintType registry

	t.Run("uuid-format-constraint-registered-for-uuid-type", func(t *testing.T) {
		constraint, err := pvtypes.GetConstraint(pvtypes.FormatConstraintType, pvtypes.UUIDType)
		if err != nil {
			t.Fatalf("Failed to get UUID format constraint: %v", err)
		}
		if constraint == nil {
			t.Fatal("UUID format constraint is nil")
		}
		//t.Logf("UUID format constraint found: %T", constraint)
	})

	t.Run("string-format-constraint-registered-for-string-type", func(t *testing.T) {
		constraint, err := pvtypes.GetConstraint(pvtypes.FormatConstraintType, pvtypes.StringType)
		if err != nil {
			t.Fatalf("Failed to get string format constraint: %v", err)
		}
		if constraint == nil {
			t.Fatal("String format constraint is nil")
		}
		//t.Logf("String format constraint found: %T", constraint)
	})

	t.Run("date-format-constraint-registered-for-date-type", func(t *testing.T) {
		constraint, err := pvtypes.GetConstraint(pvtypes.FormatConstraintType, pvtypes.DateType)
		if err != nil {
			t.Fatalf("Failed to get date format constraint: %v", err)
		}
		if constraint == nil {
			t.Fatal("Date format constraint is nil")
		}
		//t.Logf("Date format constraint found: %T", constraint)
	})

	t.Run("parse-uuid-v4-constraint", func(t *testing.T) {
		constraint, err := pvtypes.GetConstraint(pvtypes.FormatConstraintType, pvtypes.UUIDType)
		if err != nil {
			t.Fatalf("Failed to get UUID format constraint: %v", err)
		}

		parsedConstraint, err := constraint.Parse("v4", pvtypes.UUIDType)
		if err != nil {
			t.Fatalf("Failed to parse UUID v4 constraint: %v", err)
		}
		if parsedConstraint == nil {
			t.Fatal("Parsed UUID v4 constraint is nil")
		}
		//t.Logf("UUID v4 constraint parsed: %T", parsedConstraint)
	})

	t.Run("parse-ulid-constraint-via-string-type", func(t *testing.T) {
		constraint, err := pvtypes.GetConstraint(pvtypes.FormatConstraintType, pvtypes.StringType)
		if err != nil {
			t.Fatalf("Failed to get string format constraint: %v", err)
		}

		parsedConstraint, err := constraint.Parse("ulid", pvtypes.StringType)
		if err != nil {
			t.Fatalf("Failed to parse ULID constraint: %v", err)
		}
		if parsedConstraint == nil {
			t.Fatal("Parsed ULID constraint is nil")
		}
		//t.Logf("ULID constraint parsed: %T", parsedConstraint)
	})

	t.Run("constraint-map-key-generation", func(t *testing.T) {
		// Test the new constraint map key generation
		key := pvtypes.GetConstraintMapKey(pvtypes.FormatConstraintType, pvtypes.UUIDTypeSlug)
		var expectedKey pvtypes.ConstraintMapKey = "uuid_format"
		if key != expectedKey {
			t.Errorf("Expected constraint map key %q, got %q", expectedKey, key)
		}
		//t.Logf("UUID format constraint map key: %s", key)

		key = pvtypes.GetConstraintMapKey(pvtypes.FormatConstraintType, pvtypes.StringTypeSlug)
		expectedKey = "string_format"
		if key != expectedKey {
			t.Errorf("Expected constraint map key %q, got %q", expectedKey, key)
		}
		//t.Logf("String format constraint map key: %s", key)
	})
}
