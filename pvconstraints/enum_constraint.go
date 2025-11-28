package pvconstraints

import (
	"fmt"
	"strings"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

func init() {
	pvtypes.RegisterConstraint(&EnumConstraint{})
}

var _ pvtypes.Constraint = (*EnumConstraint)(nil)

// EnumConstraint validates against allowed values
type EnumConstraint struct {
	pvtypes.BaseConstraint
	values map[string]bool
	list   []string
}

func NewEnumConstraint(values map[string]bool, list []string) *EnumConstraint {
	c := &EnumConstraint{values: values, list: list}
	c.BaseConstraint = pvtypes.NewBaseConstraint(c)
	return c
}

func (c *EnumConstraint) Parse(value string, dataType pvtypes.PVDataType) (pvtypes.Constraint, error) {
	return ParseEnumConstraint(value)
}

func (c *EnumConstraint) Type() pvtypes.ConstraintType {
	return pvtypes.EnumConstraintType
}

func (c *EnumConstraint) Validate(value string) (err error) {
	if !c.values[value] {
		err = fmt.Errorf("value must be one of: %s", strings.Join(c.list, ", "))
	}
	return err
}

func (c *EnumConstraint) Rule() string {
	return strings.Join(c.list, ",")
}

func (c *EnumConstraint) ValidDataTypes() []pvtypes.PVDataType {
	return []pvtypes.PVDataType{
		pvtypes.IntegerType,
		pvtypes.BooleanType,
		pvtypes.StringType,
		pvtypes.IdentifierType,
		pvtypes.AlphanumericType,
		pvtypes.SlugType,
		pvtypes.EmailType,
	}
}

func (c *EnumConstraint) ErrorDetail(param *pvtypes.Parameter, value string) string {
	return fmt.Sprintf("Parameter '%s' with value '%s' failed constraint validation: value '%s' is not in the allowed set: [%s]",
		param.Name,
		value,
		value,
		strings.Join(c.list, ", "),
	)
}

func (c *EnumConstraint) Example(err error) (ex any) {
	// Return the first enum value as the example
	if len(c.list) > 0 {
		ex = c.list[0]
	}
	return ex
}

// ParseEnumConstraint parses val1,val2,val3 format
func ParseEnumConstraint(enumSpec string) (constraint *EnumConstraint, err error) {
	var values []string
	var valueMap map[string]bool
	var value string
	var errs []error

	enumError := func() error {
		return pvtypes.NewErr(
			ErrEnumValueIsEmpty,
			"enum", enumSpec,
		)
	}

	if enumSpec == "" {
		err = enumError()
		goto end
	}

	// Split by comma
	values = strings.Split(enumSpec, ",")
	valueMap = make(map[string]bool)

	for _, value = range values {
		value = strings.TrimSpace(value)
		if value == "" {
			errs = append(errs, enumError())
			continue
		}
		// TODO Ensure value is a valid identifier
		valueMap[value] = true
	}
	err = pvtypes.CombineErrs(errs)
	if err != nil {
		goto end
	}

	constraint = NewEnumConstraint(valueMap, values)

end:
	if err != nil {
		err = pvtypes.WithErr(err,
			ErrInvalidEnumConstraint,
			"enum_spec", enumSpec,
		)
	}
	return constraint, err
}
