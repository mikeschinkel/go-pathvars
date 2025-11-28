package test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
)

type ep struct {
	method pathvars.HTTPMethod
	path   pathvars.Template
}

func TestPathVars(t *testing.T) {
	api := []ep{
		{
			method: "GET",
			path:   "/foos/{foo:string}/bars/{bars:int}",
		},
	}
	tests := []struct {
		method   string
		path     string
		expected pathvars.MatchResult
		wantErr  bool
	}{
		{
			method: "GET",
			path:   "/foos/myfoo/bars/1",
			expected: pathvars.NewMatchResult(&pathvars.Route{Index: 0}, newValuesMap(
				"foo", "myfoo",
				"bars", "1",
			)),
		},
		{
			method:  "POST",
			path:    "/foos/myfoo/bars/1",
			wantErr: true,
		},
		{
			method:  "GET",
			path:    "/foos/1/bars/mybar",
			wantErr: true,
		},
		{
			method:  "GET",
			path:    "/foos/myfoo/bars/1/extra",
			wantErr: true,
		},
		{
			method:  "POST",
			path:    "/foos/myfoo/bars/1",
			wantErr: true,
		},
		{
			method:  "GET",
			path:    "/not/a/match/at/all",
			wantErr: true,
		},
	}
	router := pathvars.NewRouter()
	for _, ep := range api {
		err := router.AddRoute(ep.method, ep.path, nil)
		if err != nil {
			t.Error(err)
		}
	}

	for _, tt := range tests {
		name := fmt.Sprintf("%s %s", tt.method, tt.path)
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			result, err := router.Match(req)
			// t.Log("Result:", result)

			hasError := err != nil

			switch {
			case tt.wantErr && !hasError:
				t.Errorf("Expected error but got success")
				return
			case !tt.wantErr && hasError:
				t.Errorf("Expected success but got error: %v", err)
				return
			}

			// Validate expected index
			if result.Index != tt.expected.Index {
				t.Errorf("Expected index %d, got %d", tt.expected.Index, result.Index)
			}

			// Validate expected parameters
			if !tt.expected.HasVars() {
				return
			}

			for expectedParam, expectedValue := range tt.expected.ValuesMap().Iterator() {
				actualValue, found := result.GetValue(expectedParam)
				if !found {
					t.Errorf("Expected parameter %q not found", expectedParam)
					continue
				}
				if actualValue != expectedValue {
					t.Errorf("Expected parameter %q = %q, got %q", expectedParam, expectedValue, actualValue)
				}
			}

			// Check for unexpected parameters
			result.ForEachVar(func(name pathvars.Identifier, value any) bool {
				_, expected := tt.expected.GetValue(name)
				if !expected {
					t.Errorf("Unexpected parameter %q = %q", name, value)
				}
				return true
			})
		})
	}
}
