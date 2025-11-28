package pvtypes

import (
	"fmt"
	"strings"
)

type ConstraintError struct {
	Err           error
	Constraint    string
	FaultSource   FaultSource
	ReceivedValue string
	errString     string
}

func (e ConstraintError) Error() string {
	var meta []any
	var format string

	if e.errString != "" {
		goto end
	}
	meta = []any{ErrInvalidConstraint}
	if e.Err != nil {
		meta = append(meta, e.Err.Error())
	}
	meta = append(meta, fmt.Sprintf(
		"received_value=%s\n"+
			"fault_source=%s",
		e.ReceivedValue,
		e.FaultSource.Slug(),
	))
	format = strings.Repeat("%s\n", len(meta))
	e.errString = fmt.Sprintf(format[:len(format)-1], meta...)
end:
	return e.errString
}

// Unwrap exposes the inner error for errors.Is() and errors.As().
func (e ConstraintError) Unwrap() error {
	return e.Err
}

//func ExtractConstraintError(err error, ce *ConstraintError) (extracted bool) {
//	if !errors.As(err, &ce) {
//		err = NewErr(ErrBug, ErrConstraintValidateDoesNotReturnConstraintError)
//		goto end
//	}
//	extracted = true
//end:
//	return extracted
//}

// Is allows Is(err, ConstraintError{}) to match any instance by type
// and also matches the ErrInvalidConstraint sentinel.
//
//goland:noinspection GoTypeAssertionOnErrors
func (e ConstraintError) Is(target error) (ok bool) {
	// Check for sentinel error
	if target == ErrInvalidConstraint {
		return true
	}
	// Check for type match (value or pointer)
	_, ok = target.(ConstraintError)
	if ok {
		goto end
	}
	_, ok = target.(*ConstraintError)
	if ok {
		goto end
	}
end:
	return ok
}
