// Package pvtypes/parameter defines parameter types and parsing functionality
// for path and query parameters. Parameters can have data types, constraints,
// default values, and can be either required or optional.
package pvtypes

import (
	"errors"
	"fmt"
	"strings"
)

// Parameter represents a path or query parameter with its type, constraints, and configuration.
// Parameters can be required or optional, have default values, and span multiple path segments.
type Parameter struct {
	// location indicates whether this is a path, query, body, or header parameter.
	location LocationType

	// dataType specifies the expected data type for validation.
	dataType PVDataType

	// constraints contains validation rules applied to parameter values.
	constraints []Constraint

	// position indicates the parameter's position among path parameters for regex capture groups.
	position int

	// original stores the original parameter specification string for reference.
	original string

	nameProps
}

func (p Parameter) Location() LocationType {
	return p.location
}

func (p Parameter) WithLocation(lt LocationType) Parameter {
	p.location = lt
	return p
}

func (p Parameter) Constraints() []Constraint {
	return p.constraints
}

func (p Parameter) WithConstraints(cs []Constraint) Parameter {
	p.constraints = cs
	return p
}

func (p Parameter) Position() int {
	return p.position
}

func (p Parameter) WithPosition(pos int) Parameter {
	p.position = pos
	return p
}

type nameProps = NameSpecProps

// NewParameter creates a new Parameter instance with the specified configuration.
func NewParameter(args ParameterArgs) Parameter {
	return Parameter{
		location:    args.Location,
		dataType:    args.DataType,
		constraints: args.Constraints,
		position:    args.Position,
		original:    args.Original,
		nameProps:   args.NameProps,
	}
}

// ParameterArgs contains arguments for creating a Parameter instance.
// This struct allows for easy parameter construction with named fields.
type ParameterArgs struct {
	// NameProps contains properties defined in the name
	NameProps NameSpecProps

	// Location indicates if this is a path, query, body, or header parameter.
	Location LocationType

	// DataType specifies the expected data type.
	DataType PVDataType

	// Constraints contains validation rules.
	Constraints []Constraint

	// Position indicates the parameter's position for regex matching.
	Position int

	// Original stores the original parameter specification.
	Original string
}

func isBraceEnclosed(s string) (enclosed bool) {
	switch {
	case len(s) < 2:
		goto end
	case s[0] != '{':
		goto end
	case s[len(s)-1] != '}':
		goto end
	default:
		enclosed = true
	}
end:
	return enclosed
}

var ErrNotBraceEnclosed = errors.New("not brace enclosed with '{' and '}'")

func ParseBraceEnclosed(s string) (_ string, err error) {
	// Remove braces
	if !isBraceEnclosed(s) {
		err = NewErr(ErrNotBraceEnclosed, "value", s)
		goto end
	}

	s = s[1 : len(s)-1]
	if s == "" {
		err = NewErr(ErrValueCannotBeEmpty, "value", s)
		goto end
	}
end:
	return s, err
}

// ParseParameter parses a parameter specification like {id:int:range[1..100]} or {date*:date:yyyy/mm/dd}.
// Also supports optional parameters: {name?:type} or {name?default:type:constraints}.
// The position parameter indicates the parameter's position for regex capture group ordering.
func ParseParameter(spec string, location LocationType) (p Parameter, err error) {
	var content string
	var parts []string
	var dataType PVDataType
	var constraints []Constraint
	var props *NameSpecProps

	// ParseBytes the {name:type:constraints} or {name*:type:constraints} format
	// Return Parameter object with parsed components

	if location == "" {
		err = NewErr(
			ErrInvalidParameter,
			ErrParameterLocationNotSpecified,
		)
		goto end
	}

	content, err = ParseBraceEnclosed(spec)
	if err != nil {
		err = WithErr(err,
			ErrInvalidParameter,
			ErrInvalidParameterSyntax,
		)
		goto end
	}

	// Split by colon, but only split on first two colons to handle constraints with colons
	// e.g., "name:type:hh:mm:ss" -> ["name", "type", "hh:mm:ss"]
	parts = strings.SplitN(content, ":", 3)

	// ParseBytes the first part which may contain name, optional marker (?), and default value
	// Possible formats:
	// - "name" -> required parameter
	// - "name?" -> optional parameter, no default
	// - "name?default" -> optional parameter with default value
	// - "name*" -> multi-segment required parameter
	// - "name*?" -> multi-segment optional parameter, no default
	// - "name*?default" -> multi-segment optional parameter with default
	props, err = ParseNameSpecProps(parts[0])
	if err != nil {
		err = WithErr(err,
			"content", content,
		)
		goto end
	}
	if props == nil {
		// Added this here because Goland flags props.DataType as possibly being null. I
		// don't see how it could be possible, but maybe Goland knows something I don't?
		panic(fmt.Sprintf("NameSpecProps are nil when err is also nil; spec=%s", spec))
	}
	switch {
	case len(parts) > 1:
		// Pattern: {name:type} or {name:type:constraint} -> explicit type provided
		dataType, err = ParseParameterDataType(string(props.Name), parts[1])
		if err != nil {
			// parameter name and data type already added by ParseParameterDataType()
			goto end
		}
	case props.DataType != nil:
		// Pattern: {name} -> name matched a data type (inferred)
		dataType = *props.DataType
	case len(parts) == 1:
		// Pattern: {name} -> name not a data type, use default
		dataType = DefaultPVDataType
	}

	// Get constraints (optional)
	if len(parts) > 2 && parts[2] != "" {
		constraints, err = ParseConstraints(parts[2], dataType)
		if err != nil {
			err = WithErr(err,
				"data_type", dataType.Slug(),
				"constraint_spec", parts[2],
			)
			goto end
		}
	}

	p = Parameter{
		nameProps:   *props,
		location:    location,
		dataType:    dataType,
		constraints: constraints,
		original:    spec,
	}

	// Validate default value if provided
	if p.DefaultValue != nil {
		err = p.Validate(*p.DefaultValue)
	}

end:
	if err != nil {
		p = Parameter{}
		err = WithErr(err,
			"parameter_spec", spec,
			"parameter_location", location,
		)
	}
	return p, err
}

func ParseParameterDataType(name, typ string) (dt PVDataType, err error) {
	// Determine data type based on syntax
	switch {
	case typ != "":
		// Pattern: {name:type} or  {name:type:} or {name:type:constraint} -> explicit type
		dt, err = ParsePVDataType(typ)
		if err != nil {
			err = WithErr(err,
				"parameter_name", name,
				//"data_type", typ,  Already added by ParsePVDataType()
			)
			goto end
		}
	default:
		// Pattern: {name:} or {name::} or {name::constraint} -> infer type from name
		inferredType := GetDataType(Identifier(name))
		if inferredType != UnspecifiedDataType {
			dt = inferredType
			goto end
		}
		dt = DefaultPVDataType
	}
end:
	return dt, err
}

// DataType returns the parameter's data type.
func (p Parameter) DataType() PVDataType {
	return p.dataType
}

func (p Parameter) DataTypeSlug() PVDataTypeSlug {
	return p.dataType.Slug()
}

func (p Parameter) Example(err error, args *ExampleArgs) (example any) {
	var pe *ParameterError
	if args == nil {
		args = &ExampleArgs{}
	}
	// If we have an error with a specific constraint, use that constraint's example
	if errors.As(err, &pe) && pe.ConstraintType != "" {
		for _, c := range p.constraints {
			if c.String() != pe.ConstraintType {
				continue
			}
			example = c.Example(err)
			if example == nil {
				continue
			}
			goto end
		}
	}

	// For suggestion text with type errors (not constraint errors), use data type example
	// This ensures format[v4] failures show generic UUID (v1) in text, not constraint-specific v4
	if args.SuggestionType == DataTypeSuggestion {
		// Check if this is a type error (no constraint error)
		if pe == nil || pe.ConstraintType == "" {
			example = p.dataType.Example()
			goto end
		}
	}

	// Check if any constraint provides an example (prefer type-validating constraints first)
	for _, c := range p.constraints {
		if !c.ValidatesType() {
			continue
		}
		example = c.Example(nil)
		if example != nil {
			goto end
		}
	}
	// Check other constraints (e.g., range, length) for examples
	for _, c := range p.constraints {
		example = c.Example(nil)
		if example != nil {
			goto end
		}
	}
	example = p.dataType.Example()
end:
	// Fall back to data type example
	return example
}

func (p Parameter) ValidateForDataType(value string) (err error) {
	var newer, v DataTypeClassifier
	if p.Optional && value == "" {
		goto end
	}
	newer, err = GetDataTypeClassifier(p.dataType)
	if err != nil {
		goto end
	}
	v = newer.MakeNew(&DataTypeClassifierArgs{
		MultiSegment: p.MultiSegment,
	})
	err = v.Validate(value)
end:
	if err != nil {
		pe := p.createDataTypeViolationError(value)
		// Join the Classifier error with the ParameterError
		err = errors.Join(pe, err)
	}
	return err
}

// ConstraintValidatesType returns true if any of this parameter's constraints
// perform their own type validation, allowing them to replace default data type validation.
func (p Parameter) ConstraintValidatesType() (validates bool) {
	for _, c := range p.constraints {
		if !c.ValidatesType() {
			continue
		}
		validates = true
		goto end
	}
end:
	return validates
}

func (p Parameter) Validate(value string) (err error) {
	// Only validate type upfront if no constraint handles type validation
	if !p.ConstraintValidatesType() {
		err = p.ValidateForDataType(value)
		if err != nil {
			goto end
		}
	}
	err = p.ValidateConstraints(value)
end:
	if err != nil {
		err = WithErr(err, ErrParameterValidationFailed)
	}
	return err
}

// ValidateConstraints validates constraints and returns the collective errors
// for all failed constraints — if any failed — but also checks to see if type was previously validated on failure.
func (p Parameter) ValidateConstraints(value string) (err error) {
	var errs []error
	typeValidated := !p.ConstraintValidatesType()
	// Validate all constraints
	for _, c := range p.constraints {
		err := c.Validate(value)
		if err != nil {
			errs = append(errs, errors.Join(
				p.createConstraintViolationError(c, value),
				err,
			))
		}
	}
	switch {
	case len(errs) == 0:
		// In this case we've got no constraint errors so we are good to go. We don't
		// need to validate type (again.)
		goto end
	case typeValidated:
		// We got constraint errors but since we already validated type, we just need to
		// join the errors since we have already got the errors we need.
		err = CombineErrs(errs)
		goto end
	default:
		// We got constraint errors and since we have not yet validated type, we need to
		// validate type to determine which is the better error message. So fall through.
	}
	// If a type-validating constraint fails, use data type validation as tie-breaker
	err = p.ValidateForDataType(value)
	if err != nil {
		// Type validation failed → this is a type error
		err = errors.Join(
			err,
			CombineErrs(errs),
		)
		goto end
	}
	// No data type error; just use constraint errors
	err = CombineErrs(errs)
end:
	return err
}

func (p Parameter) ErrorDetail(value string) string {
	return fmt.Sprintf("Parameter '%s' expected %s type but got '%s'",
		p.Name,
		p.dataType.WithIndefiniteArticle(),
		value,
	)
}

// ErrorSuggestion generates the appropriate error suggestion based on the error type.
// For constraint violations, it calls the constraint's ErrorSuggestion method.
// For type validation errors, it returns a standard type suggestion.
func (p Parameter) ErrorSuggestion(err error, value, example string) string {
	var pe *ParameterError
	// Check if this is a ParameterError with a constraint violation
	if errors.As(err, &pe) && pe.ConstraintType != "" {
		// This is a constraint violation - find the matching constraint and use its suggestion
		for _, c := range p.constraints {
			if c.String() == pe.ConstraintType {
				return c.ErrorSuggestion(&p, value, example)
			}
		}
		// Fallback if constraint not found (shouldn't happen)
		return fmt.Sprintf("Ensure parameter '%s' satisfies the constraint: %s, for example: %s",
			p.Name, pe.ConstraintType, example)
	}
	// Type validation error or other error - use standard type suggestion
	return fmt.Sprintf("Use %s for '%s' like %v, for example: %s",
		p.dataType.WithIndefiniteArticle(),
		p.Name,
		p.Example(err, &ExampleArgs{SuggestionType: DataTypeSuggestion}),
		example,
	)
}

// createDataTypeViolationError creates a ParameterError for type validation failures.
func (p Parameter) createDataTypeViolationError(value string) error {
	pe := newParameterError(&p, value, ErrParameterDataTypeValidationFailed)
	pe.FaultSource = ClientFaultSource
	pe.Detail = p.ErrorDetail(value)
	return pe
}

// createConstraintViolationError creates a ParameterError for constraint violations.
func (p Parameter) createConstraintViolationError(c Constraint, value string) error {
	pe := newParameterError(&p, value, ErrParameterConstraintValidationFailed)
	pe.FaultSource = ClientFaultSource
	pe.Detail = c.ErrorDetail(&p, value)
	pe.ConstraintType = c.String()
	pe.Err = c.CreateError(value)
	return pe
}
