package pvtypes

import (
	"fmt"
)

// BaseConstraint provides common functionality for all constraint implementations.
// It maintains a reference to the owning constraint for proper method delegation.
type BaseConstraint struct {
	// owner holds a reference to the constraint that embeds this base.
	owner Constraint
	// IMPORTANT: If we add properties here we'll need to ensure they are set in all
	// constraints and/or everywhere SetOwner() is called.
}

func (c *BaseConstraint) CreateError(value string) *ConstraintError {
	return &ConstraintError{
		Err:           ErrConstraintValidationFailed,
		Constraint:    c.owner.String(),
		FaultSource:   ClientFaultSource,
		ReceivedValue: value,
	}
}

// NewBaseConstraint creates a new base constraint with the specified owner.
func NewBaseConstraint(owner Constraint) BaseConstraint {
	return BaseConstraint{
		owner: owner,
	}
}

func (c *BaseConstraint) SetOwner(owner Constraint) {
	c.owner = owner
}

func (c *BaseConstraint) String() string {
	return fmt.Sprintf("%s[%s]", c.owner.Type(), c.owner.Rule())
}

// EnsureBaseConstraint sets the owner reference for proper constraint operation.
func (c *BaseConstraint) EnsureBaseConstraint(owner Constraint) {
	c.owner = owner
}

// MapKey generates a constraint registry key using the owner's type and data type.
func (c *BaseConstraint) MapKey(dt PVDataTypeSlug) ConstraintMapKey {
	return GetConstraintMapKey(c.owner.Type(), dt)
}

// ValidatesType returns false by default. Format constraints override this to return true.
func (c *BaseConstraint) ValidatesType() bool {
	return false
}

// Example returns nil by default, indicating no specific example is available.
// Constraints that validate specific formats (ValidatesType()==true) should override
// this to return an appropriate example value.
// The error parameter is ignored in the base implementation but allows specific
// constraints to provide context-aware examples.
func (c *BaseConstraint) Example(err error) any {
	return nil
}

// ErrorDetail provides a default detailed error message for constraint violations.
// Specific constraints can override this to provide more detailed information.
func (c *BaseConstraint) ErrorDetail(param *Parameter, value string) string {
	return fmt.Sprintf("Parameter '%s' failed constraint '%s': value '%s' does not match required format",
		param.Name,
		c.owner.String(),
		value,
	)
}

// ErrorSuggestion provides a default suggestion message for constraint violations.
// Specific constraints can override this to provide more tailored guidance.
func (c *BaseConstraint) ErrorSuggestion(param *Parameter, value, example string) string {
	return fmt.Sprintf("Ensure parameter '%s' satisfies the constraint: %s, for example: %s",
		param.Name,
		c.String(),
		example,
	)
}
