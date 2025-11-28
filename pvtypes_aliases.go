package pathvars

import (
	pvt "github.com/mikeschinkel/go-pathvars/pvtypes"
)

type ParameterError = pvt.ParameterError
type ConstraintError = pvt.ConstraintError

type ConstraintMapKey = pvt.ConstraintMapKey

type Selector pvt.Selector
type Identifier = pvt.Identifier
type Location pvt.Location
type HTTPMethod pvt.HTTPMethod

func Identifiers[S ~string](ss []S) (ids []Identifier) {
	return pvt.Identifiers(ss)
}

type NameSpecProps = pvt.NameSpecProps
type PVNameSpec = pvt.PVNameSpec

func ParseNameSpecProps(ns string) (props *NameSpecProps, err error) {
	return pvt.ParseNameSpecProps(ns)
}

// Default parameter data type constants.
const (
	// DefaultPVDataType is the default data type used when no type is specified.
	DefaultPVDataType = pvt.DefaultPVDataType

	// DefaultPVDataTypeName is the string representation of the default data type.
	DefaultPVDataTypeName = pvt.DefaultPVDataTypeName
)

// PVDataType represents the enumerated data types supported for parameters.
type PVDataType = pvt.PVDataType

// Supported parameter data types.
const (
	AlphanumericType    = pvt.AlphanumericType
	BooleanType         = pvt.BooleanType
	DateType            = pvt.DateType
	DecimalType         = pvt.DecimalType
	EmailType           = pvt.EmailType
	IdentifierType      = pvt.IdentifierType
	IntegerType         = pvt.IntegerType
	RealType            = pvt.RealType
	SlugType            = pvt.SlugType
	StringType          = pvt.StringType
	UUIDType            = pvt.UUIDType
	UnspecifiedDataType = pvt.UnspecifiedDataType
)

// PVDataTypeSlug represents the string name of a parameter data type.
type PVDataTypeSlug = pvt.PVDataTypeSlug

// String names for parameter data types.
const (
	AlphanumTypeSlug     = pvt.AlphanumTypeSlug // Accepted alternate for "alphanumeric"
	AlphanumericTypeSlug = pvt.AlphanumericTypeSlug
	BoolTypeSlug         = pvt.BoolTypeSlug // Accepted alternate for "boolean"
	BooleanTypeSlug      = pvt.BooleanTypeSlug
	DateTypeSlug         = pvt.DateTypeSlug
	DecimalTypeSlug      = pvt.DecimalTypeSlug
	EmailTypeSlug        = pvt.EmailTypeSlug
	IdentifierTypeSlug   = pvt.IdentifierTypeSlug
	IntTypeSlug          = pvt.IntTypeSlug // Accepted alternate for "integer"
	IntegerTypeSlug      = pvt.IntegerTypeSlug
	InvalidTypeSlug      = pvt.InvalidTypeSlug
	RealTypeSlug         = pvt.RealTypeSlug
	SlugTypeSlug         = pvt.SlugTypeSlug
	StringTypeSlug       = pvt.StringTypeSlug
	UUIDTypeSlug         = pvt.UUIDTypeSlug
)

// ParsePVDataType converts a string type name to a PVDataType enum value.
// Returns an error if the type name is not recognized.
func ParsePVDataType(slug string) (dt PVDataType, err error) {
	return pvt.ParsePVDataType(slug)
}

// GetDataType returns the data type if matched, or UnspecifiedDataType if not.
func GetDataType(name Identifier) PVDataType {
	return pvt.GetDataType(name)
}

type DataTypeClassifier = pvt.DataTypeClassifier

func GetDataTypeClassifier(dt PVDataType) (v DataTypeClassifier, err error) {
	return pvt.GetDataTypeClassifier(dt)
}

type Parameter = pvt.Parameter

func NewParameter(args ParameterArgs) Parameter {
	return pvt.NewParameter(args)
}

type ParameterArgs = pvt.ParameterArgs

func ParseBraceEnclosed(s string) (_ string, err error) {
	return pvt.ParseBraceEnclosed(s)
}

func ParseParameter(spec string, location LocationType) (p Parameter, err error) {
	return pvt.ParseParameter(spec, location)
}

func ParseParameterDataType(name, typ string) (dt PVDataType, err error) {
	return pvt.ParseParameterDataType(name, typ)
}

// LocationType is an open-ended string type with predefined path and query
type LocationType = pvt.LocationType

const (
	UnspecifiedLocationType = pvt.UnspecifiedLocationType
	PathLocation            = pvt.PathLocation
	QueryLocation           = pvt.QueryLocation
	IrrelevantLocationType  = pvt.IrrelevantLocationType
)

func ParseIdentifier(s string) (id Identifier, err error) {
	return pvt.ParseIdentifier(s)
}

func ParseLeadingIdentifier(s string) (id Identifier, err error) {
	return pvt.ParseLeadingIdentifier(s)
}

type FaultSource = pvt.FaultSource

const (
	UnspecifiedFaultSource = pvt.UnspecifiedFaultSource
	ClientFaultSource      = pvt.ClientFaultSource
	ServerFaultSource      = pvt.ServerFaultSource
)

type ConstraintType = pvt.ConstraintType

const (
	EnumConstraintType     = pvt.EnumConstraintType
	FormatConstraintType   = pvt.FormatConstraintType
	LengthConstraintType   = pvt.LengthConstraintType
	NotEmptyConstraintType = pvt.NotEmptyConstraintType
	RangeConstraintType    = pvt.RangeConstraintType
	RegexConstraintType    = pvt.RegexConstraintType
)

type Constraints = pvt.Constraints

type Constraint = pvt.Constraint

// ParseConstraints parses constraint specifications from a string.
//
// ParseBytes constraint specs like:
//   - notempty
//   - range[0..100]
//   - length[5..50]
//   - regex[^[0-9]+$]
//   - enum[val1,val2,val3]
//   - For dates: format[iso8601], format[yyyy-mm-dd], etc.
//   - Multiple constraints: regex[^[0-9]+$],length[3..10]
func ParseConstraints(spec string, dataType PVDataType) (constraints []Constraint, err error) {
	return pvt.ParseConstraints(spec, dataType)
}

type ConstraintsMap = pvt.ConstraintsMap

type DataTypeAliasMap = pvt.DataTypeAliasMap

func RegisterConstraint(c Constraint) {
	pvt.RegisterConstraint(c)
}

func GetConstraintsMap() ConstraintsMap {
	return pvt.GetConstraintsMap()
}

func GetConstraintMapKey(ct ConstraintType, dtn PVDataTypeSlug) ConstraintMapKey {
	return pvt.GetConstraintMapKey(ct, dtn)
}

func GetConstraint(ct ConstraintType, dt PVDataType) (c Constraint, err error) {
	return pvt.GetConstraint(ct, dt)
}
