package test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

func TestImplicitTypeInference(t *testing.T) {
	tests := []struct {
		name         string
		template     string
		expectError  bool
		expectedType string
		description  string
	}{
		// Implicit type inference from parameter name
		{"infer-string", "/users/{string}", false, "string", "Infer string type from parameter name"},
		{"infer-int", "/users/{int}", false, "integer", "Infer int type from parameter name"},
		{"infer-decimal", "/posts/{decimal}", false, "decimal", "Infer decimal type from parameter name"},
		{"infer-real", "/measurements/{real}", false, "real", "Infer real type from parameter name"},
		{"infer-identifier", "/api/{identifier}", false, "identifier", "Infer identifier type from parameter name"},
		{"infer-date", "/events/{date}", false, "date", "Infer date type from parameter name"},
		{"infer-uuid", "/objects/{uuid}", false, "uuid", "Infer uuid type from parameter name"},
		{"infer-alphanum", "/codes/{alphanum}", false, "alphanumeric", "Infer alphanum type from parameter name"},
		{"infer-slug", "/articles/{slug}", false, "slug", "Infer slug type from parameter name"},
		{"infer-bool", "/settings/{bool}", false, "boolean", "Infer bool type from parameter name"},
		{"infer-email", "/contacts/{email}", false, "email", "Infer email type from parameter name"},

		// Non-matching names should default to string
		{"no-infer-custom", "/users/{userId}", false, "string", "Non-matching name defaults to string"},
		{"no-infer-mixed", "/api/{myValue}", false, "string", "Mixed name defaults to string"},

		// Multi-segment with implicit types
		{"infer-multiseg-date", "/archive/{date*}", false, "date", "Multi-segment with inferred date type"},
		{"infer-multiseg-string", "/files/{string*}", false, "string", "Multi-segment with inferred string type"},

		// Optional parameters with implicit types
		{"infer-optional-int", "/users?{int?42}", false, "integer", "Optional parameter with inferred int type"},
		{"infer-optional-bool", "/settings?{bool?true}", false, "boolean", "Optional parameter with inferred bool type"},

		// Double colon syntax for constraints with implicit types
		{"double-colon-slug", "/articles/{slug::enum[news,sports,tech]}", false, "slug", "Double colon with slug type and enum constraint"},
		{"double-colon-int", "/items/{int::range[1..100]}", false, "integer", "Double colon with int type and range constraint"},
		{"double-colon-date", "/events/{date::format[utc]}", false, "date", "Double colon with date type and format constraint"},

		// Double colon syntax
		{"double-colon-invalid", "/items/{invalidname::range[1..100]}", true, "", "Double colon with non-type name should error"},
		{"double-colon-mixed", "/users/{userId::length[5..20]}", false, "", "Double colon with non-type name should NOT error"},

		// Mixed syntax in same template
		{"mixed-explicit-implicit", "/users/{id:int}/posts/{slug}", false, "", "Mix explicit and implicit types"},
		{"mixed-implicit-double-colon", "/api/{date}/items/{int::range[1..100]}", false, "", "Mix implicit and double colon syntax"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := pathvars.NewRouter()
			method, path := parsePathSpec(tt.template)
			err := router.AddRoute(pathvars.HTTPMethod(method), pathvars.Template(path), nil)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for template %q, but got none", tt.template)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for template %q: %v", tt.template, err)
				return
			}

			// For successful cases, we could add more specific validation here
			// For now, just verify the template parses without error
		})
	}
}

// Note: GetDataType is not exported, so we test it indirectly through ParseParameter
// This keeps the API clean while still testing the functionality

func TestParameterParsingWithImplicitTypes(t *testing.T) {
	tests := []struct {
		name         string
		paramSpec    string
		expectError  bool
		expectedType pathvars.PVDataType
		description  string
	}{
		// Basic implicit type cases
		{"simple-int", "{int}", false, pathvars.IntegerType, "Simple int parameter"},
		{"simple-string", "{string}", false, pathvars.StringType, "Simple string parameter"},
		{"simple-date", "{date}", false, pathvars.DateType, "Simple date parameter"},
		{"simple-real", "{real}", false, pathvars.RealType, "Simple real parameter"},

		// With constraints using double colon
		{"int-with-constraint", "{int::range[1..100]}", false, pathvars.IntegerType, "Int with range constraint"},
		{"slug-with-enum", "{slug::enum[news,sports,tech]}", false, pathvars.SlugType, "Slug with enum constraint"},

		// Optional with implicit types
		{"optional-bool", "{bool?true}", false, pathvars.BooleanType, "Optional bool with default"},
		{"optional-int", "{int?42}", false, pathvars.IntegerType, "Optional int with default"},

		// Multi-segment with implicit types
		{"multisegment-date", "{date*}", false, pathvars.DateType, "Multi-segment date"},

		// Non-matching names should use string type
		{"custom-name", "{userId}", false, pathvars.StringType, "Custom name defaults to string"},

		// Error cases
		{"invalid-double-colon", "{customName::range[1..100]}", true, pathvars.UnspecifiedDataType, "Invalid name with double colon"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			param, err := pathvars.ParseParameter(tt.paramSpec, pathvars.IrrelevantLocationType)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for spec %q, but got none", tt.paramSpec)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for spec %q: %v", tt.paramSpec, err)
				return
			}

			if param.DataType() != tt.expectedType {
				t.Errorf("Expected type %v for spec %q, got %v", tt.expectedType, tt.paramSpec, param.DataType())
			}
		})
	}
}
