// Package pvtypes defines the data types supported for path and query parameters.
// It provides type definitions, name mappings, and parsing functions for validating
// parameter values against specific data types like integers, UUIDs, dates, etc.
package pvtypes

import (
	"fmt"
)

// Default parameter data type constants.
const (
	// DefaultPVDataType is the default data type used when no type is specified.
	DefaultPVDataType = StringType

	// DefaultPVDataTypeName is the string representation of the default data type.
	DefaultPVDataTypeName = StringTypeSlug
)

// PVDataType represents the enumerated data types supported for parameters.
type PVDataType int

// Supported parameter data types.
const (
	// UnspecifiedDataType indicates no data type was specified.
	UnspecifiedDataType PVDataType = iota

	// StringType represents text data with no specific format requirements.
	StringType

	// IntegerType represents whole number values (positive, negative, or zero).
	IntegerType

	// RealType represents floating-point numeric values.
	RealType

	// DecimalType represents decimal numeric values with precise fractional parts.
	DecimalType

	// IdentifierType represents programming-style identifiers (lowercase, alphanumeric with underscores).
	IdentifierType

	// DateType represents date/time values that can be validated against various formats.
	DateType

	// UUIDType represents Universally Unique Identifier values.
	UUIDType

	// AlphanumericType represents values containing only letters and digits.
	AlphanumericType

	// SlugType represents URL-safe slug values (lowercase, hyphen-separated).
	SlugType

	// BooleanType represents true/false values.
	BooleanType

	// EmailType represents email address values.
	EmailType
)

// PVDataTypeSlug represents the string name of a parameter data type.
type PVDataTypeSlug string

// String names for parameter data types.
const (
	// InvalidTypeSlug indicates an unrecognized type name.
	InvalidTypeSlug PVDataTypeSlug = "invalid"

	// StringTypeSlug is the string representation of StringType.
	StringTypeSlug PVDataTypeSlug = "string"

	// IntegerTypeSlug is the string representation of IntegerType.
	IntegerTypeSlug PVDataTypeSlug = "integer"

	// IntTypeSlug is an accepted alternate name for IntegerType.
	IntTypeSlug PVDataTypeSlug = "int" // Accepted alternate for "integer"

	// DecimalTypeSlug is the string representation of DecimalType.
	DecimalTypeSlug PVDataTypeSlug = "decimal"

	// RealTypeSlug is the string representation of RealType.
	RealTypeSlug PVDataTypeSlug = "real"

	// IdentifierTypeSlug is the string representation of IdentifierType.
	IdentifierTypeSlug PVDataTypeSlug = "identifier"

	// DateTypeSlug is the string representation of DateType.
	DateTypeSlug PVDataTypeSlug = "date"

	// UUIDTypeSlug is the string representation of UUIDType.
	UUIDTypeSlug PVDataTypeSlug = "uuid"

	// AlphanumericTypeSlug is the string representation of AlphanumericType.
	AlphanumericTypeSlug PVDataTypeSlug = "alphanumeric"

	// AlphanumTypeSlug is an accepted alternate name for AlphanumericType.
	AlphanumTypeSlug PVDataTypeSlug = "alphanum" // Accepted alternate for "alphanumeric"

	// SlugTypeSlug is the string representation of SlugType.
	SlugTypeSlug PVDataTypeSlug = "slug"

	// BooleanTypeSlug is the string representation of BooleanType.
	BooleanTypeSlug PVDataTypeSlug = "boolean"

	// BoolTypeSlug is an accepted alternate name for BooleanType.
	BoolTypeSlug PVDataTypeSlug = "bool" // Accepted alternate for "boolean"

	// EmailTypeSlug is the string representation of EmailType.
	EmailTypeSlug PVDataTypeSlug = "email"
)

func (dt PVDataType) WithIndefiniteArticle() (wia string) {
	classifier, err := GetDataTypeClassifier(dt)
	if err != nil {
		wia = "a"
		goto end
	}
	wia = fmt.Sprintf("%s %s", classifier.IndefiniteArticle(), classifier.Slug())
end:
	return wia
}

// Slug returns the canonical lowercase string name for this data type	.
func (dt PVDataType) Slug() (slug PVDataTypeSlug) {
	classifier, err := GetDataTypeClassifier(dt)
	if err != nil {
		slug = InvalidTypeSlug
		goto end
	}
	slug = classifier.Slug()
end:
	return slug
}

func (dt PVDataType) Example() (ex any) {
	classifier, err := GetDataTypeClassifier(dt)
	if err != nil {
		ex = "Unspecified has no example"
		goto end
	}
	ex = classifier.Example()
end:
	return ex
}

// ParsePVDataType converts a string type name to a PVDataType enum value.
// Returns an error if the type name is not recognized.
func ParsePVDataType(slug string) (dt PVDataType, err error) {
	dt = FindDataType(PVDataTypeSlug(slug))
	if dt == UnspecifiedDataType {
		err = NewErr(
			ErrInvalidParameterType,
			ErrUnsupportedDataType,
			"data_type", slug,
		)
	}
	return dt, err
}

// GetDataType returns the data type if matched, or UnspecifiedDataType if not.
func GetDataType(name Identifier) PVDataType {
	dataType, err := ParsePVDataType(string(name))
	if err != nil {
		return UnspecifiedDataType
	}
	return dataType
}
