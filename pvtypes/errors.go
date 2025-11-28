// Package pathvars.Errors defines error values used throughout the pathvars package.
// These sentinel errors provide specific error types for different failure modes
// during path template parsing, route compilation, and request matching.
package pvtypes

import (
	"errors"
)

// Sentinel errors for various pathvars operations.
var (
	// ErrInvalidParameter indicates that a parameter is invalid
	ErrInvalidParameter = errors.New("invalid parameter")

	// ErrInvalidParameterSyntax indicates that a parameter syntax is malformed.
	ErrInvalidParameterSyntax = errors.New("invalid parameter syntax")

	// ErrInvalidParameterType indicates that an unknown or unsupported parameter type was specified.
	ErrInvalidParameterType = errors.New("invalid parameter value for type")

	// ErrParameterValidationFailed indicates that parameter validation failed against its type or constraints.
	ErrParameterValidationFailed = errors.New("parameter validation failed")

	// ErrUnknownConstraintType indicates that an unknown constraint type was specified.
	ErrUnknownConstraintType = errors.New("unknown constraint type")

	// ErrInvalidSyntax indicates that constraint or parameter syntax is malformed.
	ErrInvalidSyntax = errors.New("invalid syntax")

	// ErrParseFailed indicates that parsing of a constraint or parameter failed.
	ErrParseFailed = errors.New("parse failed")

	ErrInvalidNameSpec = errors.New("invalid name spec")

	ErrNameSpecNameCannotBeEmpty = errors.New("name spec name cannot be empty")

	ErrValueCannotBeEmpty = errors.New("value cannot be empty")

	ErrWhatNameSpecMustContain = errors.New("name spec must begin with a valid identifier — containing letters, digits and/or underscores — may then optionally contain an asterisk ('*': for multisegment), a question mark ('?': for optional), and if optional then optionally a default value, e.g. 'category?uncategorized'")

	ErrParameterDataTypeValidationFailed   = errors.New("parameter data type validation failed")
	ErrConstraintValidationFailed          = errors.New("constraint validation failed")
	ErrParameterConstraintValidationFailed = errors.New("parameter constraint validation failed")

	// Parameter Errors

	// ErrParameterLocationNotSpecified indicates that parameter location was not specified.
	ErrParameterLocationNotSpecified = errors.New("parameter location not specified")

	// ErrInvalidIntegerFormat indicates that value is not a valid integer.
	ErrInvalidIntegerFormat = errors.New("invalid integer format")

	// ErrInvalidDecimalFormat indicates that value is not a valid decimal.
	ErrInvalidDecimalFormat = errors.New("invalid decimal format")

	// ErrInvalidRealFormat indicates that value is not a valid real number.
	ErrInvalidRealFormat = errors.New("invalid real number format")

	// ErrUnsupportedDataType indicates that the data type is not supported.
	ErrUnsupportedDataType = errors.New("unsupported data type")

	// ErrInvalidIdentifierFormat indicates that value does not conform to identifier format.
	ErrInvalidIdentifierFormat = errors.New("must start with lowercase letter, followed by lowercase letters, digits, or underscores")

	// ErrInvalidUUIDFormatBasic indicates that value is not a valid UUID (basic validation).
	ErrInvalidUUIDFormatBasic = errors.New("invalid UUID format (expected 8-4-4-4-12 hex digits)")

	// ErrInvalidAlphanumericFormat indicates that value must contain only letters and digits.
	ErrInvalidAlphanumericFormat = errors.New("value must contain only letters and digits")

	// ErrInvalidSlugFormat indicates that value does not conform to slug format.
	ErrInvalidSlugFormat = errors.New("must be lowercase letters/digits with optional hyphens between segments")

	// ErrInvalidBooleanFormat indicates that boolean value must be 'true' or 'false'.
	ErrInvalidBooleanFormat = errors.New("boolean value must be exactly 'true' or 'false'")

	// ErrInvalidEmailFormat indicates that value is not a valid email format.
	ErrInvalidEmailFormat = errors.New("invalid email format")

	// Length Constraint Errors

	// ErrInvalidLengthRangeMinGreaterThanMax indicates that minimum length is greater than maximum.
	ErrInvalidLengthRangeMinGreaterThanMax = errors.New("invalid length range (min > max)")

	// ErrInvalidLengthRangeNegativeMin indicates that minimum length is negative.
	ErrInvalidLengthRangeNegativeMin = errors.New("invalid length range (min < 0)")

	// Constraint Type Errors

	// ErrInvalidConstraintTypeCharacter indicates that a constraint type contains an invalid character.
	ErrInvalidConstraintTypeCharacter = errors.New("invalid constraint type character")

	ErrInvalidFaultSource = errors.New("invalid fault source")

	// ErrInvalidConstraint indicates that a parameter constraint has invalid syntax.
	ErrInvalidConstraint = errors.New("invalid constraint syntax")
)

var (
	ErrFailedToMatchDateFormat = errors.New("failed to match date format")
	ErrFailedToParseDateFormat = errors.New("failed to parse date format")
	ErrInvalidYearInDate       = errors.New("invalid year in date")
	ErrInvalidMonthInDate      = errors.New("invalid month in date")
	ErrInvalidDayInDate        = errors.New("invalid day in date")
)

var (
	ErrBug                                            = errors.New("a bug exists that needs to be fixed")
	ErrParameterValidateDoesNotReturnParameterError   = errors.New("Parameter.Validate() does not return a *pathvars.ParameterError when returned err!=nil")
	ErrConstraintValidateDoesNotReturnConstraintError = errors.New("Constraint.Validate() does not return a *pathvars.ConstraintError when returned err!=nil")
)

var (
	ErrInvalidDateFormat = errors.New("invalid date format")
)

var (
	ErrDataTypeHasNoRegisteredClassifier = errors.New("data type has no registered classifier")
	ErrDataTypeClassifiersNotRegistered  = errors.New("data type classifiers not registered")
)

var ErrMustBeginWithLetterOrUnderscore = errors.New("must begin with letter or underscore")

var ErrMustOnlyContainLettersDigitsAndOrUnderscores = errors.New("must only contain letters, digits, and/or underscores")

var ErrIdentifierCannotBeEmpty = errors.New("identifier can not be empty")
