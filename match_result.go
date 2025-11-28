// Package pathvars/match_result defines the MatchResult type which contains
// the results of matching an HTTP request against a route template.
// It provides access to extracted parameter values and route information.
package pathvars

import (
	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

// MatchAttempt represents the result of attempting to match a request against a route template.
// It provides detailed information about what matched and what didn't, allowing the router
// to make intelligent decisions about whether to try the next route or return an error.
type MatchAttempt struct {
	// PathMatched indicates whether the URL path matched the route's path pattern (regex match).
	PathMatched bool

	// QueryMatched indicates whether all query parameters validated successfully.
	QueryMatched bool

	// ValuesMap contains extracted parameter values (may be partial if validation failed).
	ValuesMap pvtypes.ValuesMap
}

// Matched returns true if both path and query matched successfully.
func (ma MatchAttempt) Matched() bool {
	return ma.PathMatched && ma.QueryMatched
}

// ShouldContinue returns true if the router should try the next route.
// This happens when the path didn't match - we haven't found the right route yet.
func (ma MatchAttempt) ShouldContinue() bool {
	return !ma.PathMatched
}

// MatchResult represents the result of matching an HTTP request against a route template.
// It contains the matched route index and extracted parameter values for memory efficiency.
type MatchResult struct {
	// Index indicates which route was matched in the router's route list.
	Index int

	Route *Route

	// valuesMap contains the extracted parameter values from the matched request.
	// This field is private to control access and ensure proper initialization.
	valuesMap pvtypes.ValuesMap
}

// NewMatchResult creates a new MatchResult with the specified route index and parameter values.
func NewMatchResult(r *Route, valuesMap pvtypes.ValuesMap) MatchResult {
	return MatchResult{
		Index:     r.Index,
		Route:     r,
		valuesMap: valuesMap,
	}
}

func (m MatchResult) GetValues(names []Identifier) (pvtypes.ValuesMap, []Identifier) {
	return m.valuesMap.GetValues(names)
}

// ValuesMap returns the map of extracted parameter values.
// This includes:
//   - Path parameters (from URL path segments)
//   - Template-defined query parameters (validated from template like ?{email:string})
//
// NOTE: This does NOT include Params-defined query parameters that are not in the template.
// Those are validated separately via ValidateQueryParameters() using parsedQuery.
// If the internal map is nil, it initializes an empty map to prevent nil pointer issues.
func (m MatchResult) ValuesMap() pvtypes.ValuesMap {
	if m.valuesMap.IsNil() {
		m.valuesMap = pvtypes.NewValuesMap(0)
	}
	return m.valuesMap
}

// GetValue returns the value of a named parameter and whether it was found.
// Returns the parameter value and true if the parameter exists, or empty string and false otherwise.
func (m MatchResult) GetValue(name Identifier) (value any, found bool) {
	value, found = m.valuesMap.Get(name)
	return value, found
}

// VarCount returns the number of extracted parameters.
func (m MatchResult) VarCount() int {
	return m.valuesMap.Len()
}

// HasVars returns true if any parameters were extracted from the request.
func (m MatchResult) HasVars() bool {
	return !m.valuesMap.IsNil() && m.valuesMap.Len() > 0
}

// ForEachVar iterates over all extracted parameters, calling the provided function
// for each name-value pair. If the function returns true, iteration continues;
// if it returns false, iteration stops early.
func (m MatchResult) ForEachVar(fn func(name Identifier, value any) bool) {
	for name, value := range m.valuesMap.Iterator() {
		if fn(name, value) {
			continue
		}
		goto end
	}
end:
	return
}
