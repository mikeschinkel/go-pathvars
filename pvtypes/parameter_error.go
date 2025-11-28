package pvtypes

import (
	"errors"
	"fmt"
	"strings"
)

type ParameterError struct {
	Err            error
	ReceivedValue  string
	FaultSource    FaultSource
	Parameter      string
	ExpectedType   string
	Detail         string
	ConstraintType string
	errString      string
}

func newParameterError(p *Parameter, value string, err error) *ParameterError {
	return &ParameterError{
		Err:           err,
		ReceivedValue: value,
		Parameter:     string(p.Name),
		ExpectedType:  string(p.dataType.Slug()),
	}
}

func (e *ParameterError) Error() string {
	var meta []any
	var format string

	if e.errString != "" {
		goto end
	}
	meta = []any{ErrInvalidParameter}
	meta = append(meta, fmt.Sprintf(
		"parameter=%s\n"+
			"expected_type=%s\n"+
			"detail=%s",
		e.Parameter,
		e.ExpectedType,
		e.Detail,
	))
	if !errors.Is(e.Err, (*ConstraintError)(nil)) {
		// If we don't have a constraint error, fill these in.
		// (If we did have one, we'll already have them.)
		meta = append(meta, fmt.Sprintf(
			"received_value=%s\n"+
				"fault_source=%s",
			e.ReceivedValue,
			e.FaultSource.Slug(),
		))
	}
	if e.Err != nil {
		meta = append(meta, e.Err.Error())
	}
	format = strings.Repeat("%s\n", len(meta))
	e.errString = fmt.Sprintf(format[:len(format)-1], meta...)
end:
	return e.errString
}

// Unwrap exposes the inner error for errors.Is() and errors.As().
func (e *ParameterError) Unwrap() error {
	return e.Err
}

func ExtractParameterError(err error, pe *ParameterError) (extracted bool) {
	if !errors.As(err, &pe) {
		err = NewErr(ErrBug, ErrParameterValidateDoesNotReturnParameterError)
		goto end
	}
	extracted = true
end:
	return extracted
}

// Is allows Is(err, ParameterError{}) to match any instance by type
// and also matches the ErrInvalidParameter sentinel.
//
//goland:noinspection GoTypeAssertionOnErrors
func (e *ParameterError) Is(target error) (ok bool) {
	// Check for sentinel error
	if target == ErrInvalidParameter {
		return true
	}
	// Check for type match (value or pointer)
	_, ok = target.(*ParameterError)
	if ok {
		goto end
	}
end:
	return ok
}

// extractConstraintError retrieves ConstraintError from wrapped error chain
func (e *ParameterError) extractConstraintError() *ConstraintError {
	ce, _ := FindErr[*ConstraintError](e.Err)
	return ce
}

// GetFaultSource returns fault source, delegating to ConstraintError if present
func (e *ParameterError) GetFaultSource() FaultSource {
	if ce := e.extractConstraintError(); ce != nil {
		return ce.FaultSource
	}
	return e.FaultSource
}

// GetReceivedValue returns received value, delegating to ConstraintError if present
func (e *ParameterError) GetReceivedValue() string {
	if ce := e.extractConstraintError(); ce != nil {
		return ce.ReceivedValue
	}
	return e.ReceivedValue
}

// GetConstraintType returns the constraint type if this is a constraint error
func (e *ParameterError) GetConstraintType() string {
	if ce := e.extractConstraintError(); ce != nil {
		return e.ConstraintType
	}
	return ""
}

// GetDetail returns the detail message based on error type
func (e *ParameterError) GetDetail() string {
	if e.Detail != "" {
		return e.Detail
	}
	if ce := e.extractConstraintError(); ce != nil {
		return e.buildConstraintDetail(ce)
	}
	return e.buildTypeErrorDetail()
}

// GetSuggestion returns the suggestion based on error type
func (e *ParameterError) GetSuggestion() string {
	if ce := e.extractConstraintError(); ce != nil {
		return e.buildConstraintSuggestion(ce)
	}
	return e.buildTypeErrorSuggestion()
}

// buildConstraintDetail builds detail message for constraint validation errors
func (e *ParameterError) buildConstraintDetail(ce *ConstraintError) string {
	return fmt.Sprintf(
		"Parameter '%s' with value '%s' failed constraint validation: %s",
		e.Parameter,
		ce.ReceivedValue,
		ce.Err.Error(),
	)
}

// buildTypeErrorDetail builds detail message for type validation errors
func (e *ParameterError) buildTypeErrorDetail() string {
	return fmt.Sprintf(
		"Parameter '%s' expected type '%s' but received invalid value '%s'",
		e.Parameter,
		e.ExpectedType,
		e.ReceivedValue,
	)
}

// buildConstraintSuggestion builds suggestion for constraint validation errors
func (e *ParameterError) buildConstraintSuggestion(ce *ConstraintError) string {
	return fmt.Sprintf(
		"Ensure parameter '%s' satisfies the constraint: %s",
		e.Parameter,
		e.ConstraintType,
	)
}

// buildTypeErrorSuggestion builds suggestion for type validation errors
func (e *ParameterError) buildTypeErrorSuggestion() string {
	return fmt.Sprintf(
		"Provide a valid %s value for parameter '%s'",
		e.ExpectedType,
		e.Parameter,
	)
}
