package pvconstraints

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvtypes.RegisterConstraint(&IntegerRangeConstraint{})
}

var _ pvtypes.Constraint = (*IntegerRangeConstraint)(nil)

// IntegerRangeConstraint validates integer ranges
type IntegerRangeConstraint struct {
	pvtypes.BaseConstraint
	min int64
	max int64
}

func NewIntRangeConstraint(min int64, max int64) *IntegerRangeConstraint {
	c := &IntegerRangeConstraint{min: min, max: max}
	c.BaseConstraint = pvtypes.NewBaseConstraint(c)
	return c
}

func (c *IntegerRangeConstraint) ValidDataTypes() []pvtypes.PVDataType {
	return []pvtypes.PVDataType{pvtypes.IntegerType}
}

func (c *IntegerRangeConstraint) Parse(value string, dataType pvtypes.PVDataType) (pvtypes.Constraint, error) {
	return ParseIntRangeConstraint(value)
}

func (c *IntegerRangeConstraint) Type() pvtypes.ConstraintType {
	return pvtypes.RangeConstraintType
}

func (c *IntegerRangeConstraint) Validate(value string) (err error) {
	var n int64

	n, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		goto end
	}

	if n < c.min || n > c.max {
		err = fmt.Errorf("value must be between %d and %d", c.min, c.max)
	}

end:
	return err
}

func (c *IntegerRangeConstraint) Rule() string {
	return fmt.Sprintf("%d..%d", c.min, c.max)
}

func (c *IntegerRangeConstraint) ErrorDetail(param *pvtypes.Parameter, value string) string {
	var n int64
	var err error
	n, err = strconv.ParseInt(value, 10, 64)
	if err != nil {
		goto end
	}
	if n < c.min || n > c.max {
		return fmt.Sprintf("Parameter '%s' with value '%s' failed constraint validation: value %d is outside the allowed range of %d..%d",
			param.Name,
			value,
			n,
			c.min,
			c.max,
		)
	}
end:
	return c.BaseConstraint.ErrorDetail(param, value)
}

func (c *IntegerRangeConstraint) ErrorSuggestion(param *pvtypes.Parameter, value, example string) string {
	return fmt.Sprintf("Ensure parameter '%s' satisfies the constraint: %s, for example: %s", param.Name, c.String(), example)
}

// Example returns the midpoint of the range as a representative example value.
// The error parameter is currently unused but allows for future context-aware examples.
func (c *IntegerRangeConstraint) Example(err error) any {
	return (c.min + c.max) / 2
}

// ParseIntRangeConstraint parses min..max format for integers
func ParseIntRangeConstraint(rangeSpec string) (constraint *IntegerRangeConstraint, err error) {
	var parts []string
	var minimum, maximum int64
	var errs []error

	// Split by ".."
	parts = strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		err = pvtypes.NewErr(ErrExpectedRangeFormat)
		if err != nil {
			errs = append(errs, err)
		}
	}

	minimum, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		err = pvtypes.NewErr(
			ErrInvalidMinimumValue,
			"minimum", parts[0],
			err,
		)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(parts) == 1 {
		err = pvtypes.CombineErrs(errs)
		goto end
	}

	maximum, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		err = pvtypes.NewErr(
			ErrInvalidMaximumValue,
			"maximum", parts[1],
			err,
		)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if minimum > maximum {
		err = pvtypes.NewErr(
			ErrInvalidMinMaxValue,
			"minimum", minimum,
			"maximum", maximum,
		)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) != 0 {
		err = pvtypes.CombineErrs(errs)
		goto end
	}

	constraint = NewIntRangeConstraint(minimum, maximum)

end:
	if err != nil {
		err = pvtypes.WithErr(err,
			ErrInvalidRangeConstraint,
			"range_spec", rangeSpec,
		)
	}
	return constraint, err
}
