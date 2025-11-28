package pvconstraints

import (
	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

// ParseRangeConstraint uses the registry to find and parse the appropriate range constraint based on data type
func ParseRangeConstraint(rangeSpec string, dataType pvtypes.PVDataType) (constraint pvtypes.Constraint, err error) {
	constraint, err = pvtypes.GetConstraint(pvtypes.RangeConstraintType, dataType)
	if err != nil {
		goto end
	}

	constraint, err = constraint.Parse(rangeSpec, dataType)

end:
	if err != nil {
		// Make sure constraint is nil and not a non-nil interface containing a nil.
		constraint = nil
	}
	return constraint, err
}
