package pvconstraints

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func TestUUIDFormatConstraintParsing(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		expectError bool
		description string
	}{
		// Valid format specifications
		{"v1-format", "v1", false, "UUID version 1"},
		{"v2-format", "v2", false, "UUID version 2"},
		{"v3-format", "v3", false, "UUID version 3"},
		{"v4-format", "v4", false, "UUID version 4"},
		{"v5-format", "v5", false, "UUID version 5"},
		{"v6-format", "v6", false, "UUID version 6"},
		{"v7-format", "v7", false, "UUID version 7"},
		{"v8-format", "v8", false, "UUID version 8"},
		{"v1-5-range", "v1-5", false, "UUID versions 1-5"},
		{"v1to5-range", "v1to5", false, "UUID versions 1-5 alternative syntax"},
		{"v6-8-range", "v6-8", false, "UUID versions 6-8"},
		{"v6to8-range", "v6to8", false, "UUID versions 6-8 alternative syntax"},
		{"any-format", "any", false, "Any UUID version"},
		{"generic-format", "generic", false, "Generic UUID format"},
		{"ulid-format", "ulid", false, "ULID format"},
		{"ksuid-format", "ksuid", false, "KSUID format"},
		{"nanoid-format", "nanoid", false, "NanoID format"},

		// Case insensitive
		{"v4-uppercase", "V4", false, "Uppercase V4"},
		{"ulid-mixed-case", "ULID", false, "Uppercase ULID"},

		// Invalid format specifications
		{"unknown-format", "v9", true, "Unsupported version"},
		{"invalid-range", "v1-3", true, "Invalid range"},
		{"empty-spec", "", true, "Empty specification"},
		{"random-text", "random", true, "Random text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseUUIDFormatConstraint(tt.spec)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for spec %q, but got none", tt.spec)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for spec %q: %v", tt.spec, err)
				return
			}

			if constraint == nil {
				t.Errorf("Expected constraint but got nil for spec %q", tt.spec)
				return
			}

			// Verify constraint properties
			if constraint.Type() != pvtypes.FormatConstraintType {
				t.Errorf("Expected FormatConstraintType, got %v", constraint.Type())
			}

			validTypes := constraint.ValidDataTypes()
			if len(validTypes) != 1 || validTypes[0] != pvtypes.UUIDType {
				t.Errorf("Expected ValidDataTypes to return [UUIDType], got %v", validTypes)
			}
		})
	}
}

func TestUUIDFormatValidation(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name        string
		format      string
		value       string
		expectValid bool
		description string
	}{
		// Standard UUID v4 tests
		{"v4-valid", "v4", "550e8400-e29b-41d4-a716-446655440000", true, "Valid v4 UUID"},
		{"v4-valid-2", "v4", "6ba7b810-9dad-41d1-80b4-00c04fd430c8", true, "Another valid v4 UUID"},
		{"v4-wrong-version", "v4", "550e8400-e29b-11d4-a716-446655440000", false, "v1 UUID when expecting v4"},
		{"v4-invalid-format", "v4", "not-a-uuid", false, "Invalid UUID format"},
		{"v4-missing-hyphens", "v4", "550e8400e29b41d4a716446655440000", false, "UUID without hyphens"},
		{"v4-wrong-length", "v4", "550e8400-e29b-41d4-a716-44665544000", false, "UUID with wrong length"},

		// Standard UUID v1 tests
		{"v1-valid", "v1", "550e8400-e29b-11d1-a716-446655440000", true, "Valid v1 UUID"},
		{"v1-wrong-version", "v1", "550e8400-e29b-41d4-a716-446655440000", false, "v4 UUID when expecting v1"},

		// Standard UUID v7 tests (modern)
		{"v7-valid", "v7", "01890a5d-ac96-774b-b900-4aed2fc33a80", true, "Valid v7 UUID"},
		{"v7-wrong-version", "v7", "550e8400-e29b-41d4-a716-446655440000", false, "v4 UUID when expecting v7"},

		// Range validation tests
		{"v1-5-v1", "v1-5", "550e8400-e29b-11d1-a716-446655440000", true, "v1 UUID in v1-5 range"},
		{"v1-5-v4", "v1-5", "550e8400-e29b-41d4-a716-446655440000", true, "v4 UUID in v1-5 range"},
		{"v1-5-v5", "v1-5", "550e8400-e29b-51d5-a716-446655440000", true, "v5 UUID in v1-5 range"},
		{"v1-5-v7", "v1-5", "01890a5d-ac96-774b-b900-4aed2fc33a80", false, "v7 UUID not in v1-5 range"},

		{"v6-8-v6", "v6-8", "1e890a5d-ac96-674b-b900-4aed2fc33a80", true, "v6 UUID in v6-8 range"},
		{"v6-8-v7", "v6-8", "01890a5d-ac96-774b-b900-4aed2fc33a80", true, "v7 UUID in v6-8 range"},
		{"v6-8-v4", "v6-8", "550e8400-e29b-41d4-a716-446655440000", false, "v4 UUID not in v6-8 range"},

		// Generic UUID tests
		{"any-v1", "any", "550e8400-e29b-11d1-a716-446655440000", true, "v1 UUID for any format"},
		{"any-v4", "any", "550e8400-e29b-41d4-a716-446655440000", true, "v4 UUID for any format"},
		{"any-v7", "any", "01890a5d-ac96-774b-b900-4aed2fc33a80", true, "v7 UUID for any format"},
		{"any-invalid", "any", "not-a-uuid", false, "Invalid UUID for any format"},

		// ULID tests
		{"ulid-valid", "ulid", "01ARZ3NDEKTSV4RRFFQ69G5FAV", true, "Valid ULID"},
		{"ulid-valid-2", "ulid", "01F5B3Q9A9XZQZQZQZQZQZQZQZ", true, "Another valid ULID"},
		{"ulid-invalid-length", "ulid", "01ARZ3NDEKTSV4RRFFQ69G5FA", false, "ULID too short"},
		{"ulid-invalid-chars", "ulid", "01ARZ3NDEKTSV4RRFFQ69G5FaV", false, "ULID with lowercase"},
		{"ulid-uuid", "ulid", "550e8400-e29b-41d4-a716-446655440000", false, "UUID when expecting ULID"},

		// KSUID tests
		{"ksuid-valid", "ksuid", "1srOrx2ZWZBpBUvZwXKQmoEYga2", true, "Valid KSUID"},
		{"ksuid-valid-2", "ksuid", "1ZH4T4V8Q5ZHWC8YG93N6QLY2KT", true, "Another valid KSUID"},
		{"ksuid-invalid-length", "ksuid", "1srOrx2ZWZBpBUvZwXKQmoEYga", false, "KSUID too short"},
		{"ksuid-invalid-chars", "ksuid", "1srOrx2ZWZBpBUvZwXKQmoEYga!", false, "KSUID with invalid chars"},
		{"ksuid-uuid", "ksuid", "550e8400-e29b-41d4-a716-446655440000", false, "UUID when expecting KSUID"},

		// NanoID tests
		{"nanoid-valid", "nanoid", "V1StGXR8_Z5jdHi6B-myT", true, "Valid NanoID"},
		{"nanoid-valid-2", "nanoid", "FyiCLb5xyh6Z4_e3VSkwT", true, "Another valid NanoID"},
		{"nanoid-invalid-length", "nanoid", "V1StGXR8_Z5jdHi6B-myT1", false, "NanoID too long"},
		{"nanoid-invalid-chars", "nanoid", "V1StGXR8@Z5jdHi6B-myT", false, "NanoID with invalid chars"},
		{"nanoid-uuid", "nanoid", "550e8400-e29b-41d4-a716-446655440000", false, "UUID when expecting NanoID"},

		// Edge cases
		{"empty-value", "v4", "", false, "Empty value"},
		{"whitespace", "v4", "   ", false, "Whitespace value"},
		{"almost-uuid", "v4", "550e8400-e29b-41d4-a716-44665544000g", false, "Almost valid UUID"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			constraint, err := ParseUUIDFormatConstraint(tt.format)
			if err != nil {
				t.Fatalf("Failed to parse format %q: %v", tt.format, err)
			}

			err = constraint.Validate(tt.value)

			if tt.expectValid && err != nil {
				t.Errorf("Expected value %q to be valid for format %q, but got error: %v",
					tt.value, tt.format, err)
			} else if !tt.expectValid && err == nil {
				t.Errorf("Expected value %q to be invalid for format %q, but validation passed",
					tt.value, tt.format)
			}
		})
	}
}

func TestUUIDFormatConstraintInterface(t *testing.T) {
	constraint, err := ParseUUIDFormatConstraint("v4")
	if err != nil {
		t.Fatalf("Failed to create constraint: %v", err)
	}

	// Test Type method
	if constraint.Type() != pvtypes.FormatConstraintType {
		t.Errorf("Expected Type() to return FormatConstraintType, got %v", constraint.Type())
	}

	// Test ValidDataTypes method
	validTypes := constraint.ValidDataTypes()
	if len(validTypes) != 1 || validTypes[0] != pvtypes.UUIDType {
		t.Errorf("Expected ValidDataTypes() to return [UUIDType], got %v", validTypes)
	}

	// Test String method
	if constraint.String() != "format[v4]" {
		t.Errorf("Expected String() to return 'v4', got %q", constraint.String())
	}

	// Test ParseBytes method
	parsed, err := constraint.Parse("v7", pvtypes.UUIDType)
	if err != nil {
		t.Errorf("Expected ParseBytes() to succeed, got error: %v", err)
		return
	}
	want := "format[v7]"
	if parsed.String() != want {
		t.Errorf("Expected parsed constraint to have format %q, got %q", want, parsed.String())
	}
}

func TestUUIDVersionDetection(t *testing.T) {
	tests := []struct {
		name            string
		uuid            string
		expectedVersion int
		expectError     bool
	}{
		{"v1-uuid", "550e8400-e29b-11d1-a716-446655440000", 1, false},
		{"v2-uuid", "550e8400-e29b-21d2-a716-446655440000", 2, false},
		{"v3-uuid", "550e8400-e29b-31d3-a716-446655440000", 3, false},
		{"v4-uuid", "550e8400-e29b-41d4-a716-446655440000", 4, false},
		{"v5-uuid", "550e8400-e29b-51d5-a716-446655440000", 5, false},
		{"v6-uuid", "1e890a5d-ac96-674b-b900-4aed2fc33a80", 6, false},
		{"v7-uuid", "01890a5d-ac96-774b-b900-4aed2fc33a80", 7, false},
		{"v8-uuid", "01890a5d-ac96-874b-b900-4aed2fc33a80", 8, false},
		{"invalid-format", "not-a-uuid", 0, true},
		{"wrong-variant", "550e8400-e29b-41d4-2716-446655440000", 0, true}, // Variant bits wrong
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseStandardUUID(tt.uuid)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for UUID %q, but got none", tt.uuid)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for UUID %q: %v", tt.uuid, err)
				return
			}

			if version != tt.expectedVersion {
				t.Errorf("Expected version %d for UUID %q, got %d",
					tt.expectedVersion, tt.uuid, version)
			}
		})
	}
}

func TestUUIDFormatConstraintIntegration(t *testing.T) {
	// Test that the constraint works with the broader constraint parsing system
	constraint, err := ParseUUIDFormatConstraint("v4")
	if err != nil {
		t.Fatalf("Failed to create constraint: %v", err)
	}

	// Verify it implements the Constraint interface properly
	var _ pvtypes.Constraint = constraint

	// Test with valid UUID
	validUUID := "550e8400-e29b-41d4-a716-446655440000"
	if err := constraint.Validate(validUUID); err != nil {
		t.Errorf("Valid UUID failed validation: %v", err)
	}

	// Test with invalid UUID
	invalidUUID := "not-a-uuid"
	if err := constraint.Validate(invalidUUID); err == nil {
		t.Error("Invalid UUID passed validation")
	}
}

// TestUUIDFormatConstraint_EasterEggExamples tests that UUID format constraints
// provide appropriate Easter egg examples for each UUID version format.
// These examples use memorable/recognizable patterns.
func TestUUIDFormatConstraint_EasterEggExamples(t *testing.T) {
	tests := []struct {
		name           string
		formatSpec     string
		wantExample    string
		invalidExample string // An example that should fail validation
	}{
		{
			name:           "uuid_v1_format",
			formatSpec:     "v1",
			wantExample:    "f81d4fae-7dec-11d0-a765-00a0c91e6bf6",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // v4 UUID
		},
		{
			name:           "uuid_v4_format",
			formatSpec:     "v4",
			wantExample:    "deadbeef-cafe-4011-8123-b1d5c0d51234",
			invalidExample: "f81d4fae-7dec-11d0-a765-00a0c91e6bf6", // v1 UUID
		},
		{
			name:           "uuid_v5_format",
			formatSpec:     "v5",
			wantExample:    "2a98f1f0-0a71-50e5-9d51-8650e68d9518",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // v4 UUID
		},
		{
			name:           "uuid_v7_format",
			formatSpec:     "v7",
			wantExample:    "018d9f10-5341-7c91-9e73-b3c14d9b4b0e",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // v4 UUID
		},
		{
			name:           "uuid_v8_format",
			formatSpec:     "v8",
			wantExample:    "20251018-b26a-8025-a12b-4c5d6e7f8a9b",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // v4 UUID
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the constraint
			constraint, err := ParseUUIDFormatConstraint(tt.formatSpec)
			if err != nil {
				t.Fatalf("Failed to parse UUID format constraint %q: %v", tt.formatSpec, err)
			}

			// Get the example
			example := constraint.Example(nil)
			if example == nil {
				t.Fatal("Example() returned nil")
			}

			exampleStr, ok := example.(string)
			if !ok {
				t.Fatalf("Example() returned non-string: %T", example)
			}

			// Verify the example matches what we expect
			if exampleStr != tt.wantExample {
				t.Errorf("Example() = %q, want %q", exampleStr, tt.wantExample)
			}

			// Verify the example validates correctly
			if err := constraint.Validate(exampleStr); err != nil {
				t.Errorf("Example %q failed validation: %v", exampleStr, err)
			}

			// Verify an invalid example fails validation
			if err := constraint.Validate(tt.invalidExample); err == nil {
				t.Errorf("Invalid example %q should have failed validation", tt.invalidExample)
			}
		})
	}
}
