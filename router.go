// Package pathvars provides path variable routing and parameter extraction functionality
// for HTTP requests. It supports typed path parameters with validation constraints,
// query parameter handling, and regex-based route matching.
//
// The package enables parsing URL templates like "/users/{id:int:range[1..1000]}"
// and matching them against incoming HTTP requests, extracting and validating
// parameter values according to specified data types and constraints.
//
// Key features:
//   - Typed path parameters (string, int, uuid, date, etc.)
//   - Parameter validation constraints (range, length, regex, enum, etc.)
//   - Multi-segment parameters for capturing multiple path segments
//   - Optional parameters with default values
//   - Query parameter extraction and validation
//   - Efficient regex-based route matching
//
// Example usage:
//
//	router := pathvars.NewRouter()
//	params := []pathvars.Parameter{
//		// Parameter definitions go here
//	}
//	err := router.AddRoute("GET" "/users/{id:int}", params)
//	if err != nil {
//		// handle error
//	}
//	err = router.Compile()
//	if err != nil {
//		// handle error
//	}
//
//	// Later, during request handling:
//	result, err := router.Match(request)
//	if err == nil {
//		userId, found := result.GetValue("id")
//		// use the extracted parameter
//	}
package pathvars

import (
	"net/http"
)

// PathSpec represents a path specification string like "GET /users/{id}" or "/users/{id}".
type PathSpec string

// Method represents an HTTP method string like "GET", "POST", etc.
type Method string

// Path represents a URL path string like "/users/{id}".
type Path string

// Router holds routes and provides request matching functionality.
// Routes are compiled as they are added via AddRoute().
type Router struct {
	routes    []*Route
	maxParams int
}

// NewRouter creates a new router instance.
func NewRouter() *Router {
	return &Router{
		routes: make([]*Route, 0),
	}
}

type RouteArgs struct {
	Parameters  []Parameter
	Index       int
	Description string       // Human-readable description of the endpoint
	Cardinality Cardinality  // Expected number of result rows (one, many, etc.)
	RowType     DBRowType    // Format for returning results (json, columns, etc.)
	ColumnTypes []DBDataType // Expected data types for result columns
}

// AddRoute adds a route to the router with the specified path specification and parameters.
// The pathSpec can be in format "METHOD /path" (e.g., "GET /users/{id}") or just "/path"
// for any method. Parameters define the expected path and query parameters for this route.
func (r *Router) AddRoute(method HTTPMethod, path Template, args *RouteArgs) (err error) {
	var pt *ParsedTemplate
	var route *Route
	var paramCount int

	if args == nil {
		args = &RouteArgs{}
	}

	if path == "" {
		// Trim leading slash ('/') on sub path
		path = "/"
	}
	if path[0] != '/' {
		// Trim leading slash ('/') on sub path
		path = "/" + path
	}

	pt, err = ParseTemplate(string(path))
	if err != nil {
		err = WithErr(
			err,
			"path_spec", path,
			"method", method,
			"path", path,
		)
		goto end
	}

	if pt == nil {
		// This if statement if only here because without it Goland is reporting that
		// `pt` might be nil in the expressions below even though I traced through
		// the logic and it cannot be nil if err==nil.
		goto end
	}

	// Merge provided parameters into the parsed template's params map
	// This ensures query parameters (not in the path string) are available for validation
	// Only add parameters that don't already exist to avoid overwriting path parameters
	if len(args.Parameters) != 0 {
		for _, param := range args.Parameters {
			// Only add if not already present (don't overwrite path parameters)
			_, exists := pt.params.Get(param.Name)
			if exists {
				continue
			}
			pt.params.Set(param.Name, param)
		}
	}

	paramCount = pt.params.Len()
	if paramCount != 0 {
		// Track max params for optimization
		if paramCount > r.maxParams {
			r.maxParams = paramCount
		}
	}

	if args.Index == 0 {
		args.Index = len(r.routes)
	}

	route = &Route{
		Method:         method,
		ParsedTemplate: pt,
		Index:          args.Index,
		Description:    args.Description,
		Cardinality:    args.Cardinality,
		RowType:        args.RowType,
		ColumnTypes:    args.ColumnTypes,
	}

	r.routes = append(r.routes, route)

end:
	return err
}

// Match matches an HTTP request against the routes and returns
// the first matching route along with extracted parameter values.
// Routes match in the order they were added, giving users control
// over matching priority.
// Returns ErrNoMatch if no route matches the request.
func (r *Router) Match(req *http.Request) (result MatchResult, err error) {

	u := req.URL

	for _, route := range r.routes {
		// Check method match (empty method means any)
		if route.Method != "" && route.Method != HTTPMethod(req.Method) {
			continue
		}

		var attempt MatchAttempt
		attempt, err = route.ParsedTemplate.Match(u.Path, u.RawQuery)

		// If path didn't match, try next route (ignore any errors)
		//goland:noinspection GoDfaErrorMayBeNotNil
		if attempt.ShouldContinue() {
			continue
		}

		// Path matched - if there's an error, it's a validation failure
		if err != nil {
			goto end
		}

		// Path matched and validation passed - success
		result = MatchResult{
			Index:     route.Index,
			Route:     route,
			valuesMap: attempt.ValuesMap,
		}
		goto end
	}

	err = NewErr(
		ErrNoRouteMatched,
		"fault_source", ClientFaultSource.Slug(),
	)

end:
	if err != nil {
		err = WithErr(err,
			ErrNoMatch,
			"route_count", len(r.routes),
			"method", req.Method,
			"path", u.Path,
			"query_string", u.RawQuery,
		)
	}
	return result, err
}
