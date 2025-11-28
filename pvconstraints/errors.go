// Package pvconstraints/errors defines error values used throughout the pvconstraints package.
// These sentinel errors provide specific error types for different constraint validation failures.
package pvconstraints

import (
	"errors"
)

// Sentinel errors for constraint validation operations.
var (
	// ErrParameterValidationFailed indicates that parameter validation failed against its type or constraints.
	ErrParameterValidationFailed = errors.New("parameter validation failed")

	// ErrInvalidConstraint indicates that a parameter constraint has invalid syntax.
	ErrInvalidConstraint = errors.New("invalid constraint")

	// Length Constraint Errors

	// ErrInvalidLengthRangeMinGreaterThanMax indicates that minimum length is greater than maximum.
	ErrInvalidLengthRangeMinGreaterThanMax = errors.New("minimum length cannot be greater than maximum")

	// ErrInvalidLengthRangeNegativeMin indicates that minimum length is negative.
	ErrInvalidLengthRangeNegativeMin = errors.New("minimum length cannot be negative")

	// ErrExpectedLengthFormat indicates the expected format for length constraints.
	ErrExpectedLengthFormat = errors.New("expected format 'length['min..max]")

	// Date Format Constraint Errors

	// ErrExpectedDateOnlyFormat indicates that only date format (no time) is expected.
	ErrExpectedDateOnlyFormat = errors.New("expected YYYY-MM-DD format")

	// ErrInvalidDateFormat indicates that date format specification is invalid.
	ErrInvalidDateFormat = errors.New("invalid date format")

	// ErrInvalidDateFormatSpec indicates that date format specification is invalid.
	ErrInvalidDateFormatSpec = errors.New("invalid date format")

	ErrDateLessThanMinimum = errors.New("date must be greater than or equal to minimum")

	ErrDateGreaterThanMaximum = errors.New("date must be less than or equal to maximum")

	// ErrStringFormatOnlySupportsIDFormats indicates that string format constraint only supports ulid, ksuid, nanoid.
	ErrStringFormatOnlySupportsIDFormats = errors.New("string format constraint only supports ulid, ksuid, nanoid")

	// ErrFormatConstraintUnsupportedDataType indicates that format constraint only supports date, uuid, and string data types.
	ErrFormatConstraintUnsupportedDataType = errors.New("format constraint only supports date, uuid, and string data types")

	// ErrMoreSegmentsThanFormat indicates that value has more segments than format specification.
	ErrMoreSegmentsThanFormat = errors.New("more segments in value than in format")

	// ErrFailedToBuildPartialLayout indicates that building partial date layout failed.
	ErrFailedToBuildPartialLayout = errors.New("failed to build partial layout")

	// ErrNoValidDateTimeTokens indicates that no valid date/time tokens were found in format.
	ErrNoValidDateTimeTokens = errors.New("no valid date/time tokens found in format")

	// ErrAmbiguousMMToken indicates that 'mm' token is ambiguous.
	ErrAmbiguousMMToken = errors.New("ambiguous 'mm' token - use 'ii' for minutes or add other tokens for context")

	// Enum Constraint Errors

	// ErrEnumValueIsEmpty indicates that an enum constraint value is empty.
	ErrEnumValueIsEmpty = errors.New("enum constraint value is empty")

	// ErrInvalidEnumConstraint indicates that enum constraint syntax is invalid.
	ErrInvalidEnumConstraint = errors.New("invalid enum constraint")

	// Range Constraint Errors

	// ErrExpectedRangeFormat indicates the expected format for range constraints.
	ErrExpectedRangeFormat = errors.New("expected format 'range[min..max]")

	// ErrInvalidRangeConstraint indicates that range constraint syntax is invalid.
	ErrInvalidRangeConstraint = errors.New("invalid range constraint")

	// ErrInvalidRangeValue indicates that a range value is invalid.
	ErrInvalidRangeValue = errors.New("invalid range value")

	// ErrInvalidMinimumValue indicates that minimum value is invalid.
	ErrInvalidMinimumValue = errors.New("invalid minimum value")

	// ErrInvalidMaximumValue indicates that maximum value is invalid.
	ErrInvalidMaximumValue = errors.New("invalid maximum value")

	// ErrInvalidMinMaxValue indicates that minimum value cannot be less than maximum value.
	ErrInvalidMinMaxValue = errors.New("minimum value cannot be less than maximum value")

	// ErrInvalidMinMaxDate indicates that minimum date cannot be less than maximum date.
	ErrInvalidMinMaxDate = errors.New("minimum date cannot be less than maximum date")

	// Regex Constraint Errors

	// ErrEmptyRegexPattern indicates that regex pattern is empty.
	ErrEmptyRegexPattern = errors.New("empty regex pattern")

	// ErrInvalidRegexPattern indicates that regex pattern is invalid.
	ErrInvalidRegexPattern = errors.New("invalid regular expression")

	// ErrInvalidRegexConstraint indicates that regex constraint syntax is invalid.
	ErrInvalidRegexConstraint = errors.New("invalid regex constraint")

	// ErrRegexPatternContainsStartAnchor indicates that regex pattern contains ^ start anchor.
	ErrRegexPatternContainsStartAnchor = errors.New("regex pattern contains ^ start anchor")

	// ErrRegexPatternContainsEndAnchor indicates that regex pattern contains $ end anchor.
	ErrRegexPatternContainsEndAnchor = errors.New("regex pattern contains $ end anchor")

	// ErrRegexPatternContainsBothAnchors indicates that regex pattern contains both ^ and $ anchors.
	ErrRegexPatternContainsBothAnchors = errors.New("regex pattern contains both ^ and $ anchors")

	// UUID Format Constraint Errors

	// ErrUnsupportedUUIDFormat indicates that the UUID format is not supported.
	ErrUnsupportedUUIDFormat = errors.New("unsupported UUID format")

	// ErrUUIDVersionOutOfRange1to8 indicates that UUID version must be 1-8.
	ErrUUIDVersionOutOfRange1to8 = errors.New("UUID version must be 1-8")

	// ErrUUIDVersionOutOfRange1to5 indicates that UUID version must be 1-5.
	ErrUUIDVersionOutOfRange1to5 = errors.New("UUID version must be 1-5")

	// ErrUUIDVersionOutOfRange6to8 indicates that UUID version must be 6-8.
	ErrUUIDVersionOutOfRange6to8 = errors.New("UUID version must be 6-8")

	// ErrUUIDVersionMismatch indicates that UUID version does not match expected version.
	ErrUUIDVersionMismatch = errors.New("UUID version mismatch")

	// ErrInvalidUUIDShape indicates that UUID does not have the expected 8-4-4-4-12 format.
	ErrInvalidUUIDShape = errors.New("invalid UUID shape (expected 8-4-4-4-12 format)")

	// ErrInvalidUUIDHexEncoding indicates that UUID contains invalid hex encoding.
	ErrInvalidUUIDHexEncoding = errors.New("invalid hex encoding in UUID")

	// ErrInvalidUUIDVariant indicates that UUID variant is invalid (must be RFC 4122/9562).
	ErrInvalidUUIDVariant = errors.New("invalid UUID variant (must be RFC 4122/9562)")

	// ErrInvalidUUIDVersion indicates that UUID version is invalid (must be 1-8).
	ErrInvalidUUIDVersion = errors.New("invalid UUID version (must be 1-8)")

	// Other ID Format Errors

	// ErrInvalidULIDFormat indicates that value is not a valid ULID.
	ErrInvalidULIDFormat = errors.New("invalid ULID format")

	// ErrInvalidKSUIDFormat indicates that value is not a valid KSUID.
	ErrInvalidKSUIDFormat = errors.New("invalid KSUID format")

	// ErrInvalidNanoIDFormat indicates that value is not a valid NanoID.
	ErrInvalidNanoIDFormat = errors.New("invalid NanoID format")

	// ErrInvalidCUIDFormat indicates that value is not a valid CUID.
	ErrInvalidCUIDFormat = errors.New("invalid CUID format")

	// ErrInvalidSnowflakeFormat indicates that value is not a valid Snowflake ID.
	ErrInvalidSnowflakeFormat = errors.New("invalid Snowflake ID format")

	// ErrSnowflakeTimestampInFuture indicates that Snowflake timestamp is in the future.
	ErrSnowflakeTimestampInFuture = errors.New("Snowflake timestamp is in the future")

	// ErrSnowflakeTimestampNegative indicates that Snowflake timestamp is negative (before epoch).
	ErrSnowflakeTimestampNegative = errors.New("Snowflake timestamp is negative (before epoch)")

	// ErrInvalidSnowflakeEpoch indicates that the custom epoch parameter is invalid.
	ErrInvalidSnowflakeEpoch = errors.New("invalid Snowflake epoch parameter")
)
