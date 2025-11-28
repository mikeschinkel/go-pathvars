package pvconstraints

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvtypes.RegisterConstraint(&RegexConstraint{})
}

var _ pvtypes.Constraint = (*RegexConstraint)(nil)

// RegexConstraint validates against regex regex
type RegexConstraint struct {
	pvtypes.BaseConstraint
	regex *regexp.Regexp
	raw   string
}

func NewRegexConstraint(regex *regexp.Regexp, raw string) *RegexConstraint {
	c := &RegexConstraint{
		regex: regex,
		raw:   raw,
	}
	c.BaseConstraint = pvtypes.NewBaseConstraint(c)
	return c
}

func (c *RegexConstraint) ValidDataTypes() []pvtypes.PVDataType {
	return []pvtypes.PVDataType{
		pvtypes.StringType,
		pvtypes.SlugType,
		pvtypes.AlphanumericType,
		pvtypes.EmailType,
		pvtypes.IdentifierType,
	}
}

func (c *RegexConstraint) Parse(value string, dataType pvtypes.PVDataType) (pvtypes.Constraint, error) {
	return ParseRegexConstraint(value)
}

func (c *RegexConstraint) Type() pvtypes.ConstraintType {
	return pvtypes.RegexConstraintType
}

func (c *RegexConstraint) Validate(value string) (err error) {
	if !c.regex.MatchString(value) {
		err = fmt.Errorf("value does not match regex %s", c.raw)
	}
	return err
}

func (c *RegexConstraint) Rule() string {
	return c.raw
}

func (c *RegexConstraint) ErrorSuggestion(param *pvtypes.Parameter, value, example string) string {
	// TODO Add more specific advice, and don't use this advice when not applicable
	return "Do not include ^ or $ anchors in your regex pattern; regex patterns automatically match the full parameter value."
}

// ParseRegexConstraint parses a regex pattern and automatically anchors it for full string matching.
// Patterns must not include ^ or $ anchors - they are added automatically to ensure the pattern
// matches the complete parameter value, not just a substring.
func ParseRegexConstraint(pattern string) (constraint *RegexConstraint, err error) {
	var regex *regexp.Regexp
	var anchoredPattern string
	var errs []error
	var hasStart, hasEnd bool

	if pattern == "" {
		err = pvtypes.NewErr(ErrEmptyRegexPattern)
		goto end
	}

	// Check for anchors and collect all errors before returning
	hasStart = strings.HasPrefix(pattern, "^")
	hasEnd = strings.HasSuffix(pattern, "$")

	if hasStart && hasEnd {
		errs = append(errs, ErrRegexPatternContainsBothAnchors)
	}
	if hasStart {
		errs = append(errs, ErrRegexPatternContainsStartAnchor)
	}
	if hasEnd {
		errs = append(errs, ErrRegexPatternContainsEndAnchor)
	}

	if len(errs) > 0 {
		err = pvtypes.NewErr(
			ErrInvalidRegexPattern,
			pvtypes.CombineErrs(errs),
		)
		goto end
	}

	// Auto-anchor the pattern for full string matching
	anchoredPattern = "^" + pattern + "$"

	// Compile the anchored pattern
	regex, err = regexp.Compile(anchoredPattern)
	if err != nil {
		err = pvtypes.NewErr(
			ErrInvalidRegexPattern,
			err,
		)
		goto end
	}

	// Store original pattern (without anchors) for display
	constraint = NewRegexConstraint(regex, pattern)

end:
	if err != nil {
		err = pvtypes.WithErr(err,
			ErrInvalidRegexConstraint,
			"regex", pattern,
		)
	}
	return constraint, err
}
