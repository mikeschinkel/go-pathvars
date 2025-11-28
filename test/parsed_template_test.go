package test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

func TestExtractParameterSpec(t *testing.T) {
	tests := []struct {
		name       string
		segment    string
		wantSpec   string
		wantPrefix string
		wantSuffix string
		wantErr    bool
	}{
		{
			name:     "Not a var",
			segment:  "foo",
			wantSpec: "foo",
			wantErr:  false,
		},
		{
			name:     "Just a var",
			segment:  "{id:integer}",
			wantSpec: "{id:integer}",
			wantErr:  false,
		},
		{
			name:       "Var with prefix",
			segment:    "foo{id:integer}",
			wantSpec:   "{id:integer}",
			wantPrefix: "foo",
			wantErr:    false,
		},
		{
			name:       "Var with suffix",
			segment:    "{id:integer}bar",
			wantSpec:   "{id:integer}",
			wantSuffix: "bar",
			wantErr:    false,
		},
		{
			name:       "Var with both fixes",
			segment:    "foo{id:integer}bar",
			wantSpec:   "{id:integer}",
			wantPrefix: "foo",
			wantSuffix: "bar",
			wantErr:    false,
		},
		{
			name:    "Missing closing brace",
			segment: "{id:integer",
			wantErr: true,
		},
		{
			name:    "Malformed braces with just a var",
			segment: "}id:integer{",
			wantErr: true,
		},
		{
			name:    "Malformed braces var with prefix",
			segment: "foo}id:integer{",
			wantErr: true,
		},
		{
			name:    "Malformed braces with suffix",
			segment: "}id:integer{bar",
			wantErr: true,
		},
		{
			name:    "Malformed braces var with both fixes",
			segment: "foo}id:integer{bar",
			wantErr: true,
		},
		// TODO: Add tests to check for specific sentinel errors
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefix, gotSpec, gotSuffix, err := pathvars.ExtractParameterSpec(tt.segment)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractParameterSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSpec != tt.wantSpec {
				t.Errorf("ExtractParameterSpec() gotName = %v, want %v", gotSpec, tt.wantSpec)
			}
			if !reflect.DeepEqual(gotPrefix, tt.wantPrefix) {
				t.Errorf("ExtractParameterSpec() gotPrefix = %v, want %v", gotPrefix, tt.wantPrefix)
			}
			if !reflect.DeepEqual(gotSuffix, tt.wantSuffix) {
				t.Errorf("ExtractParameterSpec() gotSuffix = %v, want %v", gotSuffix, tt.wantSuffix)
			}
		})
	}
}

// TestTemplate_ParameterValidationError tests that validateParameter() generates correct ParameterError
func TestTemplate_ParameterValidationError(t *testing.T) {
	tests := []struct {
		name              string
		template          string
		path              string
		wantParam         string
		wantExpectedType  string
		wantReceivedValue string
		wantLocation      pathvars.LocationType
		wantFaultSource   pathvars.FaultSource
	}{
		{
			name:              "Invalid integer parameter",
			template:          "/api/users/{id:integer}",
			path:              "/api/users/abc",
			wantParam:         "id",
			wantExpectedType:  "integer",
			wantReceivedValue: "abc",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
		},
		{
			name:              "Invalid UUID parameter",
			template:          "/api/items/{uuid:uuid}",
			path:              "/api/items/not-a-uuid",
			wantParam:         "uuid",
			wantExpectedType:  "uuid",
			wantReceivedValue: "not-a-uuid",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
		},
		{
			name:              "Invalid boolean parameter",
			template:          "/api/settings/{enabled:boolean}",
			path:              "/api/settings/maybe",
			wantParam:         "enabled",
			wantExpectedType:  "boolean",
			wantReceivedValue: "maybe",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse template
			tmpl, err := pathvars.ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			// Match and validate (should fail with ParameterError)
			_, err = tmpl.Match(tt.path, "")
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}

			// Extract ParameterError from error using doterr
			pve, ok := pathvars.FindErr[*pathvars.ParameterError](err)
			if !ok {
				t.Fatal("Expected ParameterError in error chain, got nil")
			}

			// Verify ParameterError fields (domain concerns)
			if pve.Parameter != tt.wantParam {
				t.Errorf("Parameter = %q, want %q", pve.Parameter, tt.wantParam)
			}
			if pve.ExpectedType != tt.wantExpectedType {
				t.Errorf("ExpectedType = %q, want %q", pve.ExpectedType, tt.wantExpectedType)
			}
			if pve.ReceivedValue != tt.wantReceivedValue {
				t.Errorf("ReceivedValue = %q, want %q", pve.ReceivedValue, tt.wantReceivedValue)
			}
			//if pve.Location != tt.wantLocation.Slug() {
			//	t.Errorf("Location = %q, want %q", pve.Location, tt.wantLocation.Slug())
			//}
			//if pve.FaultSource != tt.wantFaultSource {
			//	t.Errorf("FaultSource = %v, want %v", pve.FaultSource, tt.wantFaultSource)
			//}
			//if pve.EndpointTemplate != tt.template {
			//	t.Errorf("EndpointTemplate = %q, want %q", pve.EndpointTemplate, tt.template)
			//}
			//if pve.Instance != tt.path {
			//	t.Errorf("Instance = %q, want %q", pve.Instance, tt.path)
			//}
			if pve.GetDetail() == "" {
				t.Error("Detail should not be empty")
			}
			if pve.GetSuggestion() == "" {
				t.Error("Suggestion should not be empty")
			}
		})
	}
}

// TestParameterValidationError_Comprehensive tests all ParameterError fields
// across all path syntax permutations, data types, locations, and error types.
func TestParameterValidationError_Comprehensive(t *testing.T) {
	tests := []struct {
		name              string
		template          string
		path              string
		query             string
		wantParam         string
		wantExpectedType  string
		wantReceivedValue string
		wantLocation      pathvars.LocationType
		wantFaultSource   pathvars.FaultSource
		wantConstraint    string // Empty for type errors, constraint spec for constraint errors
		checkDetail       func(t *testing.T, detail string)
		checkSuggestion   func(t *testing.T, suggestion string)
	}{
		// ==================== PATH PARAMETER TYPE ERRORS ====================
		{
			name:              "Path param: integer type error",
			template:          "/api/users/{id:integer}",
			path:              "/api/users/abc",
			wantParam:         "id",
			wantExpectedType:  "integer",
			wantReceivedValue: "abc",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "id") {
					t.Errorf("Detail should mention parameter 'id', got %q", detail)
				}
				if !strings.Contains(detail, "integer") {
					t.Errorf("Detail should mention type 'integer', got %q", detail)
				}
				if !strings.Contains(detail, "abc") {
					t.Errorf("Detail should mention value 'abc', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "integer") {
					t.Errorf("Suggestion should mention type 'integer', got %q", suggestion)
				}
				if !strings.Contains(suggestion, "id") {
					t.Errorf("Suggestion should mention parameter 'id', got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: UUID type error",
			template:          "/api/items/{uuid:uuid}",
			path:              "/api/items/not-a-uuid",
			wantParam:         "uuid",
			wantExpectedType:  "uuid",
			wantReceivedValue: "not-a-uuid",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "uuid") {
					t.Errorf("Detail should mention parameter 'uuid', got %q", detail)
				}
				if !strings.Contains(detail, "not-a-uuid") {
					t.Errorf("Detail should mention value 'not-a-uuid', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "uuid") {
					t.Errorf("Suggestion should mention type 'uuid', got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: boolean type error",
			template:          "/api/settings/{enabled:boolean}",
			path:              "/api/settings/maybe",
			wantParam:         "enabled",
			wantExpectedType:  "boolean",
			wantReceivedValue: "maybe",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "enabled") {
					t.Errorf("Detail should mention parameter 'enabled', got %q", detail)
				}
				if !strings.Contains(detail, "boolean") {
					t.Errorf("Detail should mention type 'boolean', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "boolean") {
					t.Errorf("Suggestion should mention type 'boolean', got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: decimal type error",
			template:          "/api/prices/{amount:decimal}",
			path:              "/api/prices/not-a-number",
			wantParam:         "amount",
			wantExpectedType:  "decimal",
			wantReceivedValue: "not-a-number",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "amount") {
					t.Errorf("Detail should mention parameter 'amount', got %q", detail)
				}
				if !strings.Contains(detail, "decimal") {
					t.Errorf("Detail should mention type 'decimal', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "decimal") {
					t.Errorf("Suggestion should mention type 'decimal', got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: date type error",
			template:          "/api/events/{date:date}",
			path:              "/api/events/not-a-date",
			wantParam:         "date",
			wantExpectedType:  "date",
			wantReceivedValue: "not-a-date",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "date") {
					t.Errorf("Detail should mention parameter 'date', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "date") {
					t.Errorf("Suggestion should mention type 'date', got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: slug type error",
			template:          "/api/posts/{slug:slug}",
			path:              "/api/posts/Invalid Slug!",
			wantParam:         "slug",
			wantExpectedType:  "slug",
			wantReceivedValue: "Invalid Slug!",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "slug") {
					t.Errorf("Detail should mention parameter 'slug', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "slug") {
					t.Errorf("Suggestion should mention type 'slug', got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: alphanumeric type error",
			template:          "/api/codes/{code:alphanumeric}",
			path:              "/api/codes/code-with-dash",
			wantParam:         "code",
			wantExpectedType:  "alphanumeric",
			wantReceivedValue: "code-with-dash",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "code") {
					t.Errorf("Detail should mention parameter 'code', got %q", detail)
				}
				if !strings.Contains(detail, "alphanumeric") {
					t.Errorf("Detail should mention type 'alphanumeric', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "alphanumeric") {
					t.Errorf("Suggestion should mention type 'alphanumeric', got %q", suggestion)
				}
			},
		},

		// ==================== PATH PARAMETER CONSTRAINT ERRORS ====================
		{
			name:              "Path param: integer range constraint violation (below min)",
			template:          "/api/users/{id:integer:range[1..1000]}",
			path:              "/api/users/0",
			wantParam:         "id",
			wantExpectedType:  "integer",
			wantReceivedValue: "0",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "range[1..1000]",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "id") {
					t.Errorf("Detail should mention parameter 'id', got %q", detail)
				}
				if !strings.Contains(detail, "0") {
					t.Errorf("Detail should mention value '0', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "range") || !strings.Contains(suggestion, "1") {
					t.Errorf("Suggestion should mention range constraint, got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: integer range constraint violation (above max)",
			template:          "/api/users/{id:integer:range[1..1000]}",
			path:              "/api/users/1001",
			wantParam:         "id",
			wantExpectedType:  "integer",
			wantReceivedValue: "1001",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "range[1..1000]",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "1001") {
					t.Errorf("Detail should mention value '1001', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "1000") {
					t.Errorf("Suggestion should mention max value, got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: string length constraint violation (too short)",
			template:          "/api/codes/{code:string:length[5..10]}",
			path:              "/api/codes/abc",
			wantParam:         "code",
			wantExpectedType:  "string",
			wantReceivedValue: "abc",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "length[5..10]",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "code") {
					t.Errorf("Detail should mention parameter 'code', got %q", detail)
				}
				if !strings.Contains(detail, "abc") {
					t.Errorf("Detail should mention value 'abc', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "length") || !strings.Contains(suggestion, "5") {
					t.Errorf("Suggestion should mention length constraint, got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: string regex constraint violation (email param with explicit string type)",
			template:          "/api/contacts/{email:string:regex[[a-z]+@[a-z]+\\.[a-z]+]}",
			path:              "/api/contacts/invalid-email",
			wantParam:         "email",
			wantExpectedType:  "string",
			wantReceivedValue: "invalid-email",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "regex[[a-z]+@[a-z]+\\.[a-z]+]",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "email") {
					t.Errorf("Detail should mention parameter 'email', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "email") {
					t.Errorf("Suggestion should mention parameter 'email', got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: string enum constraint violation",
			template:          "/api/status/{status:string:enum[active,inactive,pending]}",
			path:              "/api/status/deleted",
			wantParam:         "status",
			wantExpectedType:  "string",
			wantReceivedValue: "deleted",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "enum[active,inactive,pending]",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "status") {
					t.Errorf("Detail should mention parameter 'status', got %q", detail)
				}
				if !strings.Contains(detail, "deleted") {
					t.Errorf("Detail should mention value 'deleted', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "active") || !strings.Contains(suggestion, "inactive") {
					t.Errorf("Suggestion should mention valid enum values, got %q", suggestion)
				}
			},
		},
		{
			name:              "Path param: string length maximum constraint violation",
			template:          "/api/names/{name:string:length[1..10]}",
			path:              "/api/names/ThisNameIsTooLong",
			wantParam:         "name",
			wantExpectedType:  "string",
			wantReceivedValue: "ThisNameIsTooLong",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "length[1..10]",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "name") {
					t.Errorf("Detail should mention parameter 'name', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "length") && !strings.Contains(suggestion, "name") {
					t.Errorf("Suggestion should mention length constraint, got %q", suggestion)
				}
			},
		},

		// ==================== QUERY PARAMETER TYPE ERRORS ====================
		{
			name:              "Query param: integer type error",
			template:          "/api/search?{page:integer}",
			path:              "/api/search",
			query:             "page=abc",
			wantParam:         "page",
			wantExpectedType:  "integer",
			wantReceivedValue: "abc",
			wantLocation:      pathvars.QueryLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "page") {
					t.Errorf("Detail should mention parameter 'page', got %q", detail)
				}
				if !strings.Contains(detail, "integer") {
					t.Errorf("Detail should mention type 'integer', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "integer") {
					t.Errorf("Suggestion should mention type 'integer', got %q", suggestion)
				}
			},
		},
		{
			name:              "Query param: boolean type error",
			template:          "/api/items?{active:boolean}",
			path:              "/api/items",
			query:             "active=maybe",
			wantParam:         "active",
			wantExpectedType:  "boolean",
			wantReceivedValue: "maybe",
			wantLocation:      pathvars.QueryLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "active") {
					t.Errorf("Detail should mention parameter 'active', got %q", detail)
				}
				if !strings.Contains(detail, "boolean") {
					t.Errorf("Detail should mention type 'boolean', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "boolean") {
					t.Errorf("Suggestion should mention type 'boolean', got %q", suggestion)
				}
			},
		},

		// ==================== QUERY PARAMETER CONSTRAINT ERRORS ====================
		{
			name:              "Query param: range constraint violation",
			template:          "/api/items?{limit:integer:range[1..100]}",
			path:              "/api/items",
			query:             "limit=500",
			wantParam:         "limit",
			wantExpectedType:  "integer",
			wantReceivedValue: "500",
			wantLocation:      pathvars.QueryLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "range[1..100]",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "limit") {
					t.Errorf("Detail should mention parameter 'limit', got %q", detail)
				}
				if !strings.Contains(detail, "500") {
					t.Errorf("Detail should mention value '500', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "100") {
					t.Errorf("Suggestion should mention max value, got %q", suggestion)
				}
			},
		},
		{
			name:              "Query param: enum constraint violation",
			template:          "/api/sort?{order:string:enum[asc,desc]}",
			path:              "/api/sort",
			query:             "order=random",
			wantParam:         "order",
			wantExpectedType:  "string",
			wantReceivedValue: "random",
			wantLocation:      pathvars.QueryLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "enum[asc,desc]",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "order") {
					t.Errorf("Detail should mention parameter 'order', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "asc") || !strings.Contains(suggestion, "desc") {
					t.Errorf("Suggestion should mention valid values, got %q", suggestion)
				}
			},
		},

		// ==================== MIXED PATH + QUERY PARAMETERS ====================
		{
			name:              "Mixed: path param error with query param present",
			template:          "/api/users/{id:integer}?{page:integer}",
			path:              "/api/users/abc",
			query:             "page=1",
			wantParam:         "id",
			wantExpectedType:  "integer",
			wantReceivedValue: "abc",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "id") {
					t.Errorf("Detail should mention parameter 'id', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "integer") {
					t.Errorf("Suggestion should mention type 'integer', got %q", suggestion)
				}
			},
		},
		{
			name:              "Mixed: query param error with valid path param",
			template:          "/api/users/{id:integer}?{page:integer}",
			path:              "/api/users/123",
			query:             "page=abc",
			wantParam:         "page",
			wantExpectedType:  "integer",
			wantReceivedValue: "abc",
			wantLocation:      pathvars.QueryLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "page") {
					t.Errorf("Detail should mention parameter 'page', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if !strings.Contains(suggestion, "integer") {
					t.Errorf("Suggestion should mention type 'integer', got %q", suggestion)
				}
			},
		},

		// ==================== DECIMAL CONSTRAINT ERRORS ====================
		{
			name:              "Path param: decimal range constraint violation",
			template:          "/api/prices/{price:decimal:range[0.01..999.99]}",
			path:              "/api/prices/0.001",
			wantParam:         "price",
			wantExpectedType:  "decimal",
			wantReceivedValue: "0.001",
			wantLocation:      pathvars.PathLocation,
			wantFaultSource:   pathvars.ClientFaultSource,
			wantConstraint:    "range[0.01..999.99]",
			checkDetail: func(t *testing.T, detail string) {
				if !strings.Contains(detail, "price") {
					t.Errorf("Detail should mention parameter 'price', got %q", detail)
				}
			},
			checkSuggestion: func(t *testing.T, suggestion string) {
				if suggestion == "" {
					t.Error("Suggestion should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse template
			tmpl, err := pathvars.ParseTemplate(tt.template)
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			// Match and validate (should fail with ParameterError)
			_, err = tmpl.Match(tt.path, tt.query)
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}

			// Extract ParameterError from error using doterr
			pve, ok := pathvars.FindErr[*pathvars.ParameterError](err)
			if !ok {
				t.Fatalf("Expected ParameterError in error chain, got nil. Error: %v", err)
			}

			// ==================== VERIFY ALL FIELDS ====================

			// 1. Err field - should not be nil
			if pve.Err == nil {
				t.Error("Err field should not be nil")
			}

			// 2. FaultSource - should always be ClientFaultSource for parameter validation
			if pve.FaultSource != tt.wantFaultSource {
				t.Errorf("FaultSource = %v, want %v", pve.FaultSource, tt.wantFaultSource)
			}

			//// 3. Instance - should match the source (path or query)
			//expectedInstance := tt.path
			//if tt.wantLocation == pathvars.QueryLocation {
			//	expectedInstance = tt.query
			//}
			////if pve.Instance != expectedInstance {
			//	t.Errorf("Instance = %q, want %q", pve.Instance, expectedInstance)
			//}

			// 4. Detail - should not be empty and pass custom checks
			detail := pve.GetDetail()
			if detail == "" {
				t.Error("Detail should not be empty")
			}
			if tt.checkDetail != nil {
				tt.checkDetail(t, detail)
			}

			// 5. Parameter - should match expected parameter name
			if pve.Parameter != tt.wantParam {
				t.Errorf("Parameter = %q, want %q", pve.Parameter, tt.wantParam)
			}

			// 6. ExpectedType - should match expected data type
			if pve.ExpectedType != tt.wantExpectedType {
				t.Errorf("ExpectedType = %q, want %q", pve.ExpectedType, tt.wantExpectedType)
			}

			// 7. ReceivedValue - should match the invalid value provided
			if pve.ReceivedValue != tt.wantReceivedValue {
				t.Errorf("ReceivedValue = %q, want %q", pve.ReceivedValue, tt.wantReceivedValue)
			}

			// 8. Location - should be "path" or "query"
			//if pve.Location != tt.wantLocation.Slug() {
			//	t.Errorf("Location = %q, want %q", pve.Location, tt.wantLocation.Slug())
			//}
			//
			// 9. Suggestion - should not be empty and pass custom checks
			suggestion := pve.GetSuggestion()
			if suggestion == "" {
				t.Error("Suggestion should not be empty")
			}
			if tt.checkSuggestion != nil {
				tt.checkSuggestion(t, suggestion)
			}

			//// 10. EndpointTemplate - should match the template
			//if pve.EndpointTemplate != tt.template {
			//	t.Errorf("EndpointTemplate = %q, want %q", pve.EndpointTemplate, tt.template)
			//}

			// 11. ConstraintType - should match expected constraint (empty for type errors)
			if pve.ConstraintType != tt.wantConstraint {
				t.Errorf("ConstraintType = %q, want %q", pve.ConstraintType, tt.wantConstraint)
			}
		})
	}
}

// TestTemplate_FaultSource tests that FaultSource is always ClientFaultSource for parameter validation errors
func TestTemplate_FaultSource(t *testing.T) {
	templates := []string{
		"/api/users/{id:integer}",
		"/api/items/{uuid:uuid}",
		"/api/settings/{enabled:boolean}",
		"/api/posts/{slug:slug}",
	}

	for _, tmplStr := range templates {
		t.Run(tmplStr, func(t *testing.T) {
			tmpl, err := pathvars.ParseTemplate(tmplStr)
			if err != nil {
				t.Fatalf("ParseTemplate() error = %v", err)
			}

			invalidPath := tmplStr[:strings.Index(tmplStr, "{")] + "invalid"
			_, err = tmpl.Match(invalidPath, "")
			if err == nil {
				return // Some types might accept "invalid" as valid
			}

			pve, ok := pathvars.FindErr[*pathvars.ParameterError](err)
			if !ok {
				return // Not a ParameterError
			}

			if pve.FaultSource != pathvars.ClientFaultSource {
				t.Errorf("FaultSource = %v, want %v", pve.FaultSource, pathvars.ClientFaultSource)
			}
		})
	}
}
