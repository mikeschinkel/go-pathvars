package pvconstraints

import (
	"fmt"
	"strings"
	"time"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

// Built-in date/time format aliases
const (
	DateOnlyFormat      = "dateonly"
	UTCDateTimeFormat   = "utc"
	LocalDateTimeFormat = "local"
	DateTimeFormat      = "datetime"
)

func init() {
	pvtypes.RegisterConstraint(&DateFormatConstraint{})
}

var _ pvtypes.Constraint = (*DateFormatConstraint)(nil)

// DateFormatConstraint validates date formats
type DateFormatConstraint struct {
	pvtypes.BaseConstraint
	format string
	parser func(string) (time.Time, error)
}

func NewDateFormatConstraint(format string, parser func(string) (time.Time, error)) *DateFormatConstraint {
	c := &DateFormatConstraint{format: format, parser: parser}
	c.BaseConstraint = pvtypes.NewBaseConstraint(c)
	return c
}

func (c *DateFormatConstraint) ValidDataTypes() []pvtypes.PVDataType {
	return []pvtypes.PVDataType{pvtypes.DateType, pvtypes.UUIDType, pvtypes.StringType}
}

func (c *DateFormatConstraint) Type() pvtypes.ConstraintType {
	return pvtypes.FormatConstraintType
}

// ValidatesType returns true because format constraints perform their own type validation.
func (c *DateFormatConstraint) ValidatesType() bool {
	return true
}

func (c *DateFormatConstraint) Parse(value string, dataType pvtypes.PVDataType) (ct pvtypes.Constraint, err error) {
	// Handle different data types for format constraints
	switch dataType {
	case pvtypes.DateType:
		ct, err = ParseDateFormatConstraint(value)
	case pvtypes.UUIDType:
		ct, err = ParseUUIDFormatConstraint(value)
	case pvtypes.StringType:
		// Check if this is a UUID-like format for strings
		switch strings.ToLower(value) {
		// TODO Make constants for these
		case "ulid", "ksuid", "nanoid":
			ct, err = ParseUUIDFormatConstraint(value)
		default:
			err = pvtypes.NewErr(
				ErrStringFormatOnlySupportsIDFormats,
				"value", value,
				"data_type", dataType.Slug(),
			)
		}
	default:
		err = pvtypes.NewErr(
			ErrInvalidConstraint,
			ErrFormatConstraintUnsupportedDataType,
			"data_type", dataType.Slug(),
		)
	}
	return ct, err
}

func (c *DateFormatConstraint) Validate(value string) (err error) {
	_, err = c.parser(value)
	if err != nil {
		// For partial dates, try to validate against partial formats
		err = c.validatePartialDate(value)
		if err != nil {
			err = fmt.Errorf("invalid date format, expected %s or partial match", c.format)
		}
	}
	return err
}

// validatePartialDate validates partial date formats for multi-segment parameters
func (c *DateFormatConstraint) validatePartialDate(value string) (err error) {
	var partialLayout string

	// Extract tokens from the original format to see what partial formats are valid
	partialLayout, err = c.buildPartialLayout(value)
	if err != nil {
		goto end
	}

	_, err = time.Parse(partialLayout, value)
	if err != nil {
		err = fmt.Errorf("partial date validation failed")
	}

end:
	return err
}

// buildPartialLayout creates a Go time layout for partial date validation
func (c *DateFormatConstraint) buildPartialLayout(value string) (layout string, err error) {
	var segments []string
	var formatSegments []string
	var partialFormat string

	// Split both the value and format by slashes to see how many segments we have
	segments = strings.Split(value, "/")
	formatSegments = strings.Split(c.format, "/")

	// Build a partial format using only the segments we have
	if len(segments) > len(formatSegments) {
		err = pvtypes.NewErr(
			ErrMoreSegmentsThanFormat,
			"value_segments", len(segments),
			"format_segments", len(formatSegments),
		)
		goto end
	}

	// Take only the format segments we need for the actual value segments
	partialFormat = strings.Join(formatSegments[:len(segments)], "/")

	// Use the existing token-based parser to build the Go time layout
	layout, err = buildGoTimeLayout(partialFormat)
	if err != nil {
		err = pvtypes.NewErr(
			ErrFailedToBuildPartialLayout,
			"partial_format", partialFormat,
			err,
		)
	}

end:
	if err != nil {
		err = pvtypes.WithErr(err,
			"value", value,
			"format", c.format,
		)
	}
	return layout, err
}

func (c *DateFormatConstraint) Rule() string {
	return c.format
}

// ParseDateFormatConstraint parses date format specifications.
//
// Date format constraints support four built-in aliases:
//   - format[dateonly]: Date only (yyyy-mm-dd)
//   - format[utc]: Strict UTC timestamps (yyyy-mm-ddThh:mm:ssZ, Z required)
//   - format[local]: Timezone-naive timestamps (yyyy-mm-ddThh:mm:ss, Z forbidden)
//   - format[datetime]: Flexible timestamps (yyyy-mm-ddThh:mm:ss with optional Z, defaults to UTC)
//
// Custom formats use token-based parsing with tokens like: yyyy, mm, dd, hh, ii, ss
func ParseDateFormatConstraint(spec string) (constraint *DateFormatConstraint, err error) {
	var goLayout string
	var parser func(string) (time.Time, error)

	// Handle built-in format aliases
	switch strings.ToLower(spec) {
	case DateOnlyFormat:
		// Date only: yyyy-mm-dd
		goLayout = "2006-01-02"
		parser = func(s string) (time.Time, error) {
			return time.Parse(goLayout, s)
		}
		constraint = NewDateFormatConstraint(spec, parser)
		goto end

	case UTCDateTimeFormat:
		// Strict UTC: yyyy-mm-ddThh:mm:ssZ (Z required)
		parser = func(s string) (time.Time, error) {
			return time.Parse(time.RFC3339, s)
		}
		constraint = NewDateFormatConstraint(spec, parser)
		goto end

	case LocalDateTimeFormat:
		// Timezone-naive: yyyy-mm-ddThh:mm:ss (Z forbidden)
		goLayout = "2006-01-02T15:04:05"
		parser = func(s string) (time.Time, error) {
			return time.Parse(goLayout, s)
		}
		constraint = NewDateFormatConstraint(spec, parser)
		goto end

	case DateTimeFormat:
		// Flexible: yyyy-mm-ddThh:mm:ss with optional Z (defaults to UTC)
		parser = func(s string) (time.Time, error) {
			// Try with Z first
			t, err := time.Parse(time.RFC3339, s)
			if err == nil {
				return t, nil
			}
			// Try without Z, treat as UTC
			t, err = time.Parse("2006-01-02T15:04:05", s)
			if err != nil {
				return t, err
			}
			// Convert to UTC
			return t.UTC(), nil
		}
		constraint = NewDateFormatConstraint(spec, parser)
		goto end
	}

	// ParseBytes the format specification to build Go time layout
	goLayout, err = buildGoTimeLayout(spec)
	if err != nil {
		err = pvtypes.NewErr(
			ErrInvalidDateFormatSpec,
			"spec", spec,
			err,
		)
		goto end
	}

	parser = func(s string) (time.Time, error) {
		return time.Parse(goLayout, s)
	}

	constraint = NewDateFormatConstraint(spec, parser)

end:
	return constraint, err
}

// buildGoTimeLayout converts a date format specification to Go time layout
func buildGoTimeLayout(spec string) (layout string, err error) {
	var result []rune
	var i int
	var hasHour bool
	var token string
	var hasAnyToken bool

	// ParseBytes character by character, looking for date/time tokens
	for i < len(spec) {
		// Try to match each possible token at current position
		token, hasHour, err = matchToken(spec, i, hasHour)
		if err != nil {
			goto end
		}

		if token != "" {
			// Found a token, append its Go layout equivalent
			result = append(result, []rune(token)...)
			i += tokenLength(spec, i)
			hasAnyToken = true
		} else {
			// Not a token, append the literal character
			result = append(result, rune(spec[i]))
			i++
		}
	}

	// Validate that we found at least one date/time token
	if !hasAnyToken {
		err = pvtypes.NewErr(
			ErrNoValidDateTimeTokens,
			"spec", spec,
		)
		goto end
	}

	layout = string(result)

end:
	return layout, err
}

// matchToken attempts to match a date/time token at the given position
func matchToken(spec string, pos int, hasHour bool) (goToken string, newHasHour bool, err error) {
	newHasHour = hasHour

	// Check for each possible token
	if matchesAt(spec, pos, "yyyy") {
		goToken = "2006"
		goto end
	}
	if matchesAt(spec, pos, "yy") {
		goToken = "06"
		goto end
	}
	if matchesAt(spec, pos, "dd") {
		goToken = "02"
		goto end
	}
	if matchesAt(spec, pos, "hh") {
		goToken = "15"
		newHasHour = true
		goto end
	}
	if matchesAt(spec, pos, "ss") {
		goToken = "05"
		goto end
	}
	if matchesAt(spec, pos, "ii") {
		// ii always means minutes
		goToken = "04"
		goto end
	}
	if matchesAt(spec, pos, "mm") {
		// mm disambiguation: month if no hour seen yet, minutes if hour seen
		switch {
		case hasHour:
			goToken = "04" // minutes
		case isStandaloneMM(spec):
			// Check if this is a standalone mm (ambiguous)
			err = pvtypes.NewErr(
				ErrAmbiguousMMToken,
				"spec", spec,
				"position", pos,
			)
			goto end
		default:
			goToken = "01" // month
		}
		goto end
	}

end:
	return goToken, newHasHour, err
}

// matchesAt checks if the given token matches at the specified position
func matchesAt(spec string, pos int, token string) bool {
	var i int

	if pos+len(token) > len(spec) {
		return false
	}

	for i = 0; i < len(token); i++ {
		if spec[pos+i] != token[i] {
			return false
		}
	}

	return true
}

// tokenLength returns the length of the token at the given position
func tokenLength(spec string, pos int) int {
	if matchesAt(spec, pos, "yyyy") {
		return 4
	}
	if matchesAt(spec, pos, "yy") || matchesAt(spec, pos, "mm") ||
		matchesAt(spec, pos, "dd") || matchesAt(spec, pos, "hh") ||
		matchesAt(spec, pos, "ii") || matchesAt(spec, pos, "ss") {
		return 2
	}
	return 1
}

// isStandaloneMM checks if the spec contains only "mm" as a token (ambiguous case)
func isStandaloneMM(spec string) bool {
	var tokenCount int
	var i int

	// Count tokens in the spec
	for i < len(spec) {
		if matchesAt(spec, i, "yyyy") {
			tokenCount++
			i += 4
		} else if matchesAt(spec, i, "yy") || matchesAt(spec, i, "mm") ||
			matchesAt(spec, i, "dd") || matchesAt(spec, i, "hh") ||
			matchesAt(spec, i, "ii") || matchesAt(spec, i, "ss") {
			tokenCount++
			i += 2
		} else {
			i++
		}
	}

	// If only one token and it's mm, it's standalone and ambiguous
	return tokenCount == 1 && hasOnlyMMToken(spec)
}

// hasOnlyMMToken checks if the spec contains only the mm token
func hasOnlyMMToken(spec string) bool {
	var i int

	for i < len(spec) {
		if matchesAt(spec, i, "mm") {
			// Found mm, continue to check if there are other tokens
			i += 2
			continue
		}
		if matchesAt(spec, i, "yyyy") || matchesAt(spec, i, "yy") ||
			matchesAt(spec, i, "dd") || matchesAt(spec, i, "hh") ||
			matchesAt(spec, i, "ii") || matchesAt(spec, i, "ss") {
			// Found another token
			return false
		}
		i++
	}

	// Check if we actually found mm in the spec
	return containsToken(spec, "mm")
}

// containsToken checks if the spec contains the given token
func containsToken(spec string, token string) bool {
	var i int

	for i < len(spec) {
		if matchesAt(spec, i, token) {
			return true
		}
		i++
	}
	return false
}
