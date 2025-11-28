package test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

// TestSnowflakeComponentValidation tests that Snowflake IDs are properly decoded and validated
func TestSnowflakeComponentValidation(t *testing.T) {
	tests := []struct {
		name        string
		paramSpec   string
		value       string
		shouldPass  bool
		description string
	}{
		{
			name:        "valid_wikipedia_example",
			paramSpec:   "{id:uuid:format[snowflake]}",
			value:       "1888944671579078978",
			shouldPass:  true,
			description: "Wikipedia example with valid timestamp and components",
		},
		{
			name:        "valid_with_custom_epoch",
			paramSpec:   "{id:uuid:format[snowflake:1288834974657]}",
			value:       "1888944671579078978",
			shouldPass:  true,
			description: "Same ID with explicit Twitter epoch",
		},
		{
			name:        "valid_min_value",
			paramSpec:   "{id:uuid:format[snowflake]}",
			value:       "1",
			shouldPass:  true,
			description: "Minimum valid Snowflake (all zeros except ID=1)",
		},
		{
			name:        "invalid_non_numeric",
			paramSpec:   "{id:uuid:format[snowflake]}",
			value:       "abc123",
			shouldPass:  false,
			description: "Non-numeric characters should fail",
		},
		{
			name:        "invalid_uuid_format",
			paramSpec:   "{id:uuid:format[snowflake]}",
			value:       "deadbeef-cafe-4011-8123-b1d5c0d51234",
			shouldPass:  false,
			description: "UUID format should fail Snowflake validation",
		},
		{
			name:        "invalid_empty_string",
			paramSpec:   "{id:uuid:format[snowflake]}",
			value:       "",
			shouldPass:  false,
			description: "Empty string should fail",
		},
		{
			name:        "valid_zero",
			paramSpec:   "{id:uuid:format[snowflake]}",
			value:       "0",
			shouldPass:  true,
			description: "Zero is technically valid (epoch time)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the parameter
			param, err := pathvars.ParseParameter(tt.paramSpec, pathvars.PathLocation)
			if err != nil {
				t.Fatalf("Failed to parse parameter %q: %v", tt.paramSpec, err)
			}

			// Validate the value
			err = param.Validate(tt.value)

			if tt.shouldPass && err != nil {
				t.Errorf("%s: Expected validation to pass but got error: %v", tt.description, err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("%s: Expected validation to fail but it passed", tt.description)
			}
		})
	}
}

// TestSnowflakeEpochParameter tests custom epoch parameter parsing
func TestSnowflakeEpochParameter(t *testing.T) {
	tests := []struct {
		name       string
		paramSpec  string
		testValue  string // Use appropriate test value for each epoch
		shouldFail bool
	}{
		{
			name:       "default_epoch",
			paramSpec:  "{id:uuid:format[snowflake]}",
			testValue:  "1888944671579078978", // Wikipedia example (Twitter epoch)
			shouldFail: false,
		},
		{
			name:       "twitter_epoch_explicit",
			paramSpec:  "{id:uuid:format[snowflake:1288834974657]}",
			testValue:  "1888944671579078978", // Same Wikipedia example
			shouldFail: false,
		},
		{
			name:       "discord_epoch",
			paramSpec:  "{id:uuid:format[snowflake:1420070400000]}",
			testValue:  "1", // Minimal valid ID (just after Discord epoch)
			shouldFail: false,
		},
		{
			name:       "custom_epoch_zero",
			paramSpec:  "{id:uuid:format[snowflake:0]}",
			testValue:  "1", // Minimal valid ID
			shouldFail: false,
		},
		{
			name:       "invalid_epoch_text",
			paramSpec:  "{id:uuid:format[snowflake:not_a_number]}",
			testValue:  "1",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := pathvars.ParseParameter(tt.paramSpec, pathvars.PathLocation)
			if err != nil {
				t.Fatalf("Failed to parse parameter %q: %v", tt.paramSpec, err)
			}

			// Try validating the test value
			err = param.Validate(tt.testValue)

			if tt.shouldFail && err == nil {
				t.Error("Expected epoch parsing to fail but it succeeded")
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Expected epoch parsing to succeed but got error: %v", err)
			}
		})
	}
}
