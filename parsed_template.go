// Package pathvars/template provides path template parsing and matching functionality.
// Templates represent parsed URL patterns with parameters that can be matched against
// incoming HTTP requests to extract parameter values.
package pathvars

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mikeschinkel/go-pathvars/pvtypes"
)

// ParsedTemplate represents a parsed path template with parameters and compiled
// regex. Template strings like "/users/{id:int}/posts/{slug:string}" can be
// parsed into a ParsedTemplate to allow for matching incoming request paths to
// extract parameter values.
type ParsedTemplate struct {
	// raw stores the original template string for reference and error reporting.
	original string

	// segments contains the parsed path segments, both literal and parameter segments.
	segments []Segment

	// params maps parameter names to their definitions for validation and extraction.
	params *pvtypes.OrderedMap[Identifier, Parameter]

	parsedQuery *ParsedQuery

	// regex is the compiled regular expression used for efficient path matching.
	regex *regexp.Regexp
}

func (pt *ParsedTemplate) ParsedQuery() *ParsedQuery {
	return pt.parsedQuery
}

func (pt *ParsedTemplate) Original() string {
	// TODO: Verify if we should just return RAW, or if we should assemble from parsed parts.
	// If we do we can use Substitute() and pass in the template variables as if they were values.
	return pt.original
}

// Template returns the string representation of the template that was parsed but
// as a string-derived type Template.  See String() comments for more details.
func (pt *ParsedTemplate) Template() Template {
	// TODO: Verify if we should just return RAW, or if we should assemble from parsed parts.
	// If we do we can use Substitute() and pass in the template variables as if they were values.
	return Template(pt.original)
}

// String returns the string value of a parsed template which should just be what
// is original contains. However, there may be a difference between the original string and
// a recomposed string, so we need to be vigilant. This is currently (2025-10-05)
// only called implicitly from a fmt.Sprintf() from within apipkg.ParseEndpoint()
// which is the package pathvars was originally developed for. Note that I added
// Template() first then realized this was likely better as a fmt.Stringer method
// so added it too, but did not remove Template() even though I am not currently
// using it simply because it returns a Template type vs. a string type.
func (pt *ParsedTemplate) String() string {
	return pt.original
}

func (pt *ParsedTemplate) Normalize() {
	if len(pt.original) != 0 && pt.original[0] == '/' {
		pt.original = pt.original[1:]
	}
}

// Match attempts to match a path and query string against this template.
// Returns a ValuesMap containing extracted parameter values and a boolean indicating
// whether the match was successful. Both path parameters (from URL segments) and
// query parameters are extracted and validated according to their type constraints.
func (pt *ParsedTemplate) Match(path, query string) (MatchAttempt, error) {
	var errs []error
	var matchedPath, matchedQuery bool
	var err error

	valuesMap := pvtypes.NewValuesMap(0)

	// First, match path parameters using regex
	matchedPath, err = pt.matchPathParameters(path, &valuesMap)
	if err != nil {
		errs = append(errs, err)
	}
	matchedQuery, err = pt.matchQueryParameters(query, &valuesMap)
	if err != nil {
		errs = append(errs, err)
	}
	if valuesMap.Len() == 0 {
		valuesMap = pvtypes.NewValuesMap(0)
	}

	return MatchAttempt{
		PathMatched:  matchedPath,
		QueryMatched: matchedQuery,
		ValuesMap:    valuesMap,
	}, CombineErrs(errs)
}

// matchPathParameters matches path parameters using regex and adds them to vars.
// Returns false if the path doesn't match the template or if parameter validation fails.
func (pt *ParsedTemplate) matchPathParameters(path string, valuesMap *pvtypes.ValuesMap) (matched bool, err error) {
	var matches []string
	var n int
	var name Identifier
	var value string
	var param Parameter
	var exists bool
	var errs []error
	var validationErrors []paramValidationError
	var userProvidedParams pvtypes.ValuesMap

	if pt.regex == nil {
		// No path regex means no path parameters
		matched = true
		goto end
	}

	matches = pt.regex.FindStringSubmatch(path)
	if matches == nil {
		// No match is not an error, just no match
		matched = false
		goto end
	}
	matched = true

	// Extract parameters from regex groups
	n = 1 // Skip full match at index 0
	for _, segment := range pt.segments {
		if !segment.IsParameter() {
			continue
		}

		if n >= len(matches) {
			goto end
		}

		// We currently only support one parameter per segment
		name = segment.Parameters[0].Name
		value = matches[n]

		// Validate parameter type and constraints
		param, exists = pt.params.Get(name)
		if exists && param.Location() == PathLocation {
			err = param.Validate(value)
			if err != nil {
				// Collect error metadata - delay full error construction until
				// after loop completes so SuggestionURL sees complete valuesMap
				validationErrors = append(validationErrors, paramValidationError{
					param:    param,
					value:    value,
					validErr: err,
					location: PathLocation,
				})
			}
		}
		if valuesMap.IsNil() {
			*valuesMap = pvtypes.NewValuesMap(0)
		}
		(*valuesMap).Set(name, value)

		// Decompose multi-segment parameters into component values
		if param.MultiSegment {
			pt.decomposeValue(*valuesMap, name, value, param.DataType())
		}

		n++
	}

	// For path parameter errors, we need to include query params the user provided
	// Parse the query string to get user-provided query params (for ADR-018 compliance)
	userProvidedParams = pvtypes.NewValuesMap(valuesMap.Len() + 10) // Add capacity for query params
	// Add all path parameters (all are user-provided from URL path)
	for name, value := range valuesMap.Iterator() {
		userProvidedParams.Set(name, value)
	}
	// Add query parameters if available (we may not have parsed them yet)
	if pt.parsedQuery != nil {
		for paramName, values := range pt.parsedQuery.Iterator() {
			if len(values) > 0 {
				userProvidedParams.Set(Identifier(paramName), values[0])
			}
		}
	}

	// Now that valuesMap is complete, construct validation errors with proper suggestion URLs
	err = pt.buildValidationErrors(errs, validationErrors, path, &userProvidedParams)

	// buildValidationErrors replaced the following. Feel free to delete this when ready.
	//for _, ve := range validationErrors {
	//	exampleURL := pt.Example(&pvtypes.ExampleArgs{
	//		ProblematicParam:   ve.param,
	//		UserProvidedParams: &userProvidedParams,
	//		ValidationErr:      ve.validErr,
	//	})
	//	errs = append(errs, NewTemplateError(ve.validErr, TemplateErrorArgs{
	//		Endpoint:   pt.Original(),
	//		Example:    exampleURL,
	//		Source:     path,
	//		Location:   ve.location,
	//		Suggestion: ve.param.ErrorSuggestion(ve.validErr, ve.value, exampleURL),
	//		Parameter:  ve.param,
	//	}))
	//}
	//err = CombineErrs(errs)

end:
	return matched, err
}

// paramValidationError holds error metadata collected during parameter validation.
// We delay error construction until after all parameters are processed so that
// suggestion URLs can include the complete set of user-provided parameters.
type paramValidationError struct {
	param    Parameter
	value    string
	validErr error
	location LocationType
}

// matchQueryParameters matches query parameters and adds them to vars.
// Returns false if required parameters are missing or if validation fails.
// Optional parameters are handled gracefully with default values when provided.
func (pt *ParsedTemplate) matchQueryParameters(query string, valuesMap *pvtypes.ValuesMap) (matched bool, err error) {
	var p Parameter
	var value string
	var values []string
	var found bool
	var errs []error
	var addValue func(Identifier, any)
	var validationErrors []paramValidationError
	var userProvidedParams pvtypes.ValuesMap

	// ParseBytes query string
	if query != "" {
		pt.parsedQuery, err = ParseQuery(query)
		if err != nil {
			err = WithErr(err, ErrInvalidURLQueryString, "url_query", query)
			goto end
		}
		if pt.parsedQuery == nil {
			goto end
		}
	}
	if pt.parsedQuery == nil {
		pt.parsedQuery = NewParsedQuery(0)
	}

	matched = true

	addValue = func(name Identifier, value any) {
		if valuesMap.IsNil() {
			*valuesMap = pvtypes.NewValuesMap(0)
		}
		(*valuesMap).Set(name, value)
	}

	// Check each query parameter in the template
	for p = range pt.params.Values() {
		if p.Location() != QueryLocation {
			continue
		}

		// Check if parameter is present in query string
		values, found = pt.parsedQuery.Get(string(p.Name))
		switch {
		case found && len(values) > 0:
			// Use the first value if multiple are provided
			value = values[0]

			// Validate matched parameter
			err = p.Validate(value)
			if err != nil {
				// Collect error metadata - delay full error construction until
				// after loop completes so SuggestionURL sees complete valuesMap
				validationErrors = append(validationErrors, paramValidationError{
					param:    p,
					value:    value,
					validErr: err,
					location: QueryLocation,
				})
				// Still add to valuesMap even if invalid - needed for complete error suggestions
			}
			addValue(p.Name, value)

		case p.Optional:
			// Optional parameter not provided
			if p.DefaultValue != nil {
				// Use explicit default value
				addValue(p.Name, *p.DefaultValue)
				continue
			}
			// No explicit default - use type-specific implicit default
			classifier, err := GetDataTypeClassifier(p.DataType())
			if err != nil {
				errs = append(errs, err)
				continue
			}
			implicitDefault := classifier.DefaultValue()
			if implicitDefault == nil {
				// Type requires explicit default - this should be caught at parse time
				continue
			}
			addValue(p.Name, *implicitDefault)

		default:
			// If no default value, simply omit from valuesMap (empty string behavior)
			matched = false
			// Required parameter not provided
			errs = append(errs, NewErr(ErrRequiredParameterNotProvided,
				"parameter_name", p.Name,
				"data_type", p.DataTypeSlug(),
				"fault_source", ClientFaultSource.Slug(),
				"parameter_location", p.Location(),
			))
			continue
		}
	}

	// Build a map of ONLY user-provided parameters for ADR-018 compliance
	// This excludes optional parameters that got default values but weren't in the HTTP request
	userProvidedParams = pvtypes.NewValuesMap(pt.parsedQuery.Len())
	for paramName, values := range pt.parsedQuery.Iterator() {
		if len(values) > 0 {
			userProvidedParams.Set(Identifier(paramName), values[0])
		}
	}

	// Now that valuesMap is complete, construct validation errors with proper suggestion URLs
	err = pt.buildValidationErrors(errs, validationErrors, query, &userProvidedParams)

	// buildValidationErrors replaced the following. Feel free to delete this when ready.
	//for _, ve := range validationErrors {
	//	exampleURL := pt.Example(&pvtypes.ExampleArgs{
	//		ProblematicParam:   ve.param,
	//		UserProvidedParams: &userProvidedParams,
	//		ValidationErr:      ve.validErr,
	//	})
	//	errs = append(errs, NewTemplateError(ve.validErr, TemplateErrorArgs{
	//		Endpoint:   pt.Original(),
	//		Example:    exampleURL,
	//		Source:     query,
	//		Location:   ve.location,
	//		Suggestion: ve.param.ErrorSuggestion(ve.validErr, ve.value, exampleURL),
	//		Parameter:  ve.param,
	//	}))
	//}
	//err = CombineErrs(errs)
end:
	return matched, err
}

func (pt *ParsedTemplate) buildValidationErrors(errs []error, paramErrs []paramValidationError, source string, userProvidedParams *pvtypes.ValuesMap) error {
	// Now that valuesMap is complete, construct validation errors with proper suggestion URLs
	for _, ve := range paramErrs {
		exampleURL := pt.Example(&pvtypes.ExampleArgs{
			ProblematicParam:   ve.param,
			UserProvidedParams: userProvidedParams,
			ValidationErr:      ve.validErr,
		})
		errs = append(errs, NewTemplateError(ve.validErr, TemplateErrorArgs{
			Endpoint:   pt.Original(),
			Example:    exampleURL,
			Source:     source,
			Location:   ve.location,
			Suggestion: ve.param.ErrorSuggestion(ve.validErr, ve.value, exampleURL),
			Parameter:  ve.param,
		}))
	}
	return CombineErrs(errs)
}

// decomposeValue decomposes a multi-segment value into its component parts and adds them
// to the values map with suffixed keys. For dates, creates param_year, param_month, param_day.
// For other types, creates param_1, param_2, param_3, etc.
func (pt *ParsedTemplate) decomposeValue(valuesMap pvtypes.ValuesMap, name Identifier, value string, dataType PVDataType) {
	// Split by the appropriate separator
	var parts []string
	var separator string

	switch dataType {
	case DateType:
		// Date type uses slash separator
		separator = "/"
		parts = strings.Split(value, separator)

		// Add decomposed date components with semantic names
		if len(parts) >= 1 && parts[0] != "" {
			valuesMap.Set(Identifier(string(name)+"_year"), parts[0])
		}
		if len(parts) >= 2 && parts[1] != "" {
			valuesMap.Set(Identifier(string(name)+"_month"), parts[1])
		}
		if len(parts) >= 3 && parts[2] != "" {
			valuesMap.Set(Identifier(string(name)+"_day"), parts[2])
		}

	default:
		// For all other types, use slash separator and numeric suffixes
		separator = "/"
		parts = strings.Split(value, separator)

		// Add decomposed components with numeric suffixes
		for i, part := range parts {
			if part == "" {
				continue
			}
			key := Identifier(fmt.Sprintf("%s_%d", name, i+1))
			valuesMap.Set(key, part)
		}
	}
}

// Parameters returns the Ordered Map of parameters
func (pt *ParsedTemplate) Parameters() *pvtypes.OrderedMap[Identifier, Parameter] {
	return pt.params
}

// Validate checks parameter values against the template requirements.
// It validates each provided parameter value against its type and constraints,
// and ensures all required parameters are present.
func (pt *ParsedTemplate) Validate(params map[Identifier]any) (err error) {
	var errs []error
	var value any
	var found bool
	var valueStr string

	// Iterate through all template parameters
	for p := range pt.params.Values() {
		// Check if parameter value provided
		value, found = params[p.Name]

		// Handle missing parameters
		if !found {
			if !p.Optional {
				// Required parameter missing
				errs = append(errs, NewErr(
					ErrRequiredParameterNotProvided,
					"parameter", p.Name,
				))
			}
			// Optional parameters are OK to be missing
			continue
		}

		// Convert value to string for validation
		valueStr = fmt.Sprintf("%v", value)

		// Validate using existing Parameter.Validate method
		err = p.Validate(valueStr)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return CombineErrs(errs)
}

// Substitute builds a path from parameter values by replacing template placeholders.
// TODO: Implementation needed - should build path by substituting values.
func (pt *ParsedTemplate) Substitute(values *pvtypes.OrderedMap[Identifier, any]) (result string, err error) {
	var errs []error
	var query string

	sbp := strings.Builder{}
	n := 0
	for _, seg := range pt.segments {
		sbp.WriteByte('/')
		if seg.IsLiteral() {
			sbp.WriteString(seg.Raw)
			continue
		}
		if seg.Prefix != "" {
			sbp.WriteString(seg.Prefix)
		}
		// We currently only support one parameter per segment
		paramName := seg.Parameters[0].Name
		value, ok := values.Get(paramName)
		if !ok {
			errs = append(errs, NewErr(
				ErrParameterNotFoundInValuesMap,
				ErrPathParameterNotFoundInValuesMap,
				"parameter_name", paramName,
				"values_map", values,
			))
			continue
		}
		sbp.WriteString(fmt.Sprintf("%v", value))
		if seg.Suffix != "" {
			sbp.WriteString(seg.Suffix)
		}
		n++
	}
	sbq := strings.Builder{}
	for name, value := range values.Iterator() {
		p, ok := pt.params.Get(name)
		if !ok {
			errs = append(errs, NewErr(
				ErrParameterNotFoundInValuesMap,
				ErrQueryParameterNotFoundInValuesMap,
				"parameter_name", p.Name,
				"values_map", values,
			))
			continue
		}
		if p.Location() != QueryLocation {
			continue
		}
		sbq.WriteString(fmt.Sprintf("%s=%v&", p.Name, value))
	}
	if len(errs) > 0 {
		err = CombineErrs(errs)
		goto end
	}
	result = sbp.String()
	query = sbq.String()
	if len(query) != 0 {
		result += "?" + query[:len(query)-1]
	}
end:
	return result, err
}

// Example generates an example URL for this template.
// When called with empty args, generates a simple example with all required parameters.
// When called with error context (ProblematicParam, UserProvidedParams, ValidationErr),
// generates a context-aware example following ADR-018 guidelines:
// - Only includes required parameters OR parameters user actually provided
// - Correct parameters shown as {PLACEHOLDER}
// - Problematic parameter shown with Example() value (using constraint's example if available)
// - Problematic parameter positioned LAST in query string
// - Query parameters appear in the order user provided them (preserving request structure)
func (pt *ParsedTemplate) Example(args ...*pvtypes.ExampleArgs) (result string) {
	// Handle simple case - no args provided
	if len(args) == 0 || (args[0].ProblematicParam.Name == "" && args[0].UserProvidedParams == nil) {
		params := pvtypes.NewOrderedMap[Identifier, any](pt.params.Len())
		for param := range pt.params.Values() {
			if param.Optional {
				continue
			}
			params.Set(param.Name, param.Example(nil, nil))
		}
		result, _ = pt.Substitute(params)
		return result
	}

	// Handle error-context case with full ADR-018 logic
	arg := args[0]
	// Build separate maps for path and query parameters
	pathParams := pvtypes.NewOrderedMap[Identifier, any](pt.params.Len())
	correctQueryParams := pvtypes.NewOrderedMap[Identifier, any](pt.params.Len())
	problematicQueryParams := pvtypes.NewOrderedMap[Identifier, any](1)

	// Track which parameters we've already added
	addedParams := make(map[Identifier]bool, pt.params.Len())

	if arg.UserProvidedParams == nil {
		vm := pvtypes.NewValuesMap(0)
		arg.UserProvidedParams = &vm
	}

	// First pass: Add user-provided parameters in request order
	for name := range arg.UserProvidedParams.Keys() {
		param, ok := pt.params.Get(name)
		if !ok {
			continue // Skip parameters not in template
		}

		isProblematic := param.Name == arg.ProblematicParam.Name

		// Determine the value to show
		var value any
		if isProblematic {
			// Problematic parameter: use Example() value with validation error context
			value = param.Example(arg.ValidationErr, nil)
		} else {
			// Correct parameter: use {PLACEHOLDER} format
			value = fmt.Sprintf("{%s}", strings.ToUpper(string(param.Name)))
		}

		// Add to appropriate map based on location and whether it's problematic
		if param.Location() == PathLocation {
			// Path parameters must stay in position, can't move to end
			pathParams.Set(param.Name, value)
		} else if isProblematic {
			// Problematic query parameter goes in separate map (for positioning last)
			problematicQueryParams.Set(param.Name, value)
		} else {
			// Correct query parameters go first (in request order)
			correctQueryParams.Set(param.Name, value)
		}

		addedParams[param.Name] = true
	}

	// Second pass: Add required parameters that weren't user-provided (in API definition order)
	for param := range pt.params.Values() {

		// Determine if this parameter should be included
		isRequired := !param.Optional
		value, ok := arg.UserProvidedParams.Get(param.Name)
		isUserProvided := ok && value != nil
		isProblematic := param.Name == arg.ProblematicParam.Name

		// Skip optional parameters user didn't provide (unless it's the problematic one)
		if !isRequired && !isUserProvided && !isProblematic {
			continue
		}

		// Determine the value to show
		if isProblematic {
			// Problematic parameter: use Example() value with validation error context
			value = param.Example(arg.ValidationErr, nil)
		} else {
			// Correct parameter: use {PLACEHOLDER} format
			value = fmt.Sprintf("{%s}", strings.ToUpper(string(param.Name)))
		}

		// Add to appropriate map based on location and whether it's problematic
		if param.Location() == PathLocation {
			// Path parameters must stay in position, can't move to end
			pathParams.Set(param.Name, value)
		} else if isProblematic {
			// Problematic query parameter goes in separate map (for positioning last)
			problematicQueryParams.Set(param.Name, value)
		} else {
			// Correct query parameters go first
			correctQueryParams.Set(param.Name, value)
		}
	}

	// Build the URL by combining path params, correct query params, then problematic query params
	result = pt.buildExampleURL(pathParams, correctQueryParams, problematicQueryParams)
	return result
}

// buildExampleURL constructs the final URL with path params and query params in correct order
func (pt *ParsedTemplate) buildExampleURL(pathParams, correctQueryParams, problematicQueryParams *pvtypes.OrderedMap[Identifier, any]) string {
	var errs []error

	// Build path portion
	sbp := strings.Builder{}
	for _, seg := range pt.segments {
		sbp.WriteByte('/')
		if seg.IsLiteral() {
			sbp.WriteString(seg.Raw)
			continue
		}
		if seg.Prefix != "" {
			sbp.WriteString(seg.Prefix)
		}
		// We currently only support one parameter per segment
		paramName := seg.Parameters[0].Name
		value, ok := pathParams.Get(paramName)
		if !ok {
			// This shouldn't happen if logic is correct
			errs = append(errs, NewErr(
				ErrParameterNotFoundInValuesMap,
				ErrPathParameterNotFoundInValuesMap,
				"parameter_name", paramName,
			))
			continue
		}
		sbp.WriteString(fmt.Sprintf("%v", value))
		if seg.Suffix != "" {
			sbp.WriteString(seg.Suffix)
		}
	}

	// Build query string: correct params first, then problematic params last
	sbq := strings.Builder{}

	// Add correct query parameters
	for name, value := range correctQueryParams.Iterator() {
		sbq.WriteString(fmt.Sprintf("%s=%v&", name, value))
	}

	// Add problematic query parameters (last)
	for name, value := range problematicQueryParams.Iterator() {
		sbq.WriteString(fmt.Sprintf("%s=%v&", name, value))
	}

	if len(errs) > 0 {
		// In case of errors, fall back to simple example (no args)
		return pt.Example()
	}

	result := sbp.String()
	query := sbq.String()
	if len(query) > 0 {
		// Remove trailing '&'
		result += "?" + query[:len(query)-1]
	}

	return result
}
