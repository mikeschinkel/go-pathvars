package test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

// TestUUIDFormatConstraint_Examples tests that UUID format constraints
// provide appropriate Easter egg examples for each format.
func TestUUIDFormatConstraint_Examples(t *testing.T) {
	tests := []struct {
		name           string
		paramSpec      string
		wantExample    string
		invalidExample string // An example that should fail validation
	}{
		{
			name:           "uuid_v4_format",
			paramSpec:      "{id:uuid:format[v4]}",
			wantExample:    "deadbeef-cafe-4011-8123-b1d5c0d51234",
			invalidExample: "f81d4fae-7dec-11d0-a765-00a0c91e6bf6", // v1 UUID
		},
		{
			name:           "uuid_v1_format",
			paramSpec:      "{id:uuid:format[v1]}",
			wantExample:    "f81d4fae-7dec-11d0-a765-00a0c91e6bf6",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // v4 UUID
		},
		{
			name:           "uuid_v5_format",
			paramSpec:      "{id:uuid:format[v5]}",
			wantExample:    "2a98f1f0-0a71-50e5-9d51-8650e68d9518",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // v4 UUID
		},
		{
			name:           "uuid_v7_format",
			paramSpec:      "{id:uuid:format[v7]}",
			wantExample:    "018d9f10-5341-7c91-9e73-b3c14d9b4b0e",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // v4 UUID
		},
		{
			name:           "uuid_v8_format",
			paramSpec:      "{id:uuid:format[v8]}",
			wantExample:    "20251018-b26a-8025-a12b-4c5d6e7f8a9b",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // v4 UUID
		},
		{
			name:           "ulid_format",
			paramSpec:      "{id:uuid:format[ulid]}",
			wantExample:    "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // UUID not ULID
		},
		{
			name:           "ksuid_format",
			paramSpec:      "{id:uuid:format[ksuid]}",
			wantExample:    "0ujsswThIGTUYm2K8FjOOfXtY1K",
			invalidExample: "01ARZ3NDEKTSV4RRFFQ69G5FAV", // ULID not KSUID
		},
		{
			name:           "nanoid_format",
			paramSpec:      "{id:uuid:format[nanoid]}",
			wantExample:    "V1StGXR8_Z5jdHi6B-myT",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // UUID not NanoID
		},
		{
			name:           "cuid_format",
			paramSpec:      "{id:uuid:format[cuid]}",
			wantExample:    "ckf0f9e5x0000q3yz4dq7a1qf",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // UUID not CUID
		},
		{
			name:           "snowflake_format",
			paramSpec:      "{id:uuid:format[snowflake]}",
			wantExample:    "1888944671579078978",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // UUID not Snowflake
		},
		{
			name:           "snowflake_custom_epoch",
			paramSpec:      "{id:uuid:format[snowflake:1288834974657]}",
			wantExample:    "1888944671579078978",
			invalidExample: "deadbeef-cafe-4011-8123-b1d5c0d51234", // UUID not Snowflake
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the parameter
			param, err := pathvars.ParseParameter(tt.paramSpec, pathvars.PathLocation)
			if err != nil {
				t.Fatalf("Failed to parse parameter %q: %v", tt.paramSpec, err)
			}

			// Get the example
			example := param.Example(err, nil)
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
			if err := param.Validate(exampleStr); err != nil {
				t.Errorf("Example %q failed validation: %v", exampleStr, err)
			}

			// Verify an invalid example fails validation
			if err := param.Validate(tt.invalidExample); err == nil {
				t.Errorf("Invalid example %q should have failed validation", tt.invalidExample)
			}
		})
	}
}

// TestParameterExample_FallbackToDataType tests that Parameter.Example()
// falls back to the data type example when no constraint provides one.
// Note: Data type examples may not satisfy all constraints (e.g., "abc" < 5 chars).
func TestParameterExample_FallbackToDataType(t *testing.T) {
	tests := []struct {
		name           string
		paramSpec      string
		expectType     string
		shouldValidate bool // Whether we expect the example to validate
	}{
		{
			name:           "uuid_without_format",
			paramSpec:      "{id:uuid}",
			expectType:     "string",
			shouldValidate: true, // UUID data type example should validate
		},
		{
			name:           "int_with_range",
			paramSpec:      "{id:int:range[1..100]}",
			expectType:     "int",
			shouldValidate: true, // Integer example (123) falls within range
		},
		{
			name:           "string_with_length",
			paramSpec:      "{name:string:length[5..50]}",
			expectType:     "string",
			shouldValidate: false, // String example "abc" is only 3 chars, doesn't satisfy length[5..50]
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := pathvars.ParseParameter(tt.paramSpec, pathvars.PathLocation)
			if err != nil {
				t.Fatalf("Failed to parse parameter %q: %v", tt.paramSpec, err)
			}

			example := param.Example(err, nil)
			if example == nil {
				t.Fatal("Example() returned nil")
			}

			// t.Logf("Example for %s: %v (type: %T)", tt.paramSpec, example, example)

			// Verify the example validates if expected
			exampleStr, ok := example.(string)
			if !ok {
				// For int type, example might be int - that's fine
				return
			}

			err = param.Validate(exampleStr)
			if tt.shouldValidate && err != nil {
				t.Errorf("Example %q should have validated but got error: %v", exampleStr, err)
			}
			//if !tt.shouldValidate && err == nil {
			//	// t.Logf("Note: Example %q doesn't satisfy constraints (expected)", exampleStr)
			//}
		})
	}
}
