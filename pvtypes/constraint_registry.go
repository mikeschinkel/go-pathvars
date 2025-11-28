package pvtypes

import (
	"fmt"
	"strings"
)

type ConstraintMapKey string
type ConstraintsMap map[ConstraintMapKey]Constraint

var constraintsMap = make(ConstraintsMap)

type DataTypeAliasMap = map[PVDataTypeSlug]PVDataTypeSlug

var dataTypeAliasMap = make(DataTypeAliasMap)

func RegisterDataTypeAlias(dataType PVDataType, alias PVDataTypeSlug) {
	// Make sure data types and constraints have their init() funcs run
	dtn := dataType.Slug()
	dataTypeAliasMap[dtn] = alias
	// Also register the alias in dataTypeMap so FindDataType can find it
	dataTypeMap[alias] = dataType

	// Now let's check to see if constraints we are aliasing have already been registered.
	// If they have, let's alias them as well.
	c4dt := make([]Constraint, 0)
	for key, c := range constraintsMap {
		dt, a, found := strings.Cut(string(key), "_")
		if !found {
			continue
		}
		if PVDataTypeSlug(a) == alias {
			goto end
		}
		if PVDataTypeSlug(dt) == dtn {
			c4dt = append(c4dt, c)
		}
	}
	// Alias constrains that were previously registered
	for _, c := range c4dt {
		constraintsMap[GetConstraintMapKey(c.Type(), alias)] = c
	}
end:
	return
}

func RegisterConstraint(c Constraint) {
	c.SetOwner(c)
	for _, dt := range c.ValidDataTypes() {
		name := dt.Slug()
		constraintsMap[c.MapKey(name)] = c
		alias, ok := dataTypeAliasMap[name]
		if ok {
			constraintsMap[c.MapKey(alias)] = c
		}
	}
}

func GetConstraintsMap() ConstraintsMap {
	return constraintsMap
}

func GetConstraintMapKey(ct ConstraintType, dtn PVDataTypeSlug) ConstraintMapKey {
	return ConstraintMapKey(fmt.Sprintf("%s_%s", dtn, ct))
}

func GetConstraint(ct ConstraintType, dt PVDataType) (c Constraint, err error) {
	var ok bool
	key := GetConstraintMapKey(ct, dt.Slug())
	c, ok = constraintsMap[key]
	if !ok {
		err = fmt.Errorf("constraint type '%s' not supported", ct)
	}
	return c, err
}
