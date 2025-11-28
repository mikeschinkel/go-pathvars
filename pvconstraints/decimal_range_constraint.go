package pvconstraints

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvtypes.RegisterConstraint(&DecimalRangeConstraint{})
}

var _ pvtypes.Constraint = (*DecimalRangeConstraint)(nil)

// DecimalRangeConstraint validates decimal ranges
type DecimalRangeConstraint struct {
	pvtypes.BaseConstraint
	min float64
	max float64
}

func NewDecimalRangeConstraint(min float64, max float64) *DecimalRangeConstraint {
	c := &DecimalRangeConstraint{
		min: min,
		max: max,
	}
	c.BaseConstraint = pvtypes.NewBaseConstraint(c)
	return c
}

func (c *DecimalRangeConstraint) ValidDataTypes() []pvtypes.PVDataType {
	return []pvtypes.PVDataType{pvtypes.DecimalType, pvtypes.RealType}
}

func (c *DecimalRangeConstraint) Parse(value string, dataType pvtypes.PVDataType) (pvtypes.Constraint, error) {
	return ParseDecimalRangeConstraint(value)
}

func (c *DecimalRangeConstraint) Type() pvtypes.ConstraintType {
	return pvtypes.RangeConstraintType
}

func (c *DecimalRangeConstraint) Validate(value string) (err error) {
	var n float64

	n, err = strconv.ParseFloat(value, 64)
	if err != nil {
		goto end
	}

	if n < c.min || n > c.max {
		err = fmt.Errorf("value must be between %g and %g", c.min, c.max)
	}

end:
	return err
}

func (c *DecimalRangeConstraint) Rule() string {
	return fmt.Sprintf("%g..%g", c.min, c.max)
}

// ParseDecimalRangeConstraint parses min..max format for decimals
func ParseDecimalRangeConstraint(rangeSpec string) (constraint *DecimalRangeConstraint, err error) {
	var parts []string
	var minimum, maximum float64

	// Split by ".."
	parts = strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		err = pvtypes.NewErr(ErrExpectedRangeFormat)
		goto end
	}

	minimum, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		err = pvtypes.NewErr(
			ErrInvalidMinimumValue,
			"minimum", parts[0],
			err,
		)
		goto end
	}

	maximum, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		err = pvtypes.NewErr(
			ErrInvalidMaximumValue,
			"maximum", parts[1],
			err,
		)
		goto end
	}

	if minimum > maximum {
		err = pvtypes.NewErr(
			ErrInvalidMinMaxValue,
			fmt.Errorf("minimum=%g", minimum),
			fmt.Errorf("maximum=%g", maximum),
		)
		goto end
	}

	constraint = NewDecimalRangeConstraint(minimum, maximum)

end:
	if err != nil {
		err = pvtypes.WithErr(err,
			ErrInvalidRangeValue,
			"range_spec", rangeSpec,
		)
	}
	return constraint, err
}
