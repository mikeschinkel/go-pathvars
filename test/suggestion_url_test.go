package test

import (
	"testing"

	"github.com/mikeschinkel/go-pathvars"
	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func newValuesMap(args ...any) pvtypes.ValuesMap {
	vm := pvtypes.NewValuesMap(0)
	for i := 0; i < len(args); i += 2 {
		vm.Set(pathvars.Identifier(args[i].(string)), args[i+1])
	}
	return vm
}

// TestSuggestionURL_ADR018 verifies that SuggestionURL follows ADR-018 guidelines:
// - Only includes required parameters OR parameters user actually provided
// - Correct parameters shown as {PLACEHOLDER}
// - Problematic parameter shown with Example() value
// - Problematic parameter positioned LAST in query string
func TestSuggestionURL_ADR018(t *testing.T) {
	tests := []struct {
		name               string
		templateStr        string
		problematicParam   string
		userProvidedParams pvtypes.ValuesMap
		expectedURL        string
	}{
		{
			name:             "missing_required_email_user_provided_min_score",
			templateStr:      "GET /api/users/search?{email:email}&{min_score:int}",
			problematicParam: "email",
			userProvidedParams: newValuesMap(
				"min_score", 50,
			),
			expectedURL: "/GET /api/users/search?min_score={MIN_SCORE}&email=user@example.com",
		},
		{
			name:               "invalid_min_score_no_other_params_provided",
			templateStr:        "GET /api/users/search?{min_score:int}&{email?:email}",
			problematicParam:   "min_score",
			userProvidedParams: newValuesMap("min_score", "invalid"),
			expectedURL:        "/GET /api/users/search?min_score=123",
		},
		{
			name:             "multiple_user_provided_one_failing",
			templateStr:      "GET /api/posts?{user_id:int}&{category:string}&{limit?:int}",
			problematicParam: "limit",
			userProvidedParams: newValuesMap(
				"user_id", 123,
				"category", "tech",
				"limit", "invalid",
			),
			expectedURL: "/GET /api/posts?user_id={USER_ID}&category={CATEGORY}&limit=123",
		},
		{
			name:             "path_parameter_error_with_query_params",
			templateStr:      "GET /api/users/{user_id:int}/posts?{active?:bool}",
			problematicParam: "user_id",
			userProvidedParams: newValuesMap(
				"user_id", "invalid-id",
				"active", true,
			),
			expectedURL: "/GET /api/users/123/posts?active={ACTIVE}",
		},
		{
			name:             "optional_parameter_provided_with_invalid_value",
			templateStr:      "GET /api/search?{category:string}&{created_after?:date}&{limit?:int}&{offset?:int}",
			problematicParam: "created_after",
			userProvidedParams: newValuesMap(
				"category", "tech",
				"created_after", "invalid-date",
			),
			// limit and offset NOT shown - user didn't provide them
			expectedURL: "/GET /api/search?category={CATEGORY}&created_after=1999-12-31",
		},
		{
			name:               "only_required_parameter_fails_nothing_else_provided",
			templateStr:        "GET /api/search?{email:email}&{limit?:int}&{offset?:int}",
			problematicParam:   "email",
			userProvidedParams: newValuesMap(),
			// No other params shown - user didn't provide any, none others are required
			expectedURL: "/GET /api/search?email=user@example.com",
		},
		{
			name:             "all_parameters_fail",
			templateStr:      "GET /api/search?{email:email}&{min_score:int}",
			problematicParam: "email",
			userProvidedParams: newValuesMap(
				"email", "bad",
				"min_score", "invalid",
			),
			// Both required, both provided, email is problematic so goes last
			expectedURL: "/GET /api/search?min_score={MIN_SCORE}&email=user@example.com",
		},
		{
			name:             "query_param_problematic_with_multiple_correct",
			templateStr:      "GET /api/products?{category:string}&{min_price:int}&{max_price:int}&{in_stock?:bool}",
			problematicParam: "max_price",
			userProvidedParams: newValuesMap(
				"category", "electronics",
				"min_price", 100,
				"max_price", "invalid",
				"in_stock", true,
			),
			// Note: Query param order may vary based on map iteration
			expectedURL: "/GET /api/products?category={CATEGORY}&min_price={MIN_PRICE}&in_stock={IN_STOCK}&max_price=123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the template
			template, err := pathvars.ParseTemplate(tt.templateStr)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			// Find the problematic parameter
			problematicParam, exists := template.Parameters().Get(pathvars.Identifier(tt.problematicParam))
			if !exists {
				t.Fatalf("Problematic parameter %s not found in template", tt.problematicParam)
			}

			// Generate suggestion URL
			got := template.Example(&pvtypes.ExampleArgs{
				ProblematicParam:   problematicParam,
				UserProvidedParams: &tt.userProvidedParams,
				ValidationErr:      nil,
			})

			// Compare
			if got != tt.expectedURL {
				t.Errorf("SuggestionURL() mismatch:\n  got:  %s\n  want: %s", got, tt.expectedURL)
			}
		})
	}
}

// TestSuggestionURL_PathParameterProblematic tests that path parameters can be problematic
// and are shown with Example() value (though they can't be moved to LAST position)
func TestSuggestionURL_PathParameterProblematic(t *testing.T) {
	tests := []struct {
		name               string
		templateStr        string
		problematicParam   string
		userProvidedParams pvtypes.ValuesMap
		expectedURL        string
	}{
		{
			name:             "path_param_is_problematic",
			templateStr:      "GET /api/users/{user_id:int}",
			problematicParam: "user_id",
			userProvidedParams: newValuesMap(
				"user_id", "abc",
			),
			expectedURL: "/GET /api/users/123",
		},
		{
			name:             "path_param_problematic_with_query_params",
			templateStr:      "GET /api/users/{user_id:int}/posts?{limit?:int}&{offset?:int}",
			problematicParam: "user_id",
			userProvidedParams: newValuesMap(
				"user_id", "abc",
				"limit", 10,
			),
			// limit shown because user provided it, offset not shown because user didn't provide it
			expectedURL: "/GET /api/users/123/posts?limit={LIMIT}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := pathvars.ParseTemplate(tt.templateStr)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			problematicParam, exists := template.Parameters().Get(pvtypes.Identifier(tt.problematicParam))
			if !exists {
				t.Fatalf("Problematic parameter %s not found in template", tt.problematicParam)
			}

			got := template.Example(&pvtypes.ExampleArgs{
				ProblematicParam:   problematicParam,
				UserProvidedParams: &tt.userProvidedParams,
				ValidationErr:      nil,
			})

			if got != tt.expectedURL {
				t.Errorf("SuggestionURL() mismatch:\n  got:  %s\n  want: %s", got, tt.expectedURL)
			}
		})
	}
}

// TestSuggestionURL_EmptyUserProvided tests behavior when user provides no parameters
func TestSuggestionURL_EmptyUserProvided(t *testing.T) {
	tests := []struct {
		name             string
		templateStr      string
		problematicParam string
		expectedURL      string
	}{
		{
			name:             "required_param_missing",
			templateStr:      "GET /api/users?{email:email}&{name?:string}",
			problematicParam: "email",
			// Only required param shown, optional not shown
			expectedURL: "/GET /api/users?email=user@example.com",
		},
		{
			name:             "multiple_required_params_one_missing",
			templateStr:      "GET /api/search?{query:string}&{category:string}",
			problematicParam: "query",
			// Both required params shown
			expectedURL: "/GET /api/search?category={CATEGORY}&query=abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template, err := pathvars.ParseTemplate(tt.templateStr)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			problematicParam, exists := template.Parameters().Get(pathvars.Identifier(tt.problematicParam))
			if !exists {
				t.Fatalf("Problematic parameter %s not found in template", tt.problematicParam)
			}

			got := template.Example(&pvtypes.ExampleArgs{
				ProblematicParam:   problematicParam,
				UserProvidedParams: nil,
				ValidationErr:      nil,
			})

			if got != tt.expectedURL {
				t.Errorf("SuggestionURL() mismatch:\n  got:  %s\n  want: %s", got, tt.expectedURL)
			}
		})
	}
}
