package pathvars

import (
	"fmt"
	"strings"
)

type TemplateError struct {
	Err        error
	Endpoint   string
	Example    string
	Source     string
	Suggestion string
	Location   LocationType
	errString  string
}

// extractParameterError retrieves ParameterError from wrapped error chain
func (e *TemplateError) extractParameterError() *ParameterError {
	pe, _ := FindErr[*ParameterError](e.Err)
	return pe
}

// FaultSource returns fault source, delegating to ParameterError if present
func (e *TemplateError) FaultSource() FaultSource {
	if pe := e.extractParameterError(); pe != nil {
		return pe.GetFaultSource()
	}
	return ServerFaultSource // default for template-level errors
}

// Detail returns detail message, delegating to ParameterError if present
func (e *TemplateError) Detail() string {
	if pe := e.extractParameterError(); pe != nil {
		return pe.GetDetail()
	}
	// Template-level errors might have their own detail logic
	return fmt.Sprintf("Template error in endpoint '%s'", e.Endpoint)
}

// Parameter returns parameter name, delegating to ParameterError if present
func (e *TemplateError) Parameter() string {
	if pe := e.extractParameterError(); pe != nil {
		return pe.Parameter
	}
	return ""
}

// ExpectedType returns expected type, delegating to ParameterError if present
func (e *TemplateError) ExpectedType() string {
	if pe := e.extractParameterError(); pe != nil {
		return pe.ExpectedType
	}
	return ""
}

// ReceivedValue returns received value, delegating to ParameterError if present
func (e *TemplateError) ReceivedValue() string {
	if pe := e.extractParameterError(); pe != nil {
		return pe.GetReceivedValue()
	}
	return ""
}

// ConstraintType returns constraint type, delegating to ParameterError if present
func (e *TemplateError) ConstraintType() string {
	if pe := e.extractParameterError(); pe != nil {
		return pe.GetConstraintType()
	}
	return ""
}

// GetSuggestion returns suggestion, using field if set or delegating to ParameterError
func (e *TemplateError) GetSuggestion() string {
	if e.Suggestion != "" {
		return e.Suggestion
	}
	if pe := e.extractParameterError(); pe != nil {
		return e.enhanceSuggestionWithContext(pe.GetSuggestion())
	}
	return ""
}

// enhanceSuggestionWithContext adds template context to parameter-level suggestion
func (e *TemplateError) enhanceSuggestionWithContext(baseSuggestion string) string {
	if e.Example != "" {
		return fmt.Sprintf("%s. Example: %s", baseSuggestion, e.Example)
	}
	return baseSuggestion
}

type TemplateErrorArgs struct {
	Endpoint   string
	Example    string
	Source     string
	Location   LocationType
	Suggestion string
	Parameter  Parameter
}

func NewTemplateError(err error, args TemplateErrorArgs) *TemplateError {
	return &TemplateError{
		Endpoint:   args.Endpoint,
		Example:    args.Example,
		Source:     args.Source,
		Location:   args.Location,
		Suggestion: args.Suggestion,
		Err:        err,
	}
}

func (e *TemplateError) Error() string {
	var meta []any
	var format string

	if e.errString != "" {
		goto end
	}
	meta = []any{ErrInvalidTemplate}
	meta = append(meta, fmt.Sprintf(
		"endpoint=%s\n"+
			"source=%s\n"+
			"example=%s\n"+
			"location=%s",
		e.Endpoint,
		e.Source,
		e.Example,
		e.Location,
	))
	if e.Suggestion != "" {
		meta = append(meta, fmt.Sprintf("suggestion=%s", e.Suggestion))
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
func (e *TemplateError) Unwrap() error {
	return e.Err
}

// Is allows Is(err, TemplateError{}) to match any instance by type
// and also matches the ErrInvalidTemplate sentinel.
//
//goland:noinspection GoTypeAssertionOnErrors
func (e *TemplateError) Is(target error) (ok bool) {
	// Check for sentinel error
	if target == ErrInvalidTemplate {
		return true
	}
	// Check for type match (value or pointer)
	//_, ok = target.(TemplateError)
	//if ok {
	//	goto end
	//}
	_, ok = target.(*TemplateError)
	if ok {
		goto end
	}
end:
	return ok
}
