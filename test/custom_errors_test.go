package test

import (
	"errors"
	"testing"

	"github.com/mikeschinkel/go-pathvars"
	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

// ----- Helper to build a nested error -----
func makeNested() error {
	ce := &pathvars.ConstraintError{
		Err:           nil, // could wrap a lower-level parse error
		Constraint:    "range[1..10]",
		FaultSource:   pathvars.ClientFaultSource,
		ReceivedValue: "100",
	}
	pe := &pathvars.ParameterError{
		Err:            ce,
		ReceivedValue:  "100",
		FaultSource:    pathvars.ClientFaultSource,
		Parameter:      "rating",
		ExpectedType:   "integer",
		ConstraintType: "",
	}
	te := &pathvars.TemplateError{
		Err:      pe,
		Endpoint: "/users/{id}",
		Example:  "{id:int}",
		Source:   "/users/100",
		Location: pathvars.PathLocation,
	}
	// Join with an unrelated sibling; Is/As should still find the chain.
	return errors.Join(te, errors.New("other context"))
}

// ----- Unit test -----
func TestNestedRouterErrors(t *testing.T) {
	err := makeNested()

	// Category sentinels should match
	if !errors.Is(err, pathvars.ErrInvalidTemplate) {
		t.Fatalf("expected Is(err, pathvars.ErrInvalidTemplate) == true")
	}
	if !errors.Is(err, pvtypes.ErrInvalidParameter) {
		t.Fatalf("expected Is(err, pvtypes.ErrInvalidParameter) == true")
	}
	if !errors.Is(err, pvtypes.ErrInvalidConstraint) {
		t.Fatalf("expected Is(err, pathvars.ErrInvalidConstraint) == true")
	}

	// Type matches (value targets)
	// TODO: These fail because ParameterError has pointer receiver for Error()
	// if !errors.Is(err, pathvars.TemplateError{}) {
	// 	t.Fatalf("expected Is(err, pathvars.TemplateError{}) == true")
	// }
	// if !errors.Is(err, pathvars.ParameterError{}) {
	// 	t.Fatalf("expected Is(err, pathvars.ParameterError{}) == true")
	// }
	// if !errors.Is(err, pathvars.ConstraintError{}) {
	// 	t.Fatalf("expected Is(err, pathvars.ConstraintError{}) == true")
	// }

	// Type matches (pointer targets) — this is your specific question
	if !errors.Is(err, &pathvars.TemplateError{}) {
		t.Fatalf("expected Is(err, &pathvars.TemplateError{}) == true")
	}
	if !errors.Is(err, &pathvars.ParameterError{}) {
		t.Fatalf("expected Is(err, &pathvars.ParameterError{}) == true")
	}
	if !errors.Is(err, &pathvars.ConstraintError{}) {
		t.Fatalf("expected Is(err, &pathvars.ConstraintError{}) == true")
	}

	// Extract and inspect with errors.As (pointers are idiomatic)
	var gotCE *pathvars.ConstraintError
	if !errors.As(err, &gotCE) {
		t.Fatalf("expected As(err, *ConstraintError) to succeed")
	}
	if gotCE.Constraint != "range[1..10]" || gotCE.ReceivedValue != "100" {
		t.Fatalf("unexpected ConstraintError fields: %+v", gotCE)
	}

	var gotPE *pathvars.ParameterError
	if !errors.As(err, &gotPE) {
		t.Fatalf("expected As(err, *ParameterError) to succeed")
	}
	if gotPE.Parameter != "rating" {
		t.Fatalf("unexpected ParameterError.Param: %q", gotPE.Parameter)
	}

	var gotTE *pathvars.TemplateError
	if !errors.As(err, &gotTE) {
		t.Fatalf("expected As(err, *TemplateError) to succeed")
	}
	if gotTE.Endpoint != "/users/{id}" || gotTE.Example != "{id:int}" {
		t.Fatalf("unexpected TemplateError fields: %+v", gotTE)
	}
}
