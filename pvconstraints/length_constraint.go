package pvconstraints

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvtypes.RegisterConstraint(&LengthConstraint{})
}

var _ pvtypes.Constraint = (*LengthConstraint)(nil)

// LengthConstraint validates string length
type LengthConstraint struct {
	pvtypes.BaseConstraint
	min int
	max int
}

func NewLengthConstraint(min int, max int) *LengthConstraint {
	c := &LengthConstraint{min: min, max: max}
	c.BaseConstraint = pvtypes.NewBaseConstraint(c)
	return c
}

func (c *LengthConstraint) ValidDataTypes() []pvtypes.PVDataType {
	return []pvtypes.PVDataType{
		pvtypes.StringType,
		pvtypes.IdentifierType,
		pvtypes.AlphanumericType,
		pvtypes.SlugType,
		pvtypes.EmailType,
	}
}

func (c *LengthConstraint) Parse(value string, dataType pvtypes.PVDataType) (pvtypes.Constraint, error) {
	return ParseLengthConstraint(value)
}

func (c *LengthConstraint) Type() pvtypes.ConstraintType {
	return pvtypes.LengthConstraintType
}

func (c *LengthConstraint) Validate(value string) (err error) {
	length := len(value)
	if length < c.min || length > c.max {
		err = fmt.Errorf("length must be between %d and %d", c.min, c.max)
	}
	return err
}

func (c *LengthConstraint) Rule() string {
	return fmt.Sprintf("%d..%d", c.min, c.max)
}

// ParseLengthConstraint parses min..max format
func ParseLengthConstraint(lengthSpec string) (constraint *LengthConstraint, err error) {
	var parts []string
	var minimum, maximum int

	// Split by ".."
	parts = strings.Split(lengthSpec, "..")
	if len(parts) != 2 {
		err = pvtypes.NewErr(ErrExpectedLengthFormat)
		goto end
	}

	minimum, err = strconv.Atoi(parts[0])
	if err != nil {
		err = pvtypes.NewErr(ErrInvalidMinimumValue,
			"minimum", parts[0],
			err,
		)
		goto end
	}

	maximum, err = strconv.Atoi(parts[1])
	if err != nil {
		err = pvtypes.NewErr(ErrInvalidMaximumValue,
			"maximum", parts[1],
			err,
		)
		goto end
	}

	if minimum > maximum {
		err = pvtypes.NewErr(
			ErrInvalidLengthRangeMinGreaterThanMax,
			"minimum", minimum,
			"maximum", maximum,
		)
		goto end
	}

	if minimum < 0 {
		err = pvtypes.NewErr(
			ErrInvalidLengthRangeNegativeMin,
			"minimum", minimum,
		)
		goto end
	}

	constraint = NewLengthConstraint(minimum, maximum)

end:
	if err != nil {
		err = pvtypes.WithErr(err,
			"length_spec", lengthSpec,
		)
	}
	return constraint, err
}
