// Package pvtypes/constraints defines the constraint system for parameter validation.
// Constraints provide additional validation rules beyond basic data type checking,
// such as ranges, formats, enums, and regular expressions.
package pvtypes

import (
	"errors"
	"strings"
)

// ConstraintType represents the type of constraint applied to a parameter.
type ConstraintType string

// Supported constraint types for parameter validation.
const (
	// FormatConstraintType validates parameter values against specific formats (e.g., date formats, UUID versions).
	FormatConstraintType ConstraintType = "format"

	// EnumConstraintType validates that parameter values match one of a predefined set of allowed values.
	EnumConstraintType ConstraintType = "enum"

	// LengthConstraintType validates that string parameter values fall within specified length ranges.
	LengthConstraintType ConstraintType = "length"

	// NotEmptyConstraintType validates that parameter values are not empty strings.
	NotEmptyConstraintType ConstraintType = "notempty"

	// RangeConstraintType validates that numeric parameter values fall within specified numeric ranges.
	RangeConstraintType ConstraintType = "range"

	// RegexConstraintType validates parameter values against regular expression patterns.
	RegexConstraintType ConstraintType = "regex"
)

type Constraints []Constraint

func (c Constraints) String() (s string) {
	sb := strings.Builder{}
	for _, constraint := range c {
		sb.WriteString(constraint.String())
		sb.WriteByte(',')
	}
	s = sb.String()
	if len(s) > 0 {
		s = s[:len(s)-1]
	}
	return s
}

// Constraint interface defines the contract for parameter validation constraints.
// All constraint implementations must provide validation, parsing, and metadata methods.
type Constraint interface {
	// Validate checks if the given value satisfies this constraint.
	Validate(value string) error

	// String returns a human-readable representation of this constraint.
	String() string

	// Rule returns the constraint's rule from within square brackets
	Rule() string

	// Type returns the type of this constraint.
	Type() ConstraintType

	// Parse creates a new instance of this constraint from a string specification.
	Parse(value string, dataType PVDataType) (Constraint, error)

	// ValidDataTypes returns the data types that this constraint can be applied to.
	ValidDataTypes() []PVDataType

	// MapKey generates a unique key for constraint registry lookup.
	MapKey(dt PVDataTypeSlug) ConstraintMapKey

	// ValidatesType returns true if this constraint performs type validation,
	// allowing it to replace default data type validation for its parameter.
	ValidatesType() bool

	// Example returns an example value that satisfies this constraint.
	// The error parameter provides context about what failed validation, allowing
	// the constraint to return a more appropriate example (e.g., midpoint for ranges).
	// Returns nil if no specific example is available (use data type example instead).
	// This is particularly important for format constraints that validate specific
	// formats (e.g., UUID v4 vs v1, ISO8601 dates, etc.).
	Example(err error) any

	// ErrorDetail returns a detailed error message explaining why validation failed.
	// The parameter provides context about the parameter being validated.
	ErrorDetail(param *Parameter, value string) string

	// ErrorSuggestion returns a helpful suggestion for fixing the validation error.
	// The parameter provides context about the parameter being validated.
	ErrorSuggestion(param *Parameter, value, example string) string

	// SetOwner sets owner for constraints that do not do it on instantiation.
	SetOwner(Constraint)

	CreateError(string) *ConstraintError
}

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
	var ctm ConstraintsMap
	var ct ConstraintType
	var constraint Constraint
	var pos, last, valueStart, constraintStart int
	var value string
	var errs []error
	var mode byte
	var ok bool
	var regexStart, regexEnd int

	typeName := dataType.Slug()
	const (
		typeMode  = 't'
		valueMode = 'v'
	)

	if spec == "" {
		goto end
	}

	ctm = GetConstraintsMap()

	// Pre-scan for regex constraint to find its true boundaries
	regexStart, regexEnd = findRegexBoundaries(spec)

	mode = typeMode
	last = len(spec)
	constraintStart = 0
	for pos < last {
		ch := spec[pos]
		pos++
		if isWhitespace(ch) {
			continue
		}

		//goland:noinspection GoDfaConstantCondition
		switch mode {
		case typeMode:
			if ch == '[' {
				// Found start of constraint value
				valueStart = pos
				ct = ConstraintType(spec[constraintStart : pos-1])
				key := GetConstraintMapKey(ct, typeName)
				constraint, ok = ctm[key]
				if !ok {
					errs = append(errs,
						NewErr(ErrUnknownConstraintType,
							"constraint_spec", spec,
							"constraint_type", ct,
							"data_type", typeName,
						),
					)
					continue
				}
				mode = valueMode
				continue
			}
			if ch == ',' || pos == last {
				// Found end of constraint type without brackets (like "notempty")
				constraintEnd := pos
				if ch == ',' {
					constraintEnd--
				}
				ct = ConstraintType(spec[constraintStart:constraintEnd])
				key := GetConstraintMapKey(ct, typeName)
				constraint, ok = ctm[key]
				if !ok {
					errs = append(errs,
						NewErr(ErrUnknownConstraintType,
							"constraint_spec", spec,
							"constraint_type", ct,
							"data_type", typeName,
						),
					)
					continue
				}
				// ParseBytes constraint with empty value (no arguments)
				constraint, err = constraint.Parse("", dataType)
				if err != nil {
					errs = append(errs,
						NewErr(ErrParseFailed,
							"constraint_value", "",
							"constraint_type", ct,
							"constraint_spec", spec,
							"data_type", typeName,
							err,
						),
					)
				} else {
					constraints = append(constraints, constraint)
				}

				if ch == ',' {
					// Skip past the comma and any whitespace to continue parsing next constraint
					for pos < last && (spec[pos] == ',' || isWhitespace(spec[pos])) {
						pos++
					}
					constraintStart = pos
					continue
				}
				break
			}
			if !isConstraintTypeChar(ch) {
				errs = append(errs,
					NewErr(ErrInvalidSyntax,
						ErrInvalidConstraintTypeCharacter,
						"position", pos,
						"character", string(ch),
						"constraint_type", ct,
						"constraint_spec", spec,
						"data_type", typeName,
					),
				)
				continue
			}

		case valueMode:
			// Special handling for regex constraint - use pre-scanned boundaries
			if ct == RegexConstraintType && regexStart != -1 && constraintStart == regexStart {
				// Jump to the pre-scanned end position
				pos = regexEnd + 1
				value = spec[valueStart:regexEnd]
				constraint, err = constraint.Parse(value, dataType)
				if err != nil {
					errs = append(errs,
						NewErr(ErrParseFailed,
							"constraint_value", value,
							"constraint_type", ct,
							"constraint_spec", spec,
							"data_type", dataType.Slug(),
							"start_pos", valueStart,
							"end_pos", regexEnd,
							err,
						),
					)
				} else {
					constraints = append(constraints, constraint)
				}
				mode = typeMode
				// Skip past any whitespace and comma to next constraint
				for pos < last && isWhitespace(spec[pos]) {
					pos++
				}
				if pos < last && spec[pos] == ',' {
					pos++
					for pos < last && isWhitespace(spec[pos]) {
						pos++
					}
					constraintStart = pos
				}
				continue
			}

			if ch == ']' {
				// Look ahead to see if this ends the constraint (comma or end of string)
				isEndOfConstraint := false
				if pos == last {
					// End of string
					isEndOfConstraint = true
				} else {
					// Check for optional whitespace followed by comma or end
					lookahead := pos
					for lookahead < last && isWhitespace(spec[lookahead]) {
						lookahead++
					}
					if lookahead == last || spec[lookahead] == ',' {
						isEndOfConstraint = true
					}
				}

				if isEndOfConstraint {
					// This closes the constraint
					value = spec[valueStart : pos-1]
					constraint, err = constraint.Parse(value, dataType)
					if err != nil {
						errs = append(errs,
							NewErr(ErrParseFailed,
								"constraint_value", value,
								"constraint_type", ct,
								"constraint_spec", spec,
								"data_type", dataType.Slug(),
								"start_pos", valueStart,
								"end_pos", pos-1,
								err,
							),
						)
					} else {
						constraints = append(constraints, constraint)
					}
					mode = typeMode
					// Skip past any whitespace and comma to next constraint
					for pos < last && isWhitespace(spec[pos]) {
						pos++
					}
					if pos < last && spec[pos] == ',' {
						pos++
						for pos < last && isWhitespace(spec[pos]) {
							pos++
						}
						constraintStart = pos
					}
					continue
				}
			}
			if pos == last {
				// End of string without closing bracket - malformed
				errs = append(errs,
					NewErr(ErrInvalidSyntax,
						"position", pos,
						"constraint_type", ct,
						"constraint_spec", spec,
						"data_type", dataType.Slug(),
						errors.New("constraint value not properly closed"),
					),
				)
				continue
			}
		}
	}

end:
	if len(errs) > 0 {
		err = CombineErrs(errs)
	}
	for _, c := range constraints {
		// Do this in case the constraint parser did not do this itself
		// If we add properties to baseConstraint we'll need to do the same for those properties here.
		c.SetOwner(c)
	}
	return constraints, err
}

// findRegexBoundaries uses bidirectional parsing to find the true boundaries of a regex constraint.
// Returns (-1, -1) if no regex constraint is found.
// This handles regex patterns that contain [ and ] characters by:
// 1. Finding "regex[" from the start
// 2. Finding the last "]" that could close the regex
// 3. If other constraints follow, finding the ] before them
func findRegexBoundaries(spec string) (start, end int) {
	// Find "regex[" - the start of regex constraint
	regexPrefix := "regex["
	start = strings.Index(spec, regexPrefix)
	if start == -1 {
		return -1, -1 // No regex constraint
	}

	// Find the last "]" in the spec - this is our candidate end
	end = strings.LastIndex(spec, "]")
	if end == -1 || end <= start+len(regexPrefix) {
		return -1, -1 // No closing bracket or it's before/at the opening
	}

	// Check if there are other constraints after regex
	// Look for a comma after the potential regex end
	afterEnd := end + 1
	if afterEnd < len(spec) {
		// Skip whitespace
		for afterEnd < len(spec) && isWhitespace(spec[afterEnd]) {
			afterEnd++
		}
	}

	return start, end
}

// isConstraintTypeChar returns true if the character is valid in a constraint type name.
// Constraint type names can contain lowercase letters and underscores.
func isConstraintTypeChar(ch byte) (isChar bool) {
	if 'a' <= ch && ch <= 'z' {
		isChar = true
		goto end
	}
	if ch == '_' {
		isChar = true
		goto end
	}
end:
	return isChar
}

// isWhitespace returns true if the character is considered whitespace.
// Recognized whitespace characters include space, tab, newline, and carriage return.
func isWhitespace(ch byte) (isWS bool) {
	switch ch {
	case ' ', '\t', '\n', '\r':
		isWS = true
	}
	return isWS
}
