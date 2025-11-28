package pvtypes

import (
	"maps"
	"slices"
)

// ValuesMap is an ordered map of parameter names to their extracted values.
// Order preservation is critical for:
//   - Error suggestion URLs that match the user's request parameter order (ADR-018)
//   - Deterministic test behavior (no map iteration randomness)
//   - Debug output that reflects actual HTTP request structure
type ValuesMap struct {
	*OrderedMap[Identifier, any]
}

func (vm ValuesMap) Initialized() bool {
	return vm.OrderedMap != nil
}

func NewValuesMap(cap int) ValuesMap {
	return ValuesMap{
		OrderedMap: NewOrderedMap[Identifier, any](cap),
	}
}

func (vm ValuesMap) GetValues(names []Identifier) (values ValuesMap, notFound []Identifier) {
	n := len(names)
	values = NewValuesMap(n)

	notFoundMap := make(map[Identifier]struct{}, len(names))
	for _, name := range names {
		notFoundMap[name] = struct{}{}
	}
	for _, name := range names {
		value, ok := vm.Get(name)
		if !ok {
			continue
		}
		values.Set(name, value)
		delete(notFoundMap, name)
	}
	notFound = slices.Collect(maps.Keys(notFoundMap))
	return values, notFound
}
