// Package pathvars/route defines the Route type which represents a compiled HTTP route.
// Routes are created by the Router and contain the HTTP method, parsed template,
// and index information needed for request matching.
package pathvars

import (
	"fmt"
)

// Route represents a compiled HTTP endpoint with its method, template, and routing index.
// Routes are created during router compilation and used for efficient request matching.
type Route struct {
	// Method specifies the HTTP method for this route (GET, POST, etc.).
	// An empty string means the route matches any HTTP method.
	Method HTTPMethod

	// ParsedTemplate contains the parsed path template with parameters and regex for matching.
	ParsedTemplate *ParsedTemplate

	// Index indicates the position of this route in the router's route list.
	// This can be used to identify which specific route was matched.
	Index int

	Description string       // Human-readable description of the endpoint
	Cardinality Cardinality  // Expected number of result rows (one, many, etc.)
	RowType     DBRowType    // Format for returning results (json, columns, etc.)
	ColumnTypes []DBDataType // Expected data types for result columns
}

func (r Route) Endpoint() string {
	return fmt.Sprintf("%s %s", r.Method, r.ParsedTemplate)
}

type Cardinality string
type DBRowType string
type DBDataType string

// DBDataTypes converts from any slice of string to slice of type derived from
// string into a slice of []DBDataType. It does not parse, it just converts so it
// assumes validates data as input.
func DBDataTypes[S ~string](s []S) (dts []DBDataType) {
	dts = make([]DBDataType, len(s))
	for i, d := range s {
		dts[i] = DBDataType(d)
	}
	return dts
}
