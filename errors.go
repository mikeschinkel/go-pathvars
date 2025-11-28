// Package pathvars/errors defines error values used throughout the pathvars package.
// These sentinel errors provide specific error types for different failure modes
// during path template parsing, route compilation, and request matching.
package pathvars

import (
	"errors"
)

// Sentinel errors for various pathvars operations.
var (
	// ErrInvalidParameterSyntax indicates that a parameter syntax is malformed.
	ErrInvalidParameterSyntax = errors.New("invalid parameter syntax")

	// ErrInvalidParameterType indicates that an unknown or unsupported parameter type was specified.
	ErrInvalidParameterType = errors.New("invalid parameter type")

	// ErrParseFailed indicates that parsing of a constraint or parameter failed.
	ErrParseFailed                  = errors.New("parse failed")
	ErrFailedToExtractParameterSpec = errors.New("failed to extract parameter spec")
	ErrFailedToParseParameter       = errors.New("failed to parse parameter")
	// ErrConstraintValidationFailed indicates that constraint validation failed.
	ErrConstraintValidationFailed = errors.New("constraint validation failed")

	// Template/Parser Errors

	// ErrInvalidTemplate indicates that a path template has invalid syntax.
	ErrInvalidTemplate = errors.New("invalid template syntax")

	// ErrEmptyTemplate indicates that the template string is empty.
	ErrEmptyTemplate = errors.New("empty template")

	// ErrUnmatchedClosingBrace indicates an unmatched closing brace in template.
	ErrUnmatchedClosingBrace = errors.New("unmatched closing brace")

	// ErrUnmatchedOpeningBrace indicates an unmatched opening brace in template.
	ErrUnmatchedOpeningBrace = errors.New("unmatched opening brace(s)")

	// ErrMalformedBraces indicates a closing brace before and opening brace
	ErrMalformedBraces = errors.New("malformed brace; '{' must precede '}'")

	// Router Errors

	// ErrNoRouteMatched indicates that no route matched the request.
	ErrNoRouteMatched = errors.New("no route matched the request")

	// ErrNoMatch indicates that no route matched the incoming request.
	ErrNoMatch = errors.New("no matching route")

	// Other Errors

	// ErrParsingDBExtensionFailed indicates that parsing a database extension failed.
	ErrParsingDBExtensionFailed = errors.New("parsing database extension failed")

	// ErrRequiredParameterNotProvided indicates that a required parameter was not provided.
	ErrRequiredParameterNotProvided = errors.New("required parameter not provided")

	// ErrInvalidURLQueryString indicates that the URL query string is invalid.
	ErrInvalidURLQueryString = errors.New("invalid URL query string")

	// ErrParameterNotFoundInValuesMap indicates that a parameter was not found in the values map.
	ErrParameterNotFoundInValuesMap = errors.New("parameter not found in values map")

	// ErrPathParameterNotFoundInValuesMap indicates that a path parameter was not found in the values map.
	ErrPathParameterNotFoundInValuesMap = errors.New("path parameter not found in values map")

	// ErrQueryParameterNotFoundInValuesMap indicates that a query parameter was not found in the values map.
	ErrQueryParameterNotFoundInValuesMap = errors.New("query parameter not found in values map")
)
