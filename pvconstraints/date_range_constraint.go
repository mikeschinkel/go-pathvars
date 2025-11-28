package pvconstraints

import (
	"fmt"
	"strings"
	"time"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvtypes.RegisterConstraint(&DateRangeConstraint{})
}

var _ pvtypes.Constraint = (*DateRangeConstraint)(nil)

// DateRangeConstraint validates date ranges
type DateRangeConstraint struct {
	pvtypes.BaseConstraint
	min time.Time
	max time.Time
}

func NewDateRangeConstraint(min time.Time, max time.Time) *DateRangeConstraint {
	c := &DateRangeConstraint{
		min: min,
		max: max,
	}
	c.BaseConstraint = pvtypes.NewBaseConstraint(c)
	return c
}

func (c *DateRangeConstraint) ValidDataTypes() []pvtypes.PVDataType {
	return []pvtypes.PVDataType{pvtypes.DateType}
}

func (c *DateRangeConstraint) Parse(value string, dataType pvtypes.PVDataType) (pvtypes.Constraint, error) {
	return ParseDateRangeConstraint(value)
}

func (c *DateRangeConstraint) Type() pvtypes.ConstraintType {
	return pvtypes.RangeConstraintType
}

func (c *DateRangeConstraint) Validate(value string) (err error) {
	var d time.Time

	// DateRangeConstraint only accepts YYYY-MM-DD format (same format used in range spec)
	d, err = time.Parse(time.DateOnly, value)
	if err != nil {
		err = ErrInvalidDateFormat
		goto end
	}

	if d.Before(c.min) {
		err = pvtypes.NewErr(ErrDateLessThanMinimum,
			"minimum_date", c.min.Format(time.DateOnly),
		)
		goto end
	}

	if d.After(c.max) {
		err = pvtypes.NewErr(ErrDateGreaterThanMaximum,
			"maximum_date", c.max.Format(time.DateOnly),
		)
		goto end
	}

end:
	if err != nil {
		err = pvtypes.NewErr(
			"date_value", value,
			err,
		)
	}
	return err
}

func (c *DateRangeConstraint) Rule() string {
	return fmt.Sprintf("%s..%s", c.min.Format(time.DateOnly), c.max.Format(time.DateOnly))
}

// ParseDateRangeConstraint parses min..max format for dates
func ParseDateRangeConstraint(rangeSpec string) (constraint *DateRangeConstraint, err error) {
	var parts []string
	var minimum, maximum time.Time

	// Split by ".."
	parts = strings.Split(rangeSpec, "..")
	if len(parts) != 2 {
		err = pvtypes.NewErr(ErrExpectedRangeFormat)
		goto end
	}

	// ParseBytes minimum date (try dateonly format first)
	minimum, err = time.Parse(time.DateOnly, parts[0])
	if err != nil {
		err = pvtypes.NewErr(
			ErrExpectedDateOnlyFormat,
			ErrInvalidMinimumValue,
			"minimum", parts[0],
			err,
		)
		goto end
	}

	// ParseBytes maximum date (try dateonly format first)
	maximum, err = time.Parse(time.DateOnly, parts[1])
	if err != nil {
		err = pvtypes.NewErr(
			ErrExpectedDateOnlyFormat,
			ErrInvalidMaximumValue,
			"maximum", parts[1],
			err,
		)
		goto end
	}

	if minimum.After(maximum) {
		err = pvtypes.NewErr(
			ErrExpectedDateOnlyFormat,
			ErrInvalidMinMaxDate,
			"minimum", minimum.Format(time.DateOnly),
			"maximum", maximum.Format(time.DateOnly),
		)
		goto end
	}

	constraint = NewDateRangeConstraint(minimum, maximum)

end:
	if err != nil {
		err = pvtypes.WithErr(err,
			ErrInvalidRangeConstraint,
			"range_spec", rangeSpec,
		)
	}
	return constraint, err
}
